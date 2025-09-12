package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Shop represents a merchant shop operating under a license
type Shop struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	LicenseID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_shops_license_updated,priority:1" json:"license_id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"` // owner_user_id
	Name            string         `gorm:"size:255;not null" json:"name"`
	Domain          string         `gorm:"size:100;not null;uniqueIndex" json:"domain"` // unique domain for multi-tenancy
	Photo           *string        `gorm:"size:255" json:"photo"`
	Address         *string        `gorm:"type:text" json:"address"`
	Slogan          *string        `gorm:"size:255" json:"slogan"`
	ProfitCalculate int64          `gorm:"default:0" json:"profit_calculate"` // toggle_or_mode
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `gorm:"index:idx_shops_license_updated,priority:2" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	License      License       `gorm:"foreignKey:LicenseID;constraint:OnDelete:CASCADE" json:"license,omitempty"`
	Owner        User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"owner,omitempty"`
	Categories   []Category    `gorm:"foreignKey:ShopID" json:"categories,omitempty"`
	Products     []Product     `gorm:"foreignKey:ShopID" json:"products,omitempty"`
	Carts        []Cart        `gorm:"foreignKey:ShopID" json:"carts,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:ShopID" json:"transactions,omitempty"`
	Payments     []Payment     `gorm:"foreignKey:ShopID" json:"payments,omitempty"`
	Receipts     []Receipt     `gorm:"foreignKey:ShopID" json:"receipts,omitempty"`
	Histories    []History     `gorm:"foreignKey:ShopID" json:"histories,omitempty"`
	Expenses     []Expense     `gorm:"foreignKey:ShopID" json:"expenses,omitempty"`
}

// TableName specifies the table name for Shop
func (Shop) TableName() string {
	return "shops"
}

// BeforeCreate hook to generate ID and Domain automatically
func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Domain == "" {
		s.Domain = "shop-" + s.ID.String()
	}
	return nil
}

// DomainName returns the shop domain as ("shop-" + uuid)
func (s *Shop) DomainName() string {
	return "shop-" + s.ID.String()
}

// Syncable interface implementation
func (s Shop) GetID() uuid.UUID {
	return s.ID
}

func (s Shop) GetCreatedAt() time.Time {
	return s.CreatedAt
}

func (s Shop) GetUpdatedAt() time.Time {
	return s.UpdatedAt
}

func (s Shop) GetDeletedAt() *time.Time {
	if s.DeletedAt.Valid {
		return &s.DeletedAt.Time
	}
	return nil
}
