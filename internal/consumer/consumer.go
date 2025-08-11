package consumer

import (
	"context"
	"encoding/json"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/service"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer consumes messages from Kafka and processes them.
type KafkaConsumer struct {
	consumer *kafka.Consumer
	service  service.OrderService
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(cfg *kafka.ConfigMap, topic string, service service.OrderService) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer: c,
		service:  service,
	}, nil
}

// Run starts the consumer loop.
func (kc *KafkaConsumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer...")
			kc.consumer.Close()
			return
		default:
			ev := kc.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				var order domain.Order
				if err := json.Unmarshal(e.Value, &order); err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					continue
				}
				if err := kc.service.ProcessNewOrder(ctx, &order); err != nil {
					log.Printf("Failed to process order: %v", err)
				}
			case kafka.Error:
				log.Printf("Kafka error: %v", e)
			}
		}
	}
}
