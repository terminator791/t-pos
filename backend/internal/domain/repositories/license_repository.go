package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// LicenseRepository defines the interface for license data access
type LicenseRepository interface {
	Create(ctx context.Context, license *entities.License) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.License, error)
	GetBySerialNumber(ctx context.Context, serialNumber string) (*entities.License, error)
	Update(ctx context.Context, license *entities.License) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.License, error)
}
