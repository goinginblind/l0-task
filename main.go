package main

import (
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type OrderPlacer struct {
	producer   *kafka.Producer
	topic      string
	deliverych chan kafka.Event
}

func NewOrderPlacer(p *kafka.Producer, topic string) *OrderPlacer {
	return &OrderPlacer{
		producer:   p,
		topic:      topic,
		deliverych: make(chan kafka.Event, 10000),
	}
}

func (op *OrderPlacer) placeOrder(orderType string, size int) error {
	var (
		format  = fmt.Sprintf("%s - %d", orderType, size)
		payload = []byte(format)
	)

	err := op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	},
		op.deliverych,
	)
	if err != nil {
		log.Fatal(err)
	}
	<-op.deliverych
	return nil
}

func main() {
	var (
		topic = "HVSE"
	)

	go func() {
		consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers": "localhost:9092",
			"group.id":          "foo",
			"auto.offset.reset": "smallest",
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
				// application-specific processing
				fmt.Printf("Consumed a msg from the queue: %s\n", string(e.Value))
			case kafka.Error:
				log.Fatalf("Error: %v\n", e.Error())
			}
		}
	}()

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "foo",
		"acks":              "all",
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %s\n", err)
	}

	op := NewOrderPlacer(p, topic)
	for i := range 1000 {
		err = op.placeOrder("New Order", i)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 5)
	}
}
