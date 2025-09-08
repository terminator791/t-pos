package entities

import (
	"time"
	"gorm.io/gorm"
)

// Product represents a product in the POS system
type Product struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Description    string         `json:"description"`
	SKU            string         `gorm:"uniqueIndex;not null" json:"sku"`
	Barcode        string         `json:"barcode"`
	CategoryID     *uint          `json:"category_id"`
	Price          float64        `gorm:"type:decimal(10,2);not null" json:"price"`
	Cost           *float64       `gorm:"type:decimal(10,2)" json:"cost"`
	StockQuantity  int            `gorm:"default:0" json:"stock_quantity"`
	MinStockLevel  int            `gorm:"default:0" json:"min_stock_level"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Category   *Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	OrderItems []OrderItem `gorm:"foreignKey:ProductID" json:"order_items,omitempty"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}

// IsLowStock checks if the product is below minimum stock level
func (p *Product) IsLowStock() bool {
	return p.StockQuantity <= p.MinStockLevel
}