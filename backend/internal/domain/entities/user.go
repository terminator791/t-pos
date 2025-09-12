package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents an application user (owner, admin, cashier, client)
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	LicenseID       *uuid.UUID     `gorm:"type:uuid;index:idx_users_license_updated,priority:1" json:"license_id"`
	RoleID          *uuid.UUID     `gorm:"type:uuid" json:"role_id"`
	ShopID          *uuid.UUID     `gorm:"type:uuid" json:"shop_id"` // Shop binding for cashiers
	Email           *string        `gorm:"size:255;uniqueIndex" json:"email"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at"`
	Username        *string        `gorm:"size:255;uniqueIndex;not null" json:"username"`
	Name            string         `gorm:"size:255" json:"name"`
	Password        string         `gorm:"size:255;not null" json:"-"` // hashed password
	Pin             *string        `gorm:"size:255" json:"-"`          // hashed pin code
	InfoDevice      *string        `gorm:"size:255" json:"info_device"`
	FCMToken        *string        `gorm:"size:255" json:"fcm_token"`
	RememberToken   *string        `gorm:"size:100" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `gorm:"index:idx_users_updated_at;index:idx_users_license_updated,priority:2" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	License             *License      `gorm:"foreignKey:LicenseID;constraint:OnDelete:CASCADE" json:"license,omitempty"`
	Role                *Role         `gorm:"foreignKey:RoleID;constraint:OnDelete:SET NULL" json:"role,omitempty"`
	AssignedShop        *Shop         `gorm:"foreignKey:ShopID;constraint:OnDelete:SET NULL" json:"assigned_shop,omitempty"`
	UserDomains         []UserDomain  `gorm:"foreignKey:UserID" json:"user_domains,omitempty"`
	OwnedShops          []Shop        `gorm:"foreignKey:UserID" json:"owned_shops,omitempty"`
	CashierTransactions []Transaction `gorm:"foreignKey:CashierID" json:"cashier_transactions,omitempty"`
	Carts               []Cart        `gorm:"foreignKey:UserID" json:"carts,omitempty"`
	Payments            []Payment     `gorm:"foreignKey:UserID" json:"payments,omitempty"`
	LicenseLogs         []LicenseLog  `gorm:"foreignKey:UserID" json:"license_logs,omitempty"`
	Logs                []Log         `gorm:"foreignKey:UserID" json:"logs,omitempty"`
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

// GetAccessibleShopID returns the shop ID the user can access
// For cashiers: returns their assigned shop_id
// For owner business: returns nil (can access all shops under their license)
func (u *User) GetAccessibleShopID() *uuid.UUID {
	if u.Role != nil && u.Role.Name == "cashier" && u.ShopID != nil {
		return u.ShopID
	}
	return nil // Owner business can access all shops under their license
}

// IsCashier checks if the user has cashier role
func (u *User) IsCashier() bool {
	return u.Role != nil && u.Role.Name == "cashier"
}

// IsOwnerBusiness checks if the user has owner business role
func (u *User) IsOwnerBusiness() bool {
	return u.Role != nil && u.Role.Name == "owner_business"
}

// Syncable interface implementation
func (u User) GetID() uuid.UUID {
	return u.ID
}

func (u User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

func (u User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

func (u User) GetDeletedAt() *time.Time {
	if u.DeletedAt.Valid {
		return &u.DeletedAt.Time
	}
	return nil
}
