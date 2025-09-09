package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string         `gorm:"size:100;not null;uniqueIndex" json:"name"`
	DisplayName string         `gorm:"size:255;not null" json:"display_name"`
	Description *string        `gorm:"size:500" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	UserRoles []UserRole `gorm:"foreignKey:RoleID" json:"user_roles,omitempty"`
	Policies  []Policy   `gorm:"foreignKey:RoleID" json:"policies,omitempty"`
}

// UserRole represents the many-to-many relationship between users and roles with domain support
type UserRole struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	RoleID    uuid.UUID      `gorm:"type:uuid;not null" json:"role_id"`
	Domain    string         `gorm:"size:255;not null;default:'*'" json:"domain"` // tenant/shop domain
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Role *Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
}

// Policy represents a Casbin policy entry
type Policy struct {
	ID       uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	RoleID   *uuid.UUID     `gorm:"type:uuid" json:"role_id"`
	Subject  string         `gorm:"size:255;not null" json:"subject"` // role name or user id
	Domain   string         `gorm:"size:255;not null" json:"domain"`  // tenant/shop domain
	Object   string         `gorm:"size:255;not null" json:"object"`  // resource/endpoint
	Action   string         `gorm:"size:100;not null" json:"action"`  // HTTP method or action
	Effect   string         `gorm:"size:10;not null;default:'allow'" json:"effect"`
	IsActive bool           `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Role *Role `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE" json:"role,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// TableName specifies the table name for Policy
func (Policy) TableName() string {
	return "policies"
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (ur *UserRole) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == uuid.Nil {
		ur.ID = uuid.New()
	}
	return nil
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (p *Policy) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}