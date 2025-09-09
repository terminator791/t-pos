package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockHistory represents append-only changes to product stock
type StockHistory struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ProductID uuid.UUID      `gorm:"type:uuid;not null" json:"product_id"`
	Stock     int            `gorm:"not null" json:"stock"`
	LastStock int            `gorm:"not null" json:"last_stock"`
	StockedAt time.Time      `gorm:"not null" json:"stocked_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
}

// TableName specifies the table name for StockHistory
func (StockHistory) TableName() string {
	return "stock_histories"
}

// GetStockChange returns the difference between current and last stock
func (sh *StockHistory) GetStockChange() int {
	return sh.Stock - sh.LastStock
}

// IsStockIncrease checks if this is a stock increase
func (sh *StockHistory) IsStockIncrease() bool {
	return sh.Stock > sh.LastStock
}

// IsStockDecrease checks if this is a stock decrease
func (sh *StockHistory) IsStockDecrease() bool {
	return sh.Stock < sh.LastStock
}
