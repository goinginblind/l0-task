package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goinginblind/l0-task/internal/config"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderService is a mock implementation of the OrderService.
type MockOrderService struct {
	mock.Mock
}

var _ service.OrderService = (*MockOrderService)(nil)

func (m *MockOrderService) ProcessNewOrder(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderService) GetOrder(ctx context.Context, uid string) (*domain.Order, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func TestServer_orderHandler(t *testing.T) {
	mockService, mockLogger := new(MockOrderService), logger.NewMockLogger()
	server, _ := NewServer(mockService, mockLogger, config.HTTPServerConfig{})

	t.Run("success", func(t *testing.T) {
		order := &domain.Order{OrderUID: "test-uid"}
		mockService.On("GetOrder", mock.Anything, "test-uid").Return(order, nil).Once()

		req := httptest.NewRequest("GET", "/orders/test-uid", nil)
		rr := httptest.NewRecorder()

		server.orderView(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "test-uid")
		mockService.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockService.On("GetOrder", mock.Anything, "not-found-uid").Return(nil, nil).Once()

		req := httptest.NewRequest("GET", "/orders/not-found-uid", nil)
		rr := httptest.NewRecorder()

		server.orderView(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
}
