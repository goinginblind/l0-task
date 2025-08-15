package store

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// These sentinel errors exist for full model encapsulation, so the upper layers aren't concerned with the
// underlying datastore or reliant on storage specific errors (like sql.ErrNoRows).

var (
	// ErrNotFound to be returned when no matching record has been found in the database.
	// It's value is "no such record exists".
	ErrNotFound = errors.New("no such record exists")
	// ErrAlreadyExists is to be returned when an order with such 'order_uid'
	// field already exists. It's value is 'record already exists'
	ErrAlreadyExists = errors.New("record already exists")

	// ErrConnectionFailed will later be used for retry/backoff logic (I think)
	ErrConnectionFailed = errors.New("connection to the database failed")
)

// isConnectionError return true if error was:
//   - detected via drive, so sql conn err
//   - one of the timeout errors
//   - if cancelation via context happened
//   - if a sys call cancels
func isConnectionError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.SQLClientUnableToEstablishSQLConnection,
			pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
			pgerrcode.TransactionResolutionUnknown,
			pgerrcode.ProtocolViolation:
			return true

		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for context cancellation or deadline exceeded
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for specific syscall errors that indicate a broken pipe or reset connection
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		if sysErr.Err == syscall.EPIPE || sysErr.Err == syscall.ECONNRESET {
			return true
		}
	}

	return false
}
