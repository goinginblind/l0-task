package metrics

import (
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/prometheus/client_golang/prometheus"
)

// ObserveKafkaMessageLatency extracts message header with "creation_timestamp_ms" value.
// This funcion should calculates message processing latency - time it takes from the moment kafka producer
// sends the message to consumer successfully processing it. This function should be deffered (!).
func ObserveKafkaMessageLatency(msg *kafka.Message, observer prometheus.Observer) {
	var msgCreationTime time.Time
	for _, header := range msg.Headers {
		if header.Key == "creation_timestamp_ms" {
			ms, err := strconv.ParseInt(string(header.Value), 10, 64)
			if err == nil {
				msgCreationTime = time.UnixMilli(ms)
			}
			break
		}
	}

	if !msgCreationTime.IsZero() {
		latency := time.Since(msgCreationTime).Seconds()
		observer.Observe(latency)
	}
}
