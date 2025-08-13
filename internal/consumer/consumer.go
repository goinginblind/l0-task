package consumer

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer consumes messages from Kafka and processes them.
type KafkaConsumer struct {
	consumer *kafka.Consumer
	service  service.OrderService
	logger   logger.Logger
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(cfg *kafka.ConfigMap, topic string, service service.OrderService, logger logger.Logger) (*KafkaConsumer, error) {
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
		logger:   logger,
	}, nil
}

// Run starts the consumer loop
func (kc *KafkaConsumer) Run(ctx context.Context) {
	// TODO: dont hardcode it
	const workerCount = 4

	msgch := make(chan *kafka.Message, workerCount*2) // <-- buf size might need an upd
	var wg sync.WaitGroup

	// workers
	for i := range workerCount {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for msg := range msgch {
				// TODO: sentinel errs
				var order domain.Order
				if err := json.Unmarshal(msg.Value, &order); err != nil {
					kc.logger.Errorw("Failed to unmarshal message", "error", err)
					continue
				}
				if err := kc.service.ProcessNewOrder(ctx, &order); err != nil {
					kc.logger.Errorw("Fail to process order", "error", err)
					continue
				}
				// TODO: commit offset
			}
		}(i)
	}

	// the poll goroutine
	go func() {
		defer close(msgch)
		for {
			select {
			case <-ctx.Done():
				kc.logger.Infow("Shutting down the consumer...")
				kc.consumer.Close()
				return
			default:
				ev := kc.consumer.Poll(100)
				if ev == nil {
					continue
				}
				switch e := ev.(type) {
				case *kafka.Message:
					msgch <- e
				case kafka.Error:
					kc.logger.Errorw("Kafka error", "error", e)
				}
			}
		}
	}()

	wg.Wait()
}
