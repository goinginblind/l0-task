package consumer

import (
	"context"
	"fmt"
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
	consumer      MessageConsumer
	service       service.OrderService
	logger        logger.Logger
	healthChecker *health.DBHealthChecker
	workerCount   int
	jobBuffer     int
	maxRetries    int           // passed to workers
	retryBackoff  time.Duration // passed to workers
	dlqTopic      string        // passed to workers
	dlqPublisher  DLQManager    // passed to workers
}

// Message consumer is an interface which the conrcete kafka.Consumer implements.
// It is needed to provide (some) decoupling and ease the test implementation.
type MessageConsumer interface {
	Subscribe(topic string, rebalanceCb kafka.RebalanceCb) error
	Poll(timeoutMs int) kafka.Event
	Assignment() (partitions []kafka.TopicPartition, err error)
	Pause(partitions []kafka.TopicPartition) error
	Resume(partitions []kafka.TopicPartition) error
	Close() error
	CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error)
}

// DLQManager is an interface which the concrete kafka.Producer implements.
// Again, made to ease testing
type DLQManager interface {
	Events() chan kafka.Event
	Close()
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(kafCfg config.KafkaConfig, consCfg config.ConsumerConfig,
	service service.OrderService, log logger.Logger, hc *health.DBHealthChecker) (*KafkaConsumer, error) {
	consumerKafkaConfig := &kafka.ConfigMap{
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
	consumer, err := kafka.NewConsumer(consumerKafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("fail to create consumer: %v", err)
	}

	producerCfg := &kafka.ConfigMap{
		"bootstrap.servers":   kafCfg.BootstrapServers,
		"acks":                consCfg.DLQ.Acks,
		"retries":             consCfg.DLQ.Retries,
		"delivery.timeout.ms": consCfg.DLQ.DeliveryTimeout,
		"linger.ms":           consCfg.DLQ.Linger,
		"batch.size":          consCfg.DLQ.BatchSize,
		"enable.idempotence":  consCfg.DLQ.EnableIdempotence,
	}
	dlqPublisher, err := kafka.NewProducer(producerCfg)
	if err != nil {
		return nil, fmt.Errorf("fail to create dlq-producer: %v", err)
	}

	if err := consumer.Subscribe(consCfg.Topic, nil); err != nil {
		return nil, fmt.Errorf("fail to subscribe: %v", err)
	}

	return &KafkaConsumer{
		consumer:      consumer,
		service:       service,
		logger:        log,
		healthChecker: hc,
		workerCount:   consCfg.WorkerCount,
		jobBuffer:     consCfg.JobBufferSize,
		maxRetries:    consCfg.MaxRetries,
		retryBackoff:  consCfg.RetryBackoff,
		dlqTopic:      consCfg.DLQ.Topic,
		dlqPublisher:  dlqPublisher,
	}, nil
}

// Run starts the consumer loop
func (kc *KafkaConsumer) Run(ctx context.Context) {
	// this goroutine is a background one, it makes sure the dlq send does not block and
	// that we can still use the fire-and-forget approach for the sendToDLQ function.
	go drainDLQReports(ctx, kc.dlqPublisher, kc.logger)

	jobs := make(chan *kafka.Message, kc.jobBuffer)
	var wg sync.WaitGroup

	// workers
	wDeps := workerDependencies{
		service:       kc.service,
		logger:        kc.logger,
		consumer:      kc.consumer,
		ctx:           ctx,
		healthChecker: kc.healthChecker,
		dlqTopic:      kc.dlqTopic,
		dlqPublisher:  kc.dlqPublisher,
	}

	for i := 0; i < kc.workerCount; i++ {
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
				return
			default:
				kc.manageConsumerState(&isPaused)

				ev := kc.consumer.Poll(100) // this poll ensures the consumer doesn't disconnect from kafka
				if ev == nil {
					time.Sleep(50 * time.Millisecond) // this is to not burn the CPU to ashes when nothing happens
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					if !isPaused {
						jobs <- e
					}
				case kafka.AssignedPartitions:
					kc.logger.Infow("Partitions assigned", "partitions", e.Partitions)
				case kafka.RevokedPartitions:
					kc.logger.Infow("Partitions revoked", "partitions", e.Partitions)
				case kafka.Error:
					kc.logger.Errorw("Kafka error", "error", e, "is_fatal", e.IsFatal())
				}
			}
		}
	}()

	wg.Wait()
	kc.dlqPublisher.Close()
	kc.consumer.Close()
}

// manageConsumerState pauses or resumes the consumer based on DB health.
func (kc *KafkaConsumer) manageConsumerState(isPaused *bool) {
	// db is NOT healthy and consumer is NOT paused: log, pause
	if !kc.healthChecker.IsHealthy() && !*isPaused {
		time.Sleep(100 * time.Millisecond) // avoid hot spins when the db is down
		assignedPartitions, err := kc.consumer.Assignment()
		if err == nil && len(assignedPartitions) > 0 {
			kc.logger.Warnw("DB is unhealthy. Pausing consumption on partitions.", "partitions", assignedPartitions)
			if err := kc.consumer.Pause(assignedPartitions); err != nil {
				kc.logger.Errorw("Failed to pause consumer", "error", err)
			} else {
				*isPaused = true
			}
		}
		// db IS healthy and consumer IS paused: log, unpause
	} else if kc.healthChecker.IsHealthy() && *isPaused {
		assignedPartitions, err := kc.consumer.Assignment()
		if err == nil && len(assignedPartitions) > 0 {
			kc.logger.Infow("DB is healthy again. Resuming consumption on partitions.", "partitions", assignedPartitions)
			if err := kc.consumer.Resume(assignedPartitions); err != nil {
				kc.logger.Errorw("Failed to resume consumer", "error", err)
			} else {
				*isPaused = false
			}
		}
	}
}
