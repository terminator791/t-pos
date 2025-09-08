package entities

import (
	"time"
	"gorm.io/gorm"
)

// Customer represents a customer in the POS system
type Customer struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Email         string         `json:"email"`
	Phone         string         `json:"phone"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Address       string         `json:"address"`
	LoyaltyPoints int            `gorm:"default:0" json:"loyalty_points"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Orders []Order `gorm:"foreignKey:CustomerID" json:"orders,omitempty"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// FullName returns the full name of the customer
func (c *Customer) FullName() string {
	if c.FirstName == "" && c.LastName == "" {
		return "Guest Customer"
	}
	return c.FirstName + " " + c.LastName
}