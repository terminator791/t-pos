package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application user (owner, admin, cashier, client)
type User struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	LicenseID        *uuid.UUID     `gorm:"type:uuid" json:"license_id"`
	Email            *string        `gorm:"size:255;uniqueIndex" json:"email"`
	EmailVerifiedAt  *time.Time     `json:"email_verified_at"`
	Username         *string        `gorm:"size:255" json:"username"`
	Name             string         `gorm:"size:255;not null" json:"name"`
	Password         string         `gorm:"size:255;not null" json:"-"` // hashed password
	Pin              *int           `json:"pin"`                         // pin code
	InfoDevice       *string        `gorm:"size:255" json:"info_device"`
	FCMToken         *string        `gorm:"size:255" json:"fcm_token"`
	RememberToken    *string        `gorm:"size:100" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	License           *License       `gorm:"foreignKey:LicenseID;constraint:OnDelete:CASCADE" json:"license,omitempty"`
	OwnedShops        []Shop         `gorm:"foreignKey:UserID" json:"owned_shops,omitempty"`
	CashierTransactions []Transaction `gorm:"foreignKey:CashierID" json:"cashier_transactions,omitempty"`
	CustomerTransactions []Transaction `gorm:"foreignKey:UserID" json:"customer_transactions,omitempty"`
	Carts             []Cart         `gorm:"foreignKey:UserID" json:"carts,omitempty"`
	Payments          []Payment      `gorm:"foreignKey:UserID" json:"payments,omitempty"`
	LicenseLogs       []LicenseLog   `gorm:"foreignKey:UserID" json:"license_logs,omitempty"`
	Logs              []Log          `gorm:"foreignKey:UserID" json:"logs,omitempty"`
	UserRoles         []UserRole     `gorm:"foreignKey:UserID" json:"user_roles,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}