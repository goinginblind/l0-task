package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// o.Validate is a wrapper around validator.Struct(o), it checks if the fields are valid using the
// exposed structs' validation tags, i.e. o.Payments.Currency has these:
//
//	`json:"currency" validate:"required,iso4217"`
//
// Error returned contains all the checks that failed.
func (o *Order) Validate() (bool, error) {
	valid := validator.New()
	if err := valid.Struct(o); err != nil {
		return false, fmt.Errorf("order with order_uid '%s' failed validation: %w", o.OrderUID, err)
	}

	return true, nil
}
