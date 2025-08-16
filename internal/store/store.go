package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/goinginblind/l0-task/internal/domain"
	"github.com/goinginblind/l0-task/internal/pkg/logger"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBStore is a database implementation of the OrderStore interface
type DBStore struct {
	db     *sql.DB
	logger logger.Logger
}

// NewDBStore creates a new DBStore
func NewDBStore(db *sql.DB, logger logger.Logger) *DBStore {
	return &DBStore{
		db:     db,
		logger: logger,
	}
}

// Insert adds a new order to the database. It's atomic, so if
// any of the inserts fail, the whole transaction is rolled back.
func (s *DBStore) Insert(ctx context.Context, o *domain.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		if isConnectionError(err) {
			return ErrConnectionFailed
		}
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
		// A duplicate insert check
		if isConnectionError(err) {
			return ErrConnectionFailed
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("%w: uid=%s", ErrAlreadyExists, o.OrderUID)
		}
		return fmt.Errorf("inserting order: %w", err)
	}

	_, err = tx.ExecContext(
		ctx, qInsertDeliveries,
		orderID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City,
		o.Delivery.Address, o.Delivery.Region, o.Delivery.Email,
	)
	if err != nil {
		if isConnectionError(err) {
			return ErrConnectionFailed
		}
		return fmt.Errorf("inserting delivery: %w", err)
	}

	_, err = tx.ExecContext(
		ctx, qInsertPayments,
		orderID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency, o.Payment.Provider,
		o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost,
		o.Payment.GoodsTotal, o.Payment.CustomFee,
	)
	if err != nil {
		if isConnectionError(err) {
			return ErrConnectionFailed
		}
		return fmt.Errorf("inserting payment: %w", err)
	}

	for _, item := range o.Items {
		_, err = tx.ExecContext(
			ctx, qInsertItems,
			orderID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			if isConnectionError(err) {
				return ErrConnectionFailed
			}
			return fmt.Errorf("inserting item: %w", err)
		}
	}

	return tx.Commit()
}

// Get retrieves a single order from the database by scanning the raw columns.
// It handles the one-to-many relationship between an order and its items.
func (s *DBStore) Get(ctx context.Context, orderUID string) (*domain.Order, error) {
	rows, err := s.db.QueryContext(ctx, qRetrieveOrder, orderUID)
	if err != nil {
		if isConnectionError(err) {
			return nil, ErrConnectionFailed
		}
		return nil, fmt.Errorf("querying for order %s: %w", orderUID, err)
	}
	defer rows.Close()

	var order *domain.Order
	for rows.Next() {
		// For the first row, we initialize the order and scan all its details.
		if order == nil {
			order = &domain.Order{}
		}
		var item domain.Item
		err := rows.Scan(
			// Order fields
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID,
			&order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
			// Delivery fields
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			// Payment fields
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
			&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
			// Item fields
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning order %s: %w", orderUID, err)
		}
		order.Items = append(order.Items, item)
	}

	if err = rows.Err(); err != nil {
		if isConnectionError(err) {
			return nil, ErrConnectionFailed
		}
		return nil, fmt.Errorf("iterating order rows for %s: %w", orderUID, err)
	}

	// If order is still nil, it means no rows were returned.
	if order == nil {
		return nil, ErrNotFound
	}

	return order, nil
}
