package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LicenseLog represents a log of license generation/assignment per user
type LicenseLog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       *uuid.UUID     `gorm:"type:uuid" json:"user_id"`
	LicenseID    *uuid.UUID     `gorm:"type:uuid" json:"license_id"`
	GenerateDate *time.Time     `json:"generate_date"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    *User    `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"user,omitempty"`
	License *License `gorm:"foreignKey:LicenseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"license,omitempty"`
}

// TableName specifies the table name for LicenseLog
func (LicenseLog) TableName() string {
	return "license_logs"
}
