package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	var (
		jsonFile      = flag.String("file", "mock.json", "Path to the JSON file with order data.")
		rps           = flag.Int("rps", 1, "Requests per second (rate of sending messages).")
		invalidJSON   = flag.Int("invalid-json", 0, "Number of syntactically incorrect JSON messages to send.")
		gibberish     = flag.Int("gibberish", 0, "Number of non-JSON, gibberish messages to send.")
		kafkaTopic    = flag.String("topic", getEnv("KAFKA_TOPIC", "orders"), "Kafka topic to produce to.")
		kafkaBrokers  = flag.String("brokers", getEnv("KAFKA_BROKERS", "localhost:9092"), "Kafka bootstrap servers.")
		kafkaClientID = flag.String("client-id", getEnv("KAFKA_CLIENT_ID", "orders-producer"), "Kafka client ID.")
	)
	flag.Parse()

	// Kafka Producer Setup
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  *kafkaBrokers,
		"client.id":          *kafkaClientID,
		"acks":               "all",
		"enable.idempotence": true,

		"linger.ms":        5,
		"batch.size":       65536, // 64kb
		"compression.type": "lz4",

		"retries":          5,
		"retry.backoff.ms": 100,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	defer p.Close()

	// deliveryWg waits for all messages to be delivered
	var deliveryWg sync.WaitGroup

	// Handle Kafka delivery reports, no fancy but it does the job fine
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Successfully produced message to topic %s, partition %d\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition)
				}
				deliveryWg.Done()
			}
		}
	}()

	op := newOrderPlacer(p, *kafkaTopic)
	ticker := time.NewTicker(time.Second / time.Duration(*rps))
	defer ticker.Stop()

	// handle system interruptions
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, flushing pending messages...")
		deliveryWg.Wait() // wait for acked messages
		left := p.Flush(15 * 1000)
		if left > 0 {
			log.Printf("%d messages were not delivered before shutdown.", left)
		}
		os.Exit(0)
	}()

	// send Gibberish & Invalid JSON, Gibberish == malformed data
	for i := 0; i < *gibberish; i++ {
		<-ticker.C
		payload := make([]byte, 32)
		rand.Read(payload)
		deliveryWg.Add(1)
		if err := op.placeOrder(payload); err != nil {
			log.Printf("Failed to produce gibberish message: %v\n", err)
			deliveryWg.Done()
		}
	}

	for i := 0; i < *invalidJSON; i++ {
		<-ticker.C
		payload := fmt.Appendf(nil, `{"order_uid": "invalid-%d", "items": [}`, i)
		deliveryWg.Add(1)
		if err := op.placeOrder(payload); err != nil {
			log.Printf("Failed to produce invalid JSON message: %v\n", err)
			deliveryWg.Done()
		}
	}

	// Send Orders from File
	datach, errs := streamJSONObjects(*jsonFile)
	var fileReadingWg sync.WaitGroup
	fileReadingWg.Add(1)

	// Handle file reading errors
	go func() {
		if err := <-errs; err != nil {
			log.Printf("JSON stream exited with error: %s. Continuing without file data.\n", err)
		}
	}()

	// Place orders using JSON objects from the stream
	go func() {
		defer fileReadingWg.Done()
		for data := range datach {
			<-ticker.C
			deliveryWg.Add(1)
			if err := op.placeOrder(data); err != nil {
				log.Printf("Failed to produce message from file: %v\n", err)
				deliveryWg.Done()
			}
		}
	}()

	fileReadingWg.Wait()
	deliveryWg.Wait()
	leftUnsent := p.Flush(15 * 1000)
	if leftUnsent > 0 {
		log.Printf("%d messages were not delivered.", leftUnsent)
	}
}
