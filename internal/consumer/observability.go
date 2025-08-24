package consumer

import (
	"context"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/pkg/metrics"
)

// drainDLQReports logs if the message sent to DLQ was or was not delivered.
// It could be a place used for metrics and mostly provides
// observability (with logs but i will add metrics soon too)
func drainDLQReports(ctx context.Context, p DLQManager, log logger.Logger) {
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

// monitorConsumerLag calculates consumer lag (the backpressure from the producer)
func (kc *KafkaConsumer) monitorConsumerLag(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			assignedPartitions, err := kc.consumer.Assignment()
			if err != nil {
				kc.logger.Errorw("Failed to get assigned partitions for lag monitoring", "error", err)
				continue
			}

			if len(assignedPartitions) == 0 {
				continue
			}

			committedPartitions, err := kc.consumer.Committed(assignedPartitions, 5000)
			if err != nil {
				kc.logger.Errorw("Failed to get committed offsets for lag monitoring", "error", err)
				continue
			}

			for _, p := range committedPartitions {
				low, high, err := kc.consumer.QueryWatermarkOffsets(*p.Topic, p.Partition, 5000)
				if err != nil {
					kc.logger.Errorw("Failed to query watermark offsets", "error", err, "topic", *p.Topic, "partition", p.Partition)
					continue
				}

				var lag int64
				if p.Offset < 0 {
					lag = high - low
				} else {
					lag = high - int64(p.Offset)
				}

				if lag < 0 {
					lag = 0
				}

				metrics.ConsumerLag.WithLabelValues(*p.Topic, strconv.Itoa(int(p.Partition))).Set(float64(lag))
			}
		}
	}
}
