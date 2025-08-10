package main

import (
	"log"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	var (
		topic            = getEnv("KAFKA_TOPIC", "orders")
		kafkaServiceAddr = getEnv("KAFKA_BROKERS", "localhost:9092")
		kafkaClientID    = getEnv("KAFKA_CLIENT_ID", "foo")
		jsonFile         = getEnv("JSON_FILE", "valid_mock_orders.json")
	)

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": kafkaServiceAddr,
		"client.id":         kafkaClientID,
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}
	defer p.Close()

	// handle kafka events
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Successfully produced message to topic %s partition %d\n", *ev.TopicPartition.Topic, ev.TopicPartition.Partition)
				}
			}
		}
	}()

	op := newOrderPlacer(p, topic)
	datach, errs := streamJSONObjects(jsonFile)

	var wg sync.WaitGroup
	wg.Add(1)

	// handle errors
	go func() {
		err := <-errs
		if err != nil {
			log.Fatalf("JSON stream exited with error: %s\n", err)
		}
	}()

	// place orders using json obj's from the stream
	go func() {
		defer wg.Done()
		for data := range datach {
			op.placeOrder(data)
		}
	}()

	wg.Wait()

	leftUnsent := p.Flush(5 * 100)
	if leftUnsent > 0 {
		log.Printf("%d messages not delivered", leftUnsent)
	}
}
