package store

import "errors"

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
