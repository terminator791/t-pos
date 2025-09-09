package entities

import (
	"time"

	"gorm.io/gorm"
)

// Receipt represents a receipt record pointing to a payment
type Receipt struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	ShopID     uint           `gorm:"not null" json:"shop_id"`
	PaymentsID uint           `gorm:"not null" json:"payments_id"`
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
