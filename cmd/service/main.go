package main

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	// TODO: don't leave hardcoded stuff here
	var (
		topic            = "orders"
		kafkaServiceAddr = "localhost:9092"
		kafkaClientID    = "foo"
		offsetReset      = "smallest"
	)

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServiceAddr,
		"group.id":          kafkaClientID,
		"auto.offset.reset": offsetReset,
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %s\n", err)
	}

	err = consumer.Subscribe(topic, nil)
	if err != nil {
		log.Fatalf("Failed to subscribe consumer to topics: %s\n", err)
	}

	for {
		ev := consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			// TODO I: processing
			// recieve a binary message
			// process it, unmarshal into a struct
			// validate the fields, return a struct and an error
			// TODO II: store the struct in a PSQL database
			// atomic transaction, all-or-nothing
			// return an error in case of not succeeding
			log.Printf("Consumed a message: %s\n", e.Timestamp)
		case kafka.Error:
			log.Fatalf("Error: %v\n", e.Error())
		}
	}
}
