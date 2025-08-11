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
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
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
	dbStore := store.NewDBStore(db)
	orderService := service.New(dbStore)
	server := api.NewServer(orderService)

	// Create Kafka consumer
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "foo",
		"auto.offset.reset": "smallest",
	}
	kafkaConsumer, err := consumer.NewKafkaConsumer(kafkaConfig, "orders", orderService)
	if err != nil {
		log.Fatalf("Failed to create kafka consumer: %v", err)
	}

	// Start server and consumer
	ctx, cancel := context.WithCancel(context.Background())
	go server.Start(cfg.HTTPServerPort)
	go kafkaConsumer.Run(ctx)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
