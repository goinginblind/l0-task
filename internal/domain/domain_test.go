package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getValidOrder() *Order {
	return &Order{
		OrderUID:    "testuid123",
		TrackNumber: "testtrack",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test Testov",
			Phone:   "+972501234567",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@example.com",
		},
		Payment: Payment{
			Transaction:  "testuid123",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      9934930,
				TrackNumber: "testtrack",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389222,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:          "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}
}

func TestOrder_Validate(t *testing.T) {
	t.Run("valid order", func(t *testing.T) {
		order := getValidOrder()
		err := order.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid order - missing required field", func(t *testing.T) {
		order := getValidOrder()
		order.OrderUID = "" // Missing required field
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("invalid order - invalid email format", func(t *testing.T) {
		order := getValidOrder()
		order.Delivery.Email = "invalid-email"
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("invalid order - invalid phone format", func(t *testing.T) {
		order := getValidOrder()
		order.Delivery.Phone = "12345" // Invalid E.164 format
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("invalid order - item price less than 0", func(t *testing.T) {
		order := getValidOrder()
		order.Items[0].Price = -1 // Invalid price
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("invalid order - empty items list", func(t *testing.T) {
		order := getValidOrder()
		order.Items = []Item{} // Empty items list
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})

	t.Run("invalid order - sm_id less than or equal to 0", func(t *testing.T) {
		order := getValidOrder()
		order.SmID = 0 // Invalid sm_id
		err := order.Validate()
		assert.ErrorIs(t, err, ErrInvalidOrder)
	})
}
