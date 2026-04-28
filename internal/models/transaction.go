package models

import "time"

type Transaction struct {
	ID          int               `db:"id" json:"id"`
	TotalAmount float64           `db:"total_amount" json:"total_amount"`
	CreatedAt   time.Time         `db:"created_at" json:"created_at"`
	Items       []TransactionItem `json:"items,omitempty"`
}

type TransactionItem struct {
	ID            int     `db:"id" json:"id"`
	TransactionID int     `db:"transaction_id" json:"transaction_id"`
	ProductID     int     `db:"product_id" json:"product_id"`
	ProductName   string  `db:"product_name" json:"product_name"`
	UnitPrice     float64 `db:"unit_price" json:"unit_price"`
	Quantity      int     `db:"quantity" json:"quantity"`
	Subtotal      float64 `db:"subtotal" json:"subtotal"`
}

type CheckoutRequest struct {
	Items []CheckoutItem `json:"items" binding:"required,min=1"`
}

type CheckoutItem struct {
	ProductID int `json:"product_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}
