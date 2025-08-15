package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/goinginblind/l0-task/internal/api"
	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/consumer"
	"github.com/goinginblind/l0-task/internal/pkg/health"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Create a logger and defer its buffer flush
	logger, err := logger.NewSugarLogger()
	if err != nil {
		log.Fatalf("Failed to create a logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("failed to sync logger: %v\n", err)
		}
	}()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create store, service, and server
	dbStore := store.NewDBStore(db, logger)
	orderService := service.New(dbStore, logger)
	server := api.NewServer(orderService, logger)

	// Create Kafka consumer
	hc := health.NewDBHealthChecker(db, logger, cfg.Health)
	kafkaConsumer, err := consumer.NewKafkaConsumer(cfg.Kafka, cfg.Consumer, orderService, logger, hc)
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Start server and consumer
	go hc.Start(ctx)
	go server.Start(":" + cfg.HTTPServer.Port)
	go kafkaConsumer.Run(ctx)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
