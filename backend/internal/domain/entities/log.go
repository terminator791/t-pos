package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Log represents audit trail of user actions and model changes
type Log struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      *uuid.UUID     `gorm:"type:uuid" json:"user_id"`
	Action      string         `gorm:"size:255;not null" json:"action"`
	Model       *string        `gorm:"size:255" json:"model"`
	ModelID     *uuid.UUID     `gorm:"type:uuid" json:"model_id"`
	OldValues   *string        `gorm:"type:json" json:"old_values"`
	NewValues   *string        `gorm:"type:json" json:"new_values"`
	IPAddress   *string        `gorm:"size:255" json:"ip_address"`
	UserAgent   *string        `gorm:"type:text" json:"user_agent"`
	Description *string        `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"user,omitempty"`
}

// TableName specifies the table name for Log
func (Log) TableName() string {
	return "logs"
}
