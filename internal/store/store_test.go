package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/goinginblind/l0-task/internal/domain"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

const (
	testDbUser = "test_user"
	testDbPass = "test_pass"
	testDbName = "test_orders_db"
	testDbHost = "localhost"
	testDbPort = "5433"
)

var testStore *DBStore

func TestMain(m *testing.M) {
	// Variable declarations
	var cmd *exec.Cmd
	var out []byte
	var err error

	// Start docker container
	cmd = exec.Command("docker", "compose", "up", "-d", "postgres-test")
	err = cmd.Run()
	if err != nil {
		fmt.Println("could not start docker-compose:", err)
		os.Exit(1)
	}

	// Connect to DB with retries
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDbUser, testDbPass, testDbHost, testDbPort, testDbName,
	)
	var db *sql.DB
	for range 10 {
		db, err = sql.Open("pgx", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		fmt.Println("waiting for db...")
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		fmt.Println("could not connect to test db:", err)
		os.Exit(1)
	}

	// Run migrations
	cmd = exec.Command("goose", "-dir", "../../sql", "postgres", dsn, "up")
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("could not run migrations: %s\n%s", err, string(out))
		os.Exit(1)
	}

	testStore = NewDBStore(db)

	// Run tests
	code := m.Run()

	// Teardown
	cmd = exec.Command("docker", "compose", "down", "-T", "1")
	err = cmd.Run()
	if err != nil {
		fmt.Println("could not stop docker-compose:", err)
	}

	os.Exit(code)
}

func TestDBStore_Integration(t *testing.T) {
	// Truncate tables before test to ensure clean state
	_, err := testStore.db.Exec("TRUNCATE orders, deliveries, payments, items RESTART IDENTITY CASCADE;")
	require.NoError(t, err)

	// Create a sample order
	order := &domain.Order{
		OrderUID:    "testuid123",
		TrackNumber: "trackno1",
		Entry:       "WBIL",
		Delivery: domain.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: domain.Payment{
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
		Items: []domain.Item{
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

	t.Run("Insert and Get", func(t *testing.T) {
		// Insert the order
		err = testStore.Insert(ctx, order)
		require.NoError(t, err)

		// Get the order back
		retrievedOrder, err := testStore.Get(ctx, order.OrderUID)
		require.NoError(t, err)
		require.NotNil(t, retrievedOrder)

		// Compare a few key fields
		require.Equal(t, order.OrderUID, retrievedOrder.OrderUID)
		require.Equal(t, order.Delivery.Name, retrievedOrder.Delivery.Name)
		require.Equal(t, order.Payment.Transaction, retrievedOrder.Payment.Transaction)
		require.Equal(t, len(order.Items), len(retrievedOrder.Items))
		require.Equal(t, order.Items[0].ChrtID, retrievedOrder.Items[0].ChrtID)
		require.True(t, order.DateCreated.Equal(retrievedOrder.DateCreated), "timestamps should be equal")
	})
}
