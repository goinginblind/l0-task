package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const (
	testDbUser = "test_user"
	testDbPass = "test_pass"
	testDbName = "test_orders_db"
	testDbHost = "localhost"
	testDbPort = "5433"
)

var testModel *OrdersModel

func TestMain(m *testing.M) {
	// Start docker container
	cmd := exec.Command("docker-compose", "up", "-d", "postgres-test")
	err := cmd.Run()
	if err != nil {
		fmt.Println("could not start docker-compose:", err)
		os.Exit(1)
	}

	// give db time to start
	time.Sleep(5 * time.Second)

	// Run migrations
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		testDbUser, testDbPass, testDbName, testDbHost, testDbPort)

	// goose migration
	cmd = exec.Command("goose", "-dir", "../../sql", "postgres", dsn, "up")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("could not run migrations: %s\n%s", err, string(out))
		os.Exit(1)
	}

	// connect to DB
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Println("could not connect to test db:", err)
		os.Exit(1)
	}
	if err := db.Ping(); err != nil {
		fmt.Println("could not ping test db:", err)
		os.Exit(1)
	}

	testModel = &OrdersModel{DB: db}

	// run tests
	code := m.Run()

	// teardown
	cmd = exec.Command("docker-compose", "down")
	err = cmd.Run()
	if err != nil {
		fmt.Println("could not stop docker-compose:", err)
	}

	os.Exit(code)
}

func TestOrdersModel_Integration(t *testing.T) {
	// Truncate tables before test to ensure clean state
	_, err := testModel.DB.Exec("TRUNCATE orders, deliveries, payments, items RESTART IDENTITY CASCADE;")
	require.NoError(t, err)

	// Create a sample order
	order := &Order{
		OrderUID:    "testuid123",
		TrackNumber: "trackno1",
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
			Transaction:  "testuid123",
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
				TrackNumber: "trackno1",
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
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SmID:              99,
		DateCreated:       time.Now().UTC().Truncate(time.Second),
		OofShard:          "1",
	}

	ctx := context.Background()

	t.Run("Insert and Exists", func(t *testing.T) {
		// Check that it doesn't exist initially
		exists, err := testModel.Exists(ctx, order.OrderUID)
		require.NoError(t, err)
		require.False(t, exists)

		// Insert the order
		err = testModel.Insert(ctx, order)
		require.NoError(t, err)

		// Check that it now exists
		exists, err = testModel.Exists(ctx, order.OrderUID)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("GetJson", func(t *testing.T) {
		// Get the order back as JSON
		jsonBytes, err := testModel.GetJson(ctx, order.OrderUID)
		require.NoError(t, err)
		require.NotNil(t, jsonBytes)

		// Unmarshal and verify
		var retrievedOrder Order
		err = json.Unmarshal(jsonBytes, &retrievedOrder)
		require.NoError(t, err)

		// Compare a few key fields
		require.Equal(t, order.OrderUID, retrievedOrder.OrderUID)
		require.Equal(t, order.Delivery.Name, retrievedOrder.Delivery.Name)
		require.Equal(t, order.Payment.Transaction, retrievedOrder.Payment.Transaction)
		require.Equal(t, len(order.Items), len(retrievedOrder.Items))
		require.Equal(t, order.Items[0].ChrtID, retrievedOrder.Items[0].ChrtID)
		require.True(t, order.DateCreated.Equal(retrievedOrder.DateCreated), "timestamps should be equal")
	})
}
