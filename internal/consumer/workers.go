package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/pkg/metrics"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
)

// a worker, well, does work
type worker struct {
	id           int
	deps         workerDependencies
	jobs         <-chan *kafka.Message
	maxRetries   int
	retryBackoff time.Duration
}

// workerDependencies contain all dependencies passed down to
// workers from the kafka consumer. They are shared between all the workers.
type workerDependencies struct {
	service       service.OrderService
	logger        logger.Logger
	consumer      Committer // consumer is passed into worker since the offset commits are manual, so it does need it
	ctx           context.Context
	healthChecker UnhealthyMarker
	dlqTopic      string
	dlqPublisher  DLQProducer
}

type Committer interface {
	CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error)
}

type DLQProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
}

type UnhealthyMarker interface {
	MarkUnhealthy()
}

// run processes the message
func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-w.deps.ctx.Done():
			w.deps.logger.Infow("Worker shutting down", "worker_id", w.id)
			return
		case msg, ok := <-w.jobs:
			if !ok {
				return
			}
			func(msg *kafka.Message) {
				// panic recovery
				// if something unexpected happens (altough it shouldn't normally).
				// Send to DLQ after recovery
				defer func() {
					if r := recover(); r != nil {
						w.deps.logger.Errorw("worker encountered panic",
							"worker id", w.id,
							"message", msg,
							"panic", r,
							"stack", string(debug.Stack()),
						)
						w.sendToDLQ(msg, fmt.Errorf("worker encountered panic: %v", r))
					}
				}()
				w.processMessage(msg)
			}(msg)
		}
	}
}

// processMessage unmarshals the kafka message and then orchestrates the processing
// and result handling (passing down to the helper functions).
func (w *worker) processMessage(msg *kafka.Message) {
	var order domain.Order
	dec := json.NewDecoder(bytes.NewReader(msg.Value))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&order); err != nil {
		metrics.MessagesProcessedTotal.WithLabelValues("invalid").Inc()
		w.deps.logger.Errorw("Failed to unmarshal message, discarding", "error", err)
		w.commit(msg)
		return
	}

	processErr := w.processWithRetries(&order)        // process, extract error or retry if possible
	w.handleProcessingResult(msg, &order, processErr) // error (or nil) goes here, commit or no trough select statement
}

// processWithRetries passes the message down
// to the service layer, contains the retry loop for handling transient DB errors.
func (w *worker) processWithRetries(order *domain.Order) error {
	var processErr error
	for attempt := 0; attempt < w.maxRetries; attempt++ {
		processErr = w.deps.service.ProcessNewOrder(w.deps.ctx, order)
		if processErr == nil {
			return nil // success
		}

		// any other error than a connection error should break the loop immediately
		if !errors.Is(processErr, store.ErrConnectionFailed) {
			break
		}

		metrics.DbTransientErrors.Inc()

		// If it was a connection error, log a warning and wait before the next attempt.
		w.deps.logger.Warnw("Transient DB connection error, will retry.",
			"order_uid", order.OrderUID,
			"attempt", attempt,
			"retry_in", w.retryBackoff,
			"error", processErr,
		)
		// Corrected exponential backoff: 1<<attempt is 2^attempt (степень короче)
		time.Sleep(w.retryBackoff * time.Duration(1<<attempt))
	}
	return processErr
}

// handleProcessingResult inspects the final error and decides what to do.
func (w *worker) handleProcessingResult(msg *kafka.Message, order *domain.Order, processErr error) {
	if processErr == nil {
		metrics.MessagesProcessedTotal.WithLabelValues("valid").Inc()
		w.deps.logger.Infow("order successfully processed", "worker_id", w.id, "order_uid", order.OrderUID)
		w.commit(msg)
		return
	}

	switch {
	case errors.Is(processErr, domain.ErrInvalidOrder):
		metrics.MessagesProcessedTotal.WithLabelValues("invalid").Inc()
		w.deps.logger.Warnw("Invalid order received, sending to DLQ",
			"order_uid", order.OrderUID,
			"error", processErr,
		)
		w.sendToDLQ(msg, processErr)

	case errors.Is(processErr, store.ErrAlreadyExists):
		metrics.MessagesProcessedTotal.WithLabelValues("invalid").Inc()
		w.deps.logger.Warnw("Order already exists, sending to DLQ",
			"order_uid", order.OrderUID,
			"error", processErr,
		)
		w.sendToDLQ(msg, processErr)

	case errors.Is(processErr, store.ErrConnectionFailed):
		metrics.MessagesProcessedTotal.WithLabelValues("error").Inc()
		w.deps.logger.Errorw("Worker failed to process order due to DB connection error.",
			"order_uid", order.OrderUID,
			"error", processErr,
		)
		w.deps.healthChecker.MarkUnhealthy()
		return // no commit since the message isn't processed, kafka will resend when db is up

	default:
		metrics.MessagesProcessedTotal.WithLabelValues("error").Inc()
		w.deps.logger.Errorw("Failed to process order with an unhandled error, sending to DLQ", "order_uid", order.OrderUID, "error", processErr)
		w.sendToDLQ(msg, processErr)
	}

	w.commit(msg)
}

// commit commits the message
func (w *worker) commit(msg *kafka.Message) {
	if msg == nil {
		return
	}

	_, err := w.deps.consumer.CommitMessage(msg)
	if err != nil {
		w.deps.logger.Errorw("Failed to commit message", "error", err)
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
