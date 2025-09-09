package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionProduct represents line items per transaction
type TransactionProduct struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null" json:"transaction_id"`
	ProductID     uint           `gorm:"not null" json:"product_id"`
	Quantity      int            `gorm:"not null" json:"quantity"`
	UnitPrice     float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice    float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"transaction,omitempty"`
	Product     Product     `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
}

// TableName specifies the table name for TransactionProduct
func (TransactionProduct) TableName() string {
	return "transaction_products"
}

// CalculateTotalPrice calculates the total price based on quantity and unit price
func (tp *TransactionProduct) CalculateTotalPrice() {
	tp.TotalPrice = float64(tp.Quantity) * tp.UnitPrice
}
