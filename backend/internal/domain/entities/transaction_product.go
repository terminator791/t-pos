package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionProduct represents line items per transaction
type TransactionProduct struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null;index:idx_transaction_products_transaction_id" json:"transaction_id"`
	ProductID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_transaction_products_product_updated,priority:1" json:"product_id"`
	Quantity      int            `gorm:"not null" json:"quantity"`
	UnitPrice     float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice    float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `gorm:"index:idx_transaction_products_updated_at;index:idx_transaction_products_product_updated,priority:2" json:"updated_at"`
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

// Syncable interface implementation
func (tp TransactionProduct) GetID() uuid.UUID {
	return tp.ID
}

func (tp TransactionProduct) GetCreatedAt() time.Time {
	return tp.CreatedAt
}

func (tp TransactionProduct) GetUpdatedAt() time.Time {
	return tp.UpdatedAt
}

func (tp TransactionProduct) GetDeletedAt() *time.Time {
	if tp.DeletedAt.Valid {
		return &tp.DeletedAt.Time
	}
	return nil
}
