package main

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// orderPlacer is a wrapper for a kafka.Producer.
type orderPlacer struct {
	producer *kafka.Producer
	topic    string
}

// newOrderPlacer returns a new instance of OrderPlacer.
func newOrderPlacer(p *kafka.Producer, topic string) *orderPlacer {
	return &orderPlacer{
		producer: p,
		topic:    topic,
	}
}

// placeOrder sends a binary payload. The delivery event will be sent to the
// producer's event channel.
func (op *orderPlacer) placeOrder(payload []byte) error {
	now := time.Now().UnixMilli()
	timestamp := fmt.Sprintf("%d", now) // to calculate kafka consumer lag
	return op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &op.topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers: []kafka.Header{
			{Key: "creation_timestamp_ms", Value: []byte(timestamp)},
		},
	}, nil)
}
