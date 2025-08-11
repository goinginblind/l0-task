package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/goinginblind/l0-task/internal/domain"

	_ "github.com/lib/pq"
)

// DBStore is a database implementation of the OrderStore interface
type DBStore struct {
	db *sql.DB
}

// NewDBStore creates a new DBStore
func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

// Insert adds a new order to the database. It's atomic, so if
// any of the inserts fail, the whole transaction is rolled back.
func (s *DBStore) Insert(ctx context.Context, o *domain.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var orderID int64
	err = tx.QueryRowContext(
		ctx, qInsertOrders,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID,
		o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard,
	).Scan(&orderID)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}

	_, err = tx.ExecContext(
		ctx, qInsertDeliveries,
		orderID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City,
		o.Delivery.Address, o.Delivery.Region, o.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("inserting delivery: %w", err)
	}

	_, err = tx.ExecContext(
		ctx, qInsertPayments,
		orderID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider,
		o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost,
		o.Payment.GoodsTotal, o.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("inserting payment: %w", err)
	}

	for _, item := range o.Items {
		_, err = tx.ExecContext(
			ctx, qInsertItems,
			orderID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("inserting item: %w", err)
		}
	}

	return tx.Commit()
}

// Get retrieves a single order from the database as a JSON object
func (s *DBStore) Get(ctx context.Context, orderUID string) (*domain.Order, error) {
	var jsonBytes []byte
	err := s.db.QueryRowContext(ctx, qRetrieveJSON, orderUID).Scan(&jsonBytes)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not an error, just no such order exists
		}
		return nil, fmt.Errorf("getting order json: %w", err)
	}

	var order domain.Order
	if err := json.Unmarshal(jsonBytes, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return &order, nil
}
