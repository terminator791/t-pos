package entities

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a product category per shop
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"not null" json:"shop_id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop     Shop      `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Products []Product `gorm:"foreignKey:CatID" json:"products,omitempty"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}