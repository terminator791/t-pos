package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserDomain represents the domains/shops a user can access
type UserDomain struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Domain    string         `gorm:"size:100;not null" json:"domain"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for UserDomain
func (UserDomain) TableName() string {
	return "user_domains"
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (ud *UserDomain) BeforeCreate(tx *gorm.DB) error {
	if ud.ID == uuid.Nil {
		ud.ID = uuid.New()
	}
	return nil
}
