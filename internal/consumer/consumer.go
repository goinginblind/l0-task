package consumer

import (
	"context"
	"sync"
	"time"

	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/pkg/health"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaConsumer consumes messages from Kafka and processes them.
type KafkaConsumer struct {
	consumer      *kafka.Consumer
	service       service.OrderService
	logger        logger.Logger
	healthChecker *health.DBHealthChecker
	workerCount   int
	jobBuffer     int
	maxRetries    int           // passed to workers
	retryBackoff  time.Duration // passed to workers
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(kafCfg config.KafkaConfig, consCfg config.ConsumerConfig,
	service service.OrderService, logger logger.Logger, hc *health.DBHealthChecker) (*KafkaConsumer, error) {
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers":     kafCfg.BootstrapServers,
		"group.id":              kafCfg.ConsumerGroupID,
		"auto.offset.reset":     kafCfg.AutoOffsetReset,
		"enable.auto.commit":    kafCfg.EnableAutoCommit,
		"isolation.level":       kafCfg.IsolationLevel,
		"max.poll.interval.ms":  kafCfg.MaxPollIntervalMs,
		"fetch.min.bytes":       kafCfg.MinFetchSizeBytes,
		"fetch.max.bytes":       kafCfg.MaxFetchSizeBytes,
		"session.timeout.ms":    kafCfg.SessionTimeoutMs,
		"heartbeat.interval.ms": kafCfg.HeartbeatIntervalMs,
	}
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, err
	}

	if err := consumer.Subscribe(consCfg.Topic, nil); err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		consumer:      consumer,
		service:       service,
		logger:        logger,
		healthChecker: hc,
		workerCount:   consCfg.WorkerCount,
		jobBuffer:     consCfg.JobBufferSize,
		maxRetries:    consCfg.MaxRetries,
		retryBackoff:  consCfg.RetryBackoff,
	}, nil
}

// Run starts the consumer loop
func (kc *KafkaConsumer) Run(ctx context.Context) {
	jobs := make(chan *kafka.Message, kc.jobBuffer*2)
	var wg sync.WaitGroup

	// workers
	wDeps := workerDependencies{
		service:       kc.service,
		logger:        kc.logger,
		consumer:      kc.consumer,
		ctx:           ctx,
		healthChecker: kc.healthChecker,
	}

	for i := range kc.workerCount {
		wg.Add(1)
		w := &worker{
			id:           i,
			jobs:         jobs,
			deps:         wDeps,
			maxRetries:   kc.maxRetries,
			retryBackoff: kc.retryBackoff,
		}
		go w.run(&wg)
	}

	// the poll goroutine: handles shutdown, db long disconnects and sends messages to workers
	go func() {
		defer close(jobs)
		isPaused := false // Local state to track if we've already paused.

		for {
			select {
			case <-ctx.Done():
				kc.logger.Infow("Shutting down the consumer...")
				kc.consumer.Close()
				return
			default:
				// db is NOT healthy and consumer is NOT paused: log, pause
				if !kc.healthChecker.IsHealthy() && !isPaused {
					assignedPartitions, err := kc.consumer.Assignment()
					if err == nil && len(assignedPartitions) > 0 {
						kc.logger.Warnw("DB is unhealthy. Pausing consumption on partitions.", "partitions", assignedPartitions)
						if err := kc.consumer.Pause(assignedPartitions); err != nil {
							kc.logger.Errorw("Failed to pause consumer", "error", err)
						} else {
							isPaused = true
						}
					}
					// db IS healthy and consumer IS paused: log, unpause
				} else if kc.healthChecker.IsHealthy() && isPaused {
					assignedPartitions, err := kc.consumer.Assignment()
					if err == nil && len(assignedPartitions) > 0 {
						kc.logger.Infow("DB is healthy again. Resuming consumption on partitions.", "partitions", assignedPartitions)
						if err := kc.consumer.Resume(assignedPartitions); err != nil {
							kc.logger.Errorw("Failed to resume consumer", "error", err)
						} else {
							isPaused = false
						}
					}
				}

				ev := kc.consumer.Poll(100) // this poll ensures the consumer doesn't disconnect from kafka
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					if !isPaused {
						jobs <- e
					}
				case kafka.Error:
					// TODO III: handle AssignedPartitions/RevokedPartitions events here
					kc.logger.Errorw("Kafka error", "error", e, "is_fatal", e.IsFatal())
				}
			}
		}
	}()

	wg.Wait()
}
