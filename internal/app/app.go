package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/goinginblind/l0-task/internal/api"
	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/consumer"
	"github.com/goinginblind/l0-task/internal/pkg/health"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
)

// App struct holds all the core components of the application
type App struct {
	cfg      *config.Config
	logger   logger.Logger
	db       *sql.DB
	server   *api.Server
	consumer *consumer.KafkaConsumer
	hc       *health.DBHealthChecker
}

// New returns a new App instance
func New() (*App, error) {
	// Create a appLogger and defer its buffer flush
	appLogger, err := logger.NewSugarLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to create a logger: %w", err)
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	db.SetMaxOpenConns(cfg.Database.MaxIdlingConnections)
	db.SetMaxIdleConns(cfg.Database.MaxConnections) // TODO III: maybe add idling timeout

	// Create store, service, and server
	dbStore := store.NewDBStore(db, appLogger)
	orderService := service.New(dbStore, appLogger)
	server := api.NewServer(orderService, appLogger, cfg.HTTPServer)

	hc := health.NewDBHealthChecker(db, appLogger, cfg.Health)
	kafkaConsumer, err := consumer.NewKafkaConsumer(cfg.Kafka, cfg.Consumer, orderService, appLogger, hc)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return &App{
		cfg:      cfg,
		logger:   appLogger,
		db:       db,
		server:   server,
		consumer: kafkaConsumer,
		hc:       hc,
	}, nil
}

// Run runs the whole logic, the 'command center'
func (a *App) Run() {
	defer func() {
		if err := a.logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v\n", err)
		}
		a.db.Close()
	}()

	var wg sync.WaitGroup

	// The HTTP server starts
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Infow("Starting server on:" + a.cfg.HTTPServer.Port)
		if err := a.server.Start(":" + a.cfg.HTTPServer.Port); err != nil {
			a.logger.Fatalw("Failed to start the HTTP server", "error", err)
		}
	}()

	// the consumer starts
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.logger.Infow("Starting kafka consumer")
		a.consumer.Run(ctx)
		a.logger.Infow("Stopping kafka consumer")
	}()
	go a.hc.Start(ctx)

	// block til signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	a.logger.Infow("Shutting down...")

	// proper server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.HTTPServer.ShutdownTimeout)
	defer shutdownCancel()

	cancel() // exit the consumer loop
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.logger.Errorw("HTTP server shutdown error: %v", err)
	}

	wg.Wait()
	a.logger.Infow("Shutdown complete.")
}
