package main

import "github.com/confluentinc/confluent-kafka-go/kafka"

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

// placeOrder sends a binary payload, the event (an error or a confirmation that
// it went ok) will then be recieved on op.producer event chan.
func (op *orderPlacer) placeOrder(payload []byte) error {
	err := op.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &op.topic,
			Partition: kafka.PartitionAny,
		},
		Value: payload,
	}, nil,
	)
	if err != nil {
		return err
	}

	return nil
}
