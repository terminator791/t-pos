package entities

import (
	"time"
)

// Payment represents a payment made for an order
type Payment struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	OrderID         uint      `gorm:"not null" json:"order_id"`
	Amount          float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	PaymentMethod   string    `gorm:"not null" json:"payment_method"` // cash, card, digital
	ReferenceNumber string    `json:"reference_number"`
	Status          string    `gorm:"default:completed" json:"status"` // pending, completed, failed
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Order Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}