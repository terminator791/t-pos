package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product represents a sellable item
type Product struct {
	ID           uuid.UUID            `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID       uuid.UUID            `gorm:"type:uuid;not null" json:"shop_id"`
	CatID        *uuid.UUID           `gorm:"type:uuid" json:"cat_id"` // category ID
	Photo        *string              `gorm:"size:255" json:"photo"`
	Name         string               `gorm:"size:255;not null" json:"name"`
	Barcode      *string              `gorm:"size:255" json:"barcode"`
	Unit         *string              `gorm:"size:50" json:"unit"`
	PPN          *float64             `gorm:"type:decimal(5,2)" json:"ppn"`        // tax percentage
	Sale         float64              `gorm:"type:decimal(10,2);not null" json:"sale"`
	Buy          float64              `gorm:"type:decimal(10,2);not null" json:"buy"`
	Profit       *float64             `gorm:"type:decimal(10,2)" json:"profit"`
	Stock        int                  `gorm:"default:0" json:"stock"`
	IsSchedule   bool                 `gorm:"default:false" json:"is_schedule"`
	Schedule     *string              `gorm:"type:json" json:"schedule"`
	Qty          *int                 `json:"qty"`
	IsHaveStock  bool                 `gorm:"default:true" json:"is_have_stock"` // alias: has_stock
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	DeletedAt    gorm.DeletedAt       `gorm:"index" json:"-"`

	// Relationships
	Shop               Shop                 `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Category           *Category            `gorm:"foreignKey:CatID;constraint:OnDelete:SET NULL" json:"category,omitempty"`
	Carts              []Cart               `gorm:"foreignKey:ProductID" json:"carts,omitempty"`
	TransactionProducts []TransactionProduct `gorm:"foreignKey:ProductID" json:"transaction_products,omitempty"`
	StockHistories     []StockHistory       `gorm:"foreignKey:ProductID" json:"stock_histories,omitempty"`
}

// TableName specifies the table name for Product
func (Product) TableName() string {
	return "products"
}

// IsLowStock checks if the product is below minimum stock level
func (p *Product) IsLowStock() bool {
	// Since we don't have min_stock_level in the new schema, we can define a default threshold
	return p.Stock <= 10 // configurable threshold
}

// CalculateProfit calculates profit margin
func (p *Product) CalculateProfit() {
	if p.Sale > 0 && p.Buy > 0 {
		profit := p.Sale - p.Buy
		p.Profit = &profit
	}
}