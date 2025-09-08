package entities

import (
	"time"
	"gorm.io/gorm"
)

// Category represents a product category
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}