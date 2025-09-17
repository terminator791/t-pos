package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
	TransactionStatusFailed    TransactionStatus = "failed"
)

// Transaction represents a sales transaction header
type Transaction struct {
	ID                   uuid.UUID         `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ShopID               uuid.UUID         `gorm:"type:uuid;not null;index:idx_transactions_shop_updated,priority:1" json:"shop_id"`
	CashierID            uuid.UUID         `gorm:"type:uuid;not null;index:idx_transactions_cashier_updated,priority:1" json:"cashier_id"`
	CustomerName         *string           `gorm:"size:255" json:"customer_name"`
	Discount             float64           `gorm:"type:decimal(10,2);default:0" json:"discount"`
	DiscountPercentage   float64           `gorm:"type:decimal(5,2);default:0" json:"discount_percentage"`
	AdditionalCost       float64           `gorm:"type:decimal(10,2);default:0" json:"additional_cost"`
	Status               TransactionStatus `gorm:"default:pending" json:"status"`
	TotalPrice           float64           `gorm:"type:decimal(10,2);not null" json:"total_price"`
	ProfitTransaction    *float64          `gorm:"type:decimal(10,2)" json:"profit_transaction"`
	CashierName          *string           `gorm:"size:255" json:"cashier_name"`
	Change               *float64          `gorm:"type:decimal(10,2)" json:"change"`
	Amount               int64             `gorm:"default:0" json:"amount"` // amount_paid
	InitialPaymentStatus *string           `gorm:"size:255" json:"initial_payment_status"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `gorm:"index:idx_transactions_updated_at;index:idx_transactions_shop_updated,priority:2;index:idx_transactions_cashier_updated,priority:2" json:"updated_at"`
	DeletedAt            gorm.DeletedAt    `gorm:"index" json:"-"`

	// Relationships
	Shop                Shop                 `gorm:"foreignKey:ShopID;constraint:OnDelete:CASCADE" json:"shop,omitempty"`
	Cashier             User                 `gorm:"foreignKey:CashierID;constraint:OnDelete:CASCADE" json:"cashier,omitempty"`
	TransactionProducts []TransactionProduct `gorm:"foreignKey:TransactionID" json:"transaction_products,omitempty"`
	Payments            []Payment            `gorm:"foreignKey:TransactionID" json:"payments,omitempty"`
	Histories           []History            `gorm:"foreignKey:TransactionID" json:"histories,omitempty"`
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// CalculateTotal calculates the total amount including tax and discounts
func (t *Transaction) CalculateTotal() {
	t.TotalPrice = t.TotalPrice + t.AdditionalCost - t.Discount
	if t.DiscountPercentage > 0 {
		discountAmount := t.TotalPrice * (t.DiscountPercentage / 100)
		t.TotalPrice -= discountAmount
	}
}

// IsCompleted checks if the transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// Syncable interface implementation
func (t Transaction) GetID() uuid.UUID {
	return t.ID
}

func (t Transaction) GetCreatedAt() time.Time {
	return t.CreatedAt
}

func (t Transaction) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

func (t Transaction) GetDeletedAt() *time.Time {
	if t.DeletedAt.Valid {
		return &t.DeletedAt.Time
	}
	return nil
}
