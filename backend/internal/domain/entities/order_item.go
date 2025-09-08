package entities

import (
	"time"
)

// OrderItem represents an item within an order
type OrderItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"not null" json:"order_id"`
	ProductID  uint      `gorm:"not null" json:"product_id"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	UnitPrice  float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice float64   `gorm:"type:decimal(10,2);not null" json:"total_price"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	Order   Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName specifies the table name for OrderItem
func (OrderItem) TableName() string {
	return "order_items"
}

// CalculateTotal calculates the total price for this order item
func (oi *OrderItem) CalculateTotal() {
	oi.TotalPrice = oi.UnitPrice * float64(oi.Quantity)
}