package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment represents payments linked to a transaction
type Payment struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ShopID        uuid.UUID      `gorm:"type:uuid;not null" json:"shop_id"`
	UserID        *uuid.UUID     `gorm:"type:uuid" json:"user_id"`
	TransactionID uuid.UUID      `gorm:"type:uuid;not null" json:"transaction_id"`
	Status        PaymentStatus  `gorm:"default:pending" json:"status"`
	Total         float64        `gorm:"type:decimal(10,2);not null" json:"total"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Shop        Shop        `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	User        *User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Transaction Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE" json:"transaction,omitempty"`
	Receipts    []Receipt   `gorm:"foreignKey:PaymentsID" json:"receipts,omitempty"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// IsCompleted checks if the payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == PaymentStatusCompleted
}

// Syncable interface implementation
func (p Payment) GetID() uuid.UUID {
	return p.ID
}

func (p Payment) GetCreatedAt() time.Time {
	return p.CreatedAt
}

func (p Payment) GetUpdatedAt() time.Time {
	return p.UpdatedAt
}

func (p Payment) GetDeletedAt() *time.Time {
	if p.DeletedAt.Valid {
		return &p.DeletedAt.Time
	}
	return nil
}