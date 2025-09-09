package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Receipt represents a receipt record pointing to a payment
type Receipt struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID     uuid.UUID      `gorm:"type:uuid;not null" json:"shop_id"`
	PaymentsID uuid.UUID      `gorm:"type:uuid;not null" json:"payments_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop    Shop    `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Payment Payment `gorm:"foreignKey:PaymentsID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
}

// TableName specifies the table name for Receipt
func (Receipt) TableName() string {
	return "receipts"
}
