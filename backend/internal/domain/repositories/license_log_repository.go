package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// LicenseLogRepository defines the interface for license log data access
type LicenseLogRepository interface {
	Create(ctx context.Context, licenseLog *entities.LicenseLog) error
	GetByID(ctx context.Context, id uint) (*entities.LicenseLog, error)
	GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.LicenseLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.LicenseLog, error)
	Delete(ctx context.Context, id uint) error
	DeleteByLicenseID(ctx context.Context, licenseID uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.LicenseLog, error)
}