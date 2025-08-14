package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/health"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"
	"github.com/goinginblind/l0-task/internal/store"
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

	if err := w.service.ProcessNewOrder(w.ctx, &order); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidOrder):
			// TODO III: DLQ
			w.logger.Warnw("Invalid order received, discarding", "order_uid", order.OrderUID, "error", err)
			w.commit(msg)

		case errors.Is(err, store.ErrAlreadyExists):
			// TODO III: DLQ
			w.logger.Warnw("Order already exists, discarding", "order_uid", order.OrderUID, "error", err)
			w.commit(msg)

		case errors.Is(err, store.ErrConnectionFailed):
			// TODO I: retry backoff transient errors (short ones where the retry interval is ~10ms)
			w.logger.Errorw("Worker failed to process order due to DB connection error.", "order_uid", order.OrderUID, "error", err)
			w.healthChecker.MarkUnhealthy()
			return

		default:
			// TODO I: unknown errors == msg not commited, DLQ
			w.logger.Errorw("Failed to process order with an unhandled error", "order_uid", order.OrderUID, "error", err)
			return
		}
	}
	// success == commit
	w.logger.Infow("order successfully processed", "order_uid", order.OrderUID)
	w.commit(msg)
}

func (w *worker) commit(msg *kafka.Message) {
	if _, err := w.consumer.CommitMessage(msg); err != nil {
		w.logger.Errorw("Failed to commit message", "error", err)
	}
}
