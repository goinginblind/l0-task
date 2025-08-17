package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/health"
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
	consumer      *kafka.Consumer
	ctx           context.Context
	healthChecker *health.DBHealthChecker
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
			w.processMessage(msg)
		}
	}
}

// process message does basic msg content processing:
//   - decodes json (disallowing foreign fields)
//   - passes this down to service layer (validation and db write)
//   - in case of the db being unavailable it notifies the consumer
//   - transient db hiccups are handled with retries
//   - other errors either discarded or sent to DLQ
//   - metrics are scraped with either 'error', 'valid' or 'invalid' labels
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

	var processErr error
	for attempt := 1; attempt <= w.maxRetries; attempt++ {
		metrics.DbTransientErrors.Inc()
		processErr = w.deps.service.ProcessNewOrder(w.deps.ctx, &order)
		if processErr == nil {
			// success == commit, exit
			metrics.MessagesProcessedTotal.WithLabelValues("valid").Inc()
			w.deps.logger.Infow("order successfully processed", "worker_id", w.id, "order_uid", order.OrderUID)
			w.commit(msg)
			return
		}

		// any other error == break the loop immediately
		if !errors.Is(processErr, store.ErrConnectionFailed) {
			break
		}

		// If it was a connection error, log a warning and wait before the next attempt.
		if attempt < w.maxRetries {
			w.deps.logger.Warnw("Transient DB connection error, will retry.",
				"order_uid", order.OrderUID,
				"attempt", attempt,
				"retry_in", w.retryBackoff,
				"error", processErr,
			)
			time.Sleep(w.retryBackoff)
		}
	}

	// inspect `processErr` to decide what to do (and log it)
	w.deps.logger.Errorw("Failed to process order after all attempts.",
		"order_uid", order.OrderUID,
		"attempts", w.maxRetries,
		"final_error", processErr,
	)

	switch {
	case errors.Is(processErr, domain.ErrInvalidOrder):
		// TODO III: DLQ
		metrics.MessagesProcessedTotal.WithLabelValues("invalid").Inc()
		w.deps.logger.Warnw("Invalid order received, discarding",
			"order_uid", order.OrderUID,
			"error", processErr,
		)

	case errors.Is(processErr, store.ErrAlreadyExists):
		// TODO III: DLQ
		metrics.MessagesProcessedTotal.WithLabelValues("invalid").Inc()
		w.deps.logger.Warnw("Order already exists, discarding",
			"order_uid", order.OrderUID,
			"error", processErr,
		)

	case errors.Is(processErr, store.ErrConnectionFailed):
		metrics.MessagesProcessedTotal.WithLabelValues("error").Inc()
		w.deps.logger.Errorw("Worker failed to process order due to DB connection error.",
			"order_uid", order.OrderUID,
			"error", processErr,
		)
		w.deps.healthChecker.MarkUnhealthy()
		return // no commit since the message isn't processed, kafka will resend when db is up

	default:
		// TODO III: unknown errors == msg commited, DLQ
		metrics.MessagesProcessedTotal.WithLabelValues("error").Inc()
		w.deps.logger.Errorw("Failed to process order with an unhandled error", "order_uid", order.OrderUID, "error", processErr)
	}

	w.commit(msg)
}

func (w *worker) commit(msg *kafka.Message) {
	if msg == nil {
		return
	}

	_, err := w.deps.consumer.CommitMessage(msg)
	if err != nil {
		w.deps.logger.Errorw("Failed to commit message", "error", err)
	}
}
