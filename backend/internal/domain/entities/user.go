package entities

import (
	"time"
	"gorm.io/gorm"
)

// User represents a system user (cashier, manager, admin)
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"not null" json:"-"`
	FirstName    string         `gorm:"not null" json:"first_name"`
	LastName     string         `gorm:"not null" json:"last_name"`
	Role         string         `gorm:"default:cashier" json:"role"` // cashier, manager, admin
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Orders []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}