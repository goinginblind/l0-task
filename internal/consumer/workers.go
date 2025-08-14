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
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
)

// TODO III: don't leave hardcoded
const (
	maxRetries   = 3
	retryBackoff = 250 * time.Millisecond
)

type worker struct {
	id            int
	service       service.OrderService
	logger        logger.Logger
	consumer      *kafka.Consumer
	jobs          <-chan *kafka.Message
	ctx           context.Context
	healthChecker *health.DBHealthChecker
}

func (w *worker) run(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-w.ctx.Done():
			w.logger.Infow("Worker shutting down", "worker_id", w.id)
			return
		case msg, ok := <-w.jobs:
			if !ok {
				return
			}
			w.processMessage(msg)
		}
	}
}

func (w *worker) processMessage(msg *kafka.Message) {
	var order domain.Order
	dec := json.NewDecoder(bytes.NewReader(msg.Value))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&order); err != nil {
		w.logger.Errorw("Failed to unmarshal message, discarding", "error", err)
		w.commit(msg)
		return
	}

	var processErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		processErr = w.service.ProcessNewOrder(w.ctx, &order)
		if processErr == nil {
			// success == commit, exit
			w.commit(msg)
			return
		}

		// any other error == break the loop immediately
		if !errors.Is(processErr, store.ErrConnectionFailed) {
			break
		}

		// If it was a connection error, log a warning and wait before the next attempt.
		if attempt < maxRetries {
			w.logger.Warnw("Transient DB connection error, will retry.",
				"order_uid", order.OrderUID,
				"attempt", attempt,
				"retry_in", retryBackoff,
				"error", processErr,
			)
			time.Sleep(retryBackoff)
		}
	}

	// inspect `processErr` to decide what to do
	w.logger.Errorw("Failed to process order after all attempts.",
		"order_uid", order.OrderUID,
		"attempts", maxRetries,
		"final_error", processErr,
	)

	switch {
	case errors.Is(processErr, domain.ErrInvalidOrder):
		// TODO III: DLQ
		w.logger.Warnw("Invalid order received, discarding",
			"order_uid", order.OrderUID,
			"error", processErr,
		)

	case errors.Is(processErr, store.ErrAlreadyExists):
		// TODO III: DLQ
		w.logger.Warnw("Order already exists, discarding",
			"order_uid", order.OrderUID,
			"error", processErr,
		)

	case errors.Is(processErr, store.ErrConnectionFailed):
		w.logger.Errorw("Worker failed to process order due to DB connection error.",
			"order_uid", order.OrderUID,
			"error", processErr,
		)
		w.healthChecker.MarkUnhealthy()
		return

	default:
		// TODO III: unknown errors == msg commited, DLQ
		w.logger.Errorw("Failed to process order with an unhandled error", "order_uid", order.OrderUID, "error", processErr)
		w.commit(msg) // <-- placeholder
		return
	}
	// success == commit
	w.logger.Infow("order successfully processed", "worker_id", w.id, "order_uid", order.OrderUID)
	w.commit(msg)
}

func (w *worker) commit(msg *kafka.Message) {
	if _, err := w.consumer.CommitMessage(msg); err != nil {
		w.logger.Errorw("Failed to commit message", "error", err)
	}
}
