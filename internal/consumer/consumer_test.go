package consumer

import (
	"context"
	"testing"

	"github.com/goinginblind/l0-task/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(ctx context.Context, uid string) (*domain.Order, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func TestKafkaConsumer_Run(t *testing.T) {
	// Well, behold, a placeholder
}
