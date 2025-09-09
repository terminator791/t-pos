package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// License represents a license that gates shop ownership/usage
type License struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	SerialNumber string         `gorm:"size:50;not null;uniqueIndex" json:"serial_number"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Users       []User       `gorm:"foreignKey:LicenseID" json:"users,omitempty"`
	Shops       []Shop       `gorm:"foreignKey:LicenseID" json:"shops,omitempty"`
	LicenseLogs []LicenseLog `gorm:"foreignKey:LicenseID" json:"license_logs,omitempty"`
}

// TableName specifies the table name for License
func (License) TableName() string {
	return "licenses"
}

// BeforeCreate sets the ID field to a new UUID if it's not already set
func (l *License) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}
