package models

import (
	"time"
)

// Customer represents a customer in our test database
type Customer struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Company   string    `json:"company" db:"company"`
	City      string    `json:"city" db:"city"`
	Country   string    `json:"country" db:"country"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Product represents a product in our test database
type Product struct {
	ID          int64   `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Category    string  `json:"category" db:"category"`
	Price       float64 `json:"price" db:"price"`
	StockQty    int     `json:"stock_qty" db:"stock_qty"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Order represents an order in our test database
type Order struct {
	ID         int64     `json:"id" db:"id"`
	CustomerID int64     `json:"customer_id" db:"customer_id"`
	Status     string    `json:"status" db:"status"`
	Total      float64   `json:"total" db:"total"`
	Notes      string    `json:"notes" db:"notes"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ShippedAt  *time.Time `json:"shipped_at" db:"shipped_at"`
}

// OrderItem represents an order item in our test database
type OrderItem struct {
	ID        int64   `json:"id" db:"id"`
	OrderID   int64   `json:"order_id" db:"order_id"`
	ProductID int64   `json:"product_id" db:"product_id"`
	Quantity  int     `json:"quantity" db:"quantity"`
	Price     float64 `json:"price" db:"price"`
}

// Order statuses
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Product categories
const (
	CategoryElectronics = "electronics"
	CategoryBooks       = "books"
	CategoryClothing    = "clothing"
	CategoryHome        = "home"
	CategorySports      = "sports"
	CategoryToys        = "toys"
)