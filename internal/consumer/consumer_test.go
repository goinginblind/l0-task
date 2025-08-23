package consumer

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/goinginblind/l0-task/internal/store"
	"github.com/stretchr/testify/mock"
)

// MockOrderService simulates the service layer.
type MockOrderService struct {
	mock.Mock
}

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

type MockCommitter struct {
	mock.Mock
}

func (m *MockCommitter) CommitMessage(msg *kafka.Message) ([]kafka.TopicPartition, error) {
	args := m.Called(msg)
	return nil, args.Error(0)
}

type MockDLQProducer struct {
	mock.Mock
}

func (m *MockDLQProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	args := m.Called(msg, deliveryChan)
	return args.Error(0)
}

type MockUnhealthyMarker struct {
	mock.Mock
}

func (m *MockUnhealthyMarker) MarkUnhealthy() {
	m.Called()
}

func TestWorker_ProcessMessage(t *testing.T) {
	// standard valid order and Kafka message to reuse
	validOrder := domain.Order{OrderUID: "test-uid"}
	validOrderJSON, _ := json.Marshal(validOrder)
	kafkaMsg := &kafka.Message{Value: validOrderJSON, Key: []byte(validOrder.OrderUID)}

	testCases := []struct {
		name            string
		message         *kafka.Message
		setupMocks      func(*MockOrderService, *MockCommitter, *MockDLQProducer, *MockUnhealthyMarker)
		maxRetries      int
		retryBackoff    time.Duration
		expectCommit    bool
		expectDLQ       bool
		expectUnhealthy bool
	}{
		{
			name:    "Success - Happy Path",
			message: kafkaMsg,
			setupMocks: func(s *MockOrderService, c *MockCommitter, p *MockDLQProducer, h *MockUnhealthyMarker) {
				s.On("ProcessNewOrder", mock.Anything, &validOrder).Return(nil).Once()
				c.On("CommitMessage", kafkaMsg).Return(nil).Once()
			},
			maxRetries:   3,
			retryBackoff: 1 * time.Millisecond,
			expectCommit: true,
			expectDLQ:    false,
		},
		{
			name:    "Failure - Invalid JSON",
			message: &kafka.Message{Value: []byte("not-json")},
			setupMocks: func(s *MockOrderService, c *MockCommitter, p *MockDLQProducer, h *MockUnhealthyMarker) {
				c.On("CommitMessage", mock.Anything).Return(nil).Once()
			},
			maxRetries:   3,
			retryBackoff: 1 * time.Millisecond,
			expectCommit: true,
			expectDLQ:    false,
		},
		{
			name:    "Failure - Invalid Order Data (sent to DLQ)",
			message: kafkaMsg,
			setupMocks: func(s *MockOrderService, c *MockCommitter, p *MockDLQProducer, h *MockUnhealthyMarker) {
				s.On("ProcessNewOrder", mock.Anything, &validOrder).Return(domain.ErrInvalidOrder).Once()
				p.On("Produce", mock.Anything, mock.Anything).Return(nil).Once()
				// commit should move past the bad message
				c.On("CommitMessage", kafkaMsg).Return(nil).Once()
			},
			maxRetries:   3,
			retryBackoff: 1 * time.Millisecond,
			expectCommit: true,
			expectDLQ:    true,
		},
		{
			name:    "Success - Transient DB Error with Recovery",
			message: kafkaMsg,
			setupMocks: func(s *MockOrderService, c *MockCommitter, p *MockDLQProducer, h *MockUnhealthyMarker) {
				// 1st call fails == temporary DB issue
				// 2nd call succeeds, simulating recovery
				// 3rd a commit because the process ultimately succeeds
				s.On("ProcessNewOrder", mock.Anything, &validOrder).Return(store.ErrConnectionFailed).Once()
				s.On("ProcessNewOrder", mock.Anything, &validOrder).Return(nil).Once()
				c.On("CommitMessage", kafkaMsg).Return(nil).Once()
			},
			maxRetries:   3,
			retryBackoff: 1 * time.Millisecond,
			expectCommit: true,
			expectDLQ:    false,
		},
		{
			name:    "Failure - Permanent DB Error",
			message: kafkaMsg,
			setupMocks: func(s *MockOrderService, c *MockCommitter, p *MockDLQProducer, h *MockUnhealthyMarker) {
				s.On("ProcessNewOrder", mock.Anything, &validOrder).Return(store.ErrConnectionFailed)
				h.On("MarkUnhealthy").Return().Once()
			},
			maxRetries:      3,
			retryBackoff:    1 * time.Millisecond,
			expectCommit:    false, // DO NOT commit on permanent DB failure, so Kafka can redeliver
			expectDLQ:       false,
			expectUnhealthy: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := new(MockOrderService)
			mockCommitter := new(MockCommitter)
			mockDLQProducer := new(MockDLQProducer)
			mockHealthChecker := new(MockUnhealthyMarker)

			tc.setupMocks(mockService, mockCommitter, mockDLQProducer, mockHealthChecker)

			w := &worker{
				id: 1,
				deps: workerDependencies{
					service:       mockService,
					logger:        logger.NewMockLogger(),
					consumer:      mockCommitter,
					ctx:           context.Background(),
					healthChecker: mockHealthChecker,
					dlqTopic:      "test-dlq",
					dlqPublisher:  mockDLQProducer,
				},
				maxRetries:   tc.maxRetries,
				retryBackoff: tc.retryBackoff,
			}
			w.processMessage(tc.message)

			mockService.AssertExpectations(t)
			mockHealthChecker.AssertExpectations(t)

			if tc.expectCommit {
				mockCommitter.AssertCalled(t, "CommitMessage", mock.Anything)
			} else {
				mockCommitter.AssertNotCalled(t, "CommitMessage", mock.Anything)
			}

			if tc.expectDLQ {
				mockDLQProducer.AssertCalled(t, "Produce", mock.Anything, mock.Anything)
			} else {
				mockDLQProducer.AssertNotCalled(t, "Produce", mock.Anything, mock.Anything)
			}

			if tc.expectUnhealthy {
				mockHealthChecker.AssertCalled(t, "MarkUnhealthy")
			} else {
				mockHealthChecker.AssertNotCalled(t, "MarkUnhealthy")
			}
		})
	}
}
