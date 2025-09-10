package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cart represents a user shopping cart (pre-transaction basket)
type Cart struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID    uuid.UUID      `gorm:"type:uuid;not null" json:"shop_id"`
	ProductID uuid.UUID      `gorm:"type:uuid;not null" json:"product_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Quantity  int            `gorm:"not null;default:1" json:"quantity"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop    Shop    `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Product Product `gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE" json:"product,omitempty"`
	User    User    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for Cart
func (Cart) TableName() string {
	return "carts"
}

// Syncable interface implementation
func (c Cart) GetID() uuid.UUID {
	return c.ID
}

func (c Cart) GetCreatedAt() time.Time {
	return c.CreatedAt
}

func (c Cart) GetUpdatedAt() time.Time {
	return c.UpdatedAt
}

func (c Cart) GetDeletedAt() *time.Time {
	if c.DeletedAt.Valid {
		return &c.DeletedAt.Time
	}
	return nil
}
