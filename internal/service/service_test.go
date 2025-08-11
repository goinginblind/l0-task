package service

import (
	"context"
	"errors"
	"github.com/goinginblind/l0-task/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderStore is a mock implementation of the OrderStore interface.
type MockOrderStore struct {
	mock.Mock
}

func (m *MockOrderStore) Insert(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderStore) Get(ctx context.Context, uid string) (*domain.Order, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).(*domain.Order), args.Error(1)
}

func TestOrderService_ProcessNewOrder(t *testing.T) {
	mockStore := new(MockOrderStore)
	service := New(mockStore)

	ctx := context.Background()
	order := &domain.Order{
		OrderUID:    "testuid",
		TrackNumber: "testtrack",
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "12345",
			City:    "Test City",
			Address: "Test Address",
			Region:  "Test Region",
			Email:   "test@example.com",
		},
		Payment: domain.Payment{
			Transaction:  "testuid",
			Currency:     "USD",
			Provider:     "testprovider",
			Amount:       100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "testbank",
			DeliveryCost: 10,
			GoodsTotal:   90,
			CustomFee:    0,
		},
		Items: []domain.Item{
			{
				ChrtID:      1,
				TrackNumber: "testtrack",
				Price:       90,
				Rid:         "testrid",
				Name:        "Test Item",
				Sale:        0,
				Size:        "M",
				TotalPrice:  90,
				NmID:        123,
				Brand:       "Test Brand",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "testcustomer",
		DeliveryService: "testservice",
		ShardKey:        "1",
		SmID:            1,
		DateCreated:     time.Now(),
		OofShard:        "1",
	}

	t.Run("success", func(t *testing.T) {
		mockStore.On("Insert", ctx, order).Return(nil).Once()

		err := service.ProcessNewOrder(ctx, order)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("validation failed", func(t *testing.T) {
		invalidOrder := &domain.Order{} // Missing required fields
		err := service.ProcessNewOrder(ctx, invalidOrder)
		assert.Error(t, err)
	})

	t.Run("store failed", func(t *testing.T) {
		mockStore.On("Insert", ctx, order).Return(errors.New("store error")).Once()

		err := service.ProcessNewOrder(ctx, order)

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}

func TestOrderService_GetOrder(t *testing.T) {
	mockStore := new(MockOrderStore)
	service := New(mockStore)

	ctx := context.Background()
	uid := "test-uid"
	order := &domain.Order{OrderUID: uid}

	t.Run("success", func(t *testing.T) {
		mockStore.On("Get", ctx, uid).Return(order, nil).Once()

		retrievedOrder, err := service.GetOrder(ctx, uid)

		assert.NoError(t, err)
		assert.Equal(t, order, retrievedOrder)
		mockStore.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStore.On("Get", ctx, uid).Return((*domain.Order)(nil), errors.New("not found")).Once()

		_, err := service.GetOrder(ctx, uid)

		assert.Error(t, err)
		mockStore.AssertExpectations(t)
	})
}
