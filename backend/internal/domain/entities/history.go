package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// History represents a lightweight link between shop and transaction for history views
type History struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID        uuid.UUID      `gorm:"type:uuid;not null;index:idx_histories_shop_updated,priority:1" json:"shop_id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null;index:idx_histories_transaction_updated,priority:1" json:"transaction_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `gorm:"index:idx_histories_updated_at;index:idx_histories_shop_updated,priority:2;index:idx_histories_transaction_updated,priority:2" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop        Shop        `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Transaction Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"transaction,omitempty"`
}

// TableName specifies the table name for History
func (History) TableName() string {
	return "histories"
}

// Syncable interface implementation
func (h History) GetID() uuid.UUID {
	return h.ID
}

func (h History) GetCreatedAt() time.Time {
	return h.CreatedAt
}

func (h History) GetUpdatedAt() time.Time {
	return h.UpdatedAt
}

func (h History) GetDeletedAt() *time.Time {
	if h.DeletedAt.Valid {
		return &h.DeletedAt.Time
	}
	return nil
}
