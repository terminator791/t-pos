package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cart represents a user shopping cart (pre-transaction basket)
type Cart struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"not null" json:"shop_id"`
	ProductID uint           `gorm:"not null" json:"product_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
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
