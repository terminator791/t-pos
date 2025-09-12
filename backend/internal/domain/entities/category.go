package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a product category per shop
type Category struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID    uuid.UUID      `gorm:"type:uuid;not null;index:idx_categories_shop_updated,priority:1" json:"shop_id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `gorm:"index:idx_categories_updated_at;index:idx_categories_shop_updated,priority:2" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop     Shop      `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Products []Product `gorm:"foreignKey:CatID" json:"products,omitempty"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}

// Syncable interface implementation
func (c Category) GetID() uuid.UUID {
	return c.ID
}

func (c Category) GetCreatedAt() time.Time {
	return c.CreatedAt
}

func (c Category) GetUpdatedAt() time.Time {
	return c.UpdatedAt
}

func (c Category) GetDeletedAt() *time.Time {
	if c.DeletedAt.Valid {
		return &c.DeletedAt.Time
	}
	return nil
}
