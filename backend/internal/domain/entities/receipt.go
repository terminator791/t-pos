package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Receipt represents a receipt record pointing to a payment
type Receipt struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID     uuid.UUID      `gorm:"type:uuid;not null;index:idx_receipts_shop_updated,priority:1" json:"shop_id"`
	PaymentsID uuid.UUID      `gorm:"type:uuid;not null;index:idx_receipts_payment_updated,priority:1" json:"payments_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `gorm:"index:idx_receipts_updated_at;index:idx_receipts_shop_updated,priority:2;index:idx_receipts_payment_updated,priority:2" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop    Shop    `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Payment Payment `gorm:"foreignKey:PaymentsID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
}

// TableName specifies the table name for Receipt
func (Receipt) TableName() string {
	return "receipts"
}

// Syncable interface implementation
func (r Receipt) GetID() uuid.UUID {
	return r.ID
}

func (r Receipt) GetCreatedAt() time.Time {
	return r.CreatedAt
}

func (r Receipt) GetUpdatedAt() time.Time {
	return r.UpdatedAt
}

func (r Receipt) GetDeletedAt() *time.Time {
	if r.DeletedAt.Valid {
		return &r.DeletedAt.Time
	}
	return nil
}
