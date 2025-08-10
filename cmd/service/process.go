package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ProcessMessage converts a slice of bytes into a pointer to an Order struct.
//
// It returns an error in case of:
//   - slice of bytes not being a valid json
//   - failing to unmarshal byte slice into a struct
//   - json containing foreign unknown fields
//   - json not containing an `order_uid` field (this may change)
func ProcessMessage(b []byte) (*Order, error) {
	var raw json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields() // disallows unknown fields

	var data Order
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("fail to unmarshal into a struct: %w", err)
	}

	if strings.TrimSpace(data.OrderUID) == "" {
		return nil, fmt.Errorf("Missing the primary key: OrderUID, early reject")
	}

	return &data, nil
}
