package main

import (
	"context"
	"database/sql"
	"fmt"
)

// OrdersModel is a wrapper around sql.DB connection pool,
// made to not spread sql pool around the codebase and to implement
// methods on.
type OrdersModel struct {
	DB *sql.DB
}

// Insert adds a new order to the database. It's atomic, so if
// any of the inserts fail, the whole transaction is rolled back.
func (m *OrdersModel) Insert(ctx context.Context, o *Order) error {
	tx, err := m.DB.BeginTx(ctx, nil)
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

// GetJson retrieves a single order from the database as a JSON object.
func (m *OrdersModel) GetJson(ctx context.Context, orderUID string) ([]byte, error) {
	var json []byte
	err := m.DB.QueryRowContext(ctx, qRetrieveJSON, orderUID).Scan(&json)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not an error, just no such order exists
		}
		return nil, fmt.Errorf("getting order json: %w", err)
	}
	return json, nil
}

// Exists checks if an order with the given order_uid exists in the database.
func (m *OrdersModel) Exists(ctx context.Context, orderUID string) (bool, error) {
	var exists bool
	err := m.DB.QueryRowContext(ctx, qExists, orderUID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking if order exists: %w", err)
	}
	return exists, nil
}
