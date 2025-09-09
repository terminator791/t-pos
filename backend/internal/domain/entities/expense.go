package entities

import (
	"time"

	"gorm.io/gorm"
)

// ExpenseStatus represents the status of an expense
type ExpenseStatus string

const (
	ExpenseStatusPending   ExpenseStatus = "pending"
	ExpenseStatusCompleted ExpenseStatus = "completed"
	ExpenseStatusFailed    ExpenseStatus = "failed"
	ExpenseStatusCancelled ExpenseStatus = "cancelled"
)

// Expense represents shop expenses/outflows
type Expense struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ShopID    uint           `gorm:"not null" json:"shop_id"`
	Nominal   float64        `gorm:"type:decimal(10,2);not null" json:"nominal"`
	Status    ExpenseStatus  `gorm:"default:pending" json:"status"`
	Date      time.Time      `gorm:"type:date;not null" json:"date"`
	Label     *string        `gorm:"size:255" json:"label"`
	Desc      *string        `gorm:"type:text" json:"desc"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop Shop `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
}

// TableName specifies the table name for Expense
func (Expense) TableName() string {
	return "expenses"
}

// IsCompleted checks if the expense is completed
func (e *Expense) IsCompleted() bool {
	return e.Status == ExpenseStatusCompleted
}
