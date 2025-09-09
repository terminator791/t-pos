package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order represents a transaction/order in the POS system
type Order struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OrderNumber    string         `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID         uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
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

// Customer represents a customer in the POS system
type Customer struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Email     *string        `gorm:"size:255;uniqueIndex" json:"email"`
	Phone     *string        `gorm:"size:20" json:"phone"`
	Address   *string        `gorm:"size:500" json:"address"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Orders []Order `gorm:"foreignKey:CustomerID" json:"orders,omitempty"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OrderID   uint           `gorm:"not null" json:"order_id"`
	ProductID uint           `gorm:"not null" json:"product_id"`
	Quantity  int            `gorm:"not null" json:"quantity"`
	UnitPrice float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice float64       `gorm:"type:decimal(10,2);not null" json:"total_price"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Order   Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

// TableName specifies the table name for OrderItem
func (OrderItem) TableName() string {
	return "order_items"
}

// CalculateTotal calculates the total price based on quantity and unit price
func (oi *OrderItem) CalculateTotal() {
	oi.TotalPrice = float64(oi.Quantity) * oi.UnitPrice
}