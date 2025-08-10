package main

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestProcessMessage(t *testing.T) {
	expectedOrder := &Order{
		OrderUID:    "b563feb7b2b84b6test",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: Payment{
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
	}

	testCases := []struct {
		name          string
		filepath      string
		expectedOrder *Order
		expectError   bool
	}{
		{
			name:          "Correct input",
			filepath:      "testdata/proc_ok.json",
			expectedOrder: expectedOrder,
			expectError:   false,
		},
		{
			name:          "Garbage input",
			filepath:      "testdata/proc_garbage.json",
			expectedOrder: nil,
			expectError:   true,
		},
		{
			name:          "Malformed input",
			filepath:      "testdata/proc_malformed.json",
			expectedOrder: nil,
			expectError:   true,
		},
		{
			name:          "Missing order_uid",
			filepath:      "testdata/proc_noouid.json",
			expectedOrder: nil,
			expectError:   true,
		},
		{
			name:          "Unknown fields",
			filepath:      "testdata/proc_unknown.json",
			expectedOrder: nil,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(tc.filepath)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			order, err := ProcessMessage(data)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if diff := cmp.Diff(tc.expectedOrder, order); diff != "" {
					t.Errorf("Mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
