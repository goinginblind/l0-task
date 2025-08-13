package service

import (
	"context"
	"fmt"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
)

// OrderStore defines the interface for storing and retrieving orders.
// It's implemeneted by the database layer (the store), describes the 'save' and 'load' contract.
type OrderStore interface {
	Insert(context.Context, *domain.Order) error
	Get(context.Context, string) (*domain.Order, error)
}

// OrderService defines the interface for handling orders.
// It's used to transport Orders around, used by the api and consumer packages.
// This is the buisness logic contract.
type OrderService interface {
	ProcessNewOrder(context.Context, *domain.Order) error
	GetOrder(context.Context, string) (*domain.Order, error)
}

// New creates a new OrderService.
func New(store OrderStore, logger logger.Logger) OrderService {
	return &orderService{
		store:  store,
		logger: logger,
	}
}

// orderService provides the core business logic for handling orders.
// Used to encapsulate/isolate the database.
type orderService struct {
	store  OrderStore
	logger logger.Logger
}

// ProcessNewOrder validates and stores a new order.
func (s *orderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	if err := order.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.store.Insert(ctx, order); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	s.logger.Infow("order successfully processed", "order_uid", order.OrderUID)

	return nil
}

// GetOrder retrieves an order by its UID. Currently looks like a wrapper. And it is.
// Later won't be. Caching, baby. Allows the buisness logic of retrieving an order to be pretty && clean.
func (s *orderService) GetOrder(ctx context.Context, uid string) (*domain.Order, error) {
	order, err := s.store.Get(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}
