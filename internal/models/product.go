package models

import "time"

type Product struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Price     float64   `db:"price" json:"price"`
	Stock     int       `db:"stock" json:"stock"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateProductRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
	Stock int     `json:"stock" binding:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}
