package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/goinginblind/l0-task/internal/api"
	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/consumer"
	"github.com/goinginblind/l0-task/internal/pkg/health"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
	"github.com/joho/godotenv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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

	// Try to load .env
	err = godotenv.Load(".env")
	if err != nil {
		log.Printf("fail to parse .env: %v\n", err)
		log.Println("looking for the enviromental variables in the enviroment...")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := sql.Open("pgx", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create store, service, and server
	dbStore := store.NewDBStore(db, logger)
	orderService := service.New(dbStore, logger)
	server := api.NewServer(orderService, logger)

	// Create Kafka consumer
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":     "localhost:9092",
		"group.id":              "orders-consumer",
		"auto.offset.reset":     "earliest",
		"enable.auto.commit":    false,
		"isolation.level":       "read_committed",
		"max.poll.interval.ms":  300000, // 5 min
		"fetch.min.bytes":       1,
		"fetch.max.bytes":       1048576, // 1Mb
		"session.timeout.ms":    10000,   // 10 sec
		"heartbeat.interval.ms": 3000,    //3 sec
	}
	ctx, cancel := context.WithCancel(context.Background())
	// TODO II: replace these hardcoded time intervals
	hc := health.NewDBHealthChecker(db, logger, time.Second*5, time.Second*180)
	kafkaConsumer, err := consumer.NewKafkaConsumer(kafkaConfig, "orders", orderService, logger, hc) // <- topic is hardcoded
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}

	// Start server and consumer
	go hc.Start(ctx)
	go server.Start(cfg.HTTPServerPort)
	go kafkaConsumer.Run(ctx)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
