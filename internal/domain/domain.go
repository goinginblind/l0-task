package domain

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required,alphanum"`
	TrackNumber       string    `json:"track_number" validate:"required,alphanum"`
	Entry             string    `json:"entry" validate:"required"`
	Delivery          Delivery  `json:"delivery" validate:"required"`
	Payment           Payment   `json:"payment" validate:"required"`
	Items             []Item    `json:"items" validate:"required,min=1,dive"`
	Locale            string    `json:"locale" validate:"required,bcp47_language_tag"`
	InternalSignature string    `json:"internal_signature" validate:"omitempty"`
	CustomerID        string    `json:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" validate:"required"`
	ShardKey          string    `json:"shardkey" validate:"required,numeric"`
	SmID              int       `json:"sm_id" validate:"required,gt=0"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"required,numeric"`
}

type Delivery struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required,e164"`
	Zip     string `json:"zip" validate:"required,numeric"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required,alphanum"`
	RequestID    string `json:"request_id" validate:"omitempty"`
	Currency     string `json:"currency" validate:"required,iso4217"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"required,gte=0"`
	PaymentDt    int64  `json:"payment_dt" validate:"required,gt=0"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"required,gt=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,gt=0"`
	TrackNumber string `json:"track_number" validate:"required,alphanum"`
	Price       int    `json:"price" validate:"required,gte=0"`
	Rid         string `json:"rid" validate:"required,alphanum"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size" validate:"required"`
	TotalPrice  int    `json:"total_price" validate:"gte=0"`
	NmID        int    `json:"nm_id" validate:"required,gt=0"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}

// Validate checks if the fields are valid using the
// exposed structs' validation tags.
func (o *Order) Validate() (bool, error) {
	valid := validator.New()
	if err := valid.Struct(o); err != nil {
		return false, fmt.Errorf("order with order_uid '%s' failed validation: %w", o.OrderUID, err)
	}

	return true, nil
}
