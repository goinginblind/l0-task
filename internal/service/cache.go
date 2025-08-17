package service

import (
	"context"
	"errors"
	"time"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/pkg/metrics"
	"github.com/goinginblind/l0-task/internal/store"
)

// CachingOrderService is a decorator that adds in-memory caching to OrderService
type CachingOrderService struct {
	next   OrderService
	store  OrderStore // ensure preloads
	cache  *LRUCache
	logger logger.Logger
}

// NewCachingOrderService creates a caching decorator for OrderService
func NewCachingOrderService(next OrderService, store OrderStore, logger logger.Logger, entryCountCap, entrySizeCap int) *CachingOrderService {
	return &CachingOrderService{
		next:   next,
		store:  store,
		cache:  NewLRUCache(entryCountCap, entrySizeCap),
		logger: logger,
	}
}

// GetOrder gets the order from the store layer, but befor doing so it
// checks the cache. If hit - early return, but if not - cache gets updated.
func (s *CachingOrderService) GetOrder(ctx context.Context, uid string) (*domain.Order, error) {
	start := time.Now()
	order, found := s.cache.Get(uid)
	if found {
		duration := float64(time.Since(start).Seconds())
		metrics.CacheResponseTime.WithLabelValues("get_order").Observe(duration)
		metrics.CacheHits.Inc()

		s.logger.Infow("Cache hit", "order_uid", uid)
		return order, nil
	}

	metrics.CacheMisses.Inc()
	s.logger.Infow("Cache miss", "order_uid", uid)
	order, err := s.next.GetOrder(ctx, uid)
	if err != nil {
		return nil, err
	}

	start = time.Now() // for metrics

	s.cache.Insert(order)

	duration := float64(time.Since(start).Seconds()) // metrics
	metrics.CacheResponseTime.WithLabelValues("insert_order").Observe(duration)

	return order, nil
}

// ProcessNewOrder calls the underlying service to process the order and, on success,
// adds the new order to the cache.
func (s *CachingOrderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	err := s.next.ProcessNewOrder(ctx, order)
	// Uncommenting will make it a sorta write-through style caching (not really, but almost!)
	//if err == nil {
	//	s.logger.Infow("Adding new order to cache", "order_uid", order.OrderUID)
	//	s.cache.Insert(order)
	//}
	return err
}

// Preload is used in case there's already something to cache
func (s *CachingOrderService) Preload(ctx context.Context, limit int) error {
	s.logger.Infow("Preloading cache...")
	orders, err := s.store.GetLatestOrders(ctx, limit)
	if err != nil {
		if errors.Is(err, store.ErrConnectionFailed) {
			s.logger.Warnw("Fail to preload cache, db is down")
			return nil
		}
		return err
	}

	for _, order := range orders {
		s.cache.Insert(order)
	}
	s.logger.Infow("Cache preload complete", "count", len(orders))
	return nil
}
