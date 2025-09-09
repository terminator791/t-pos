package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// History represents a lightweight link between shop and transaction for history views
type History struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID        uuid.UUID      `gorm:"type:uuid;not null" json:"shop_id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null" json:"transaction_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop        Shop        `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Transaction Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"transaction,omitempty"`
}

// TableName specifies the table name for History
func (History) TableName() string {
	return "histories"
}
