package model

import "time"

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type ProductDetails struct {
	ID    int64   `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type Order struct {
	ID          int64       `json:"id"`
	CustomerID  int64       `json:"customerId"`
	ProductID   int64       `json:"productId"`
	ProductName string      `json:"productName"`
	Quantity    int         `json:"quantity"`
	Price       float64     `json:"price"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CreateOrderRequest struct {
	CustomerID int64 `json:"customerId" binding:"required,min=1"`
	ProductID  int64 `json:"productId" binding:"required,min=1"`
	Quantity   int   `json:"quantity" binding:"required,min=1,max=1000"`
}

type UpdateStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
