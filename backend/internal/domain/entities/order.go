package entities

import (
	"time"
	"gorm.io/gorm"
)

// Order represents a transaction/order in the POS system
type Order struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OrderNumber    string         `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID         uint           `gorm:"not null" json:"user_id"`
	CustomerID     *uint          `json:"customer_id"`
	Subtotal       float64        `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount      float64        `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	DiscountAmount float64        `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TotalAmount    float64        `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status         string         `gorm:"default:completed" json:"status"` // pending, completed, cancelled, refunded
	PaymentMethod  string         `gorm:"not null" json:"payment_method"`  // cash, card, digital
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User      User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Customer  *Customer   `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	Payments  []Payment   `gorm:"foreignKey:OrderID" json:"payments,omitempty"`
}

// TableName specifies the table name for Order
func (Order) TableName() string {
	return "orders"
}

// CalculateTotal calculates the total amount including tax and discounts
func (o *Order) CalculateTotal() {
	o.TotalAmount = o.Subtotal + o.TaxAmount - o.DiscountAmount
}