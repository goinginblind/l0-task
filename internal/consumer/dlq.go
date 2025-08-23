package consumer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
)

// drainDLQReports logs if the message sent to DLQ was or was not delivered.
// It could be a place used for metrics and mostly provides observability.
func drainDLQReports(ctx context.Context, p *kafka.Producer, log logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-p.Events():
			if !ok {
				return
			}
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Errorw("DLQ delivery failed",
						"error", ev.TopicPartition.Error,
						"order_uid", string(ev.Key))
				} else {
					log.Infow("DLQ message delivered",
						"topic", *ev.TopicPartition.Topic,
						"partition", ev.TopicPartition.Partition,
						"offset", ev.TopicPartition.Offset,
						"order_uid", string(ev.Key))
				}
			}
		}
	}
}

// sendToDLQ sends the message to the dead-line queue topic
// specified in the config (default 'orders-dlq').
func (w *worker) sendToDLQ(msg *kafka.Message, reason error) {
	dlqHeaders := msg.Headers
	if reason != nil {
		dlqHeaders = append(dlqHeaders, kafka.Header{
			Key:   "DLQ REASON",
			Value: []byte(reason.Error()),
		})
	}

	err := w.deps.dlqPublisher.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &w.deps.dlqTopic, Partition: kafka.PartitionAny},
		Key:            msg.Key,
		Value:          msg.Value,
		Headers:        dlqHeaders,
	}, nil)

	if err != nil {
		w.deps.logger.Errorw("Failed to produce message to DLQ", "error", err, "order_uid", string(msg.Key))
	}
}
