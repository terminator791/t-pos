package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// LicenseRepositoryImpl implements LicenseRepository interface
type LicenseRepositoryImpl struct {
	db *gorm.DB
}

// NewLicenseRepository creates a new license repository
func NewLicenseRepository(db *gorm.DB) *LicenseRepositoryImpl {
	return &LicenseRepositoryImpl{db: db}
}

// Create creates a new license
func (r *LicenseRepositoryImpl) Create(ctx context.Context, license *entities.License) error {
	return r.db.WithContext(ctx).Create(license).Error
}

// GetByID retrieves a license by ID
func (r *LicenseRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.License, error) {
	var license entities.License
	err := r.db.WithContext(ctx).First(&license, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &license, nil
}

// GetBySerialNumber retrieves a license by serial number
func (r *LicenseRepositoryImpl) GetBySerialNumber(ctx context.Context, serialNumber string) (*entities.License, error) {
	var license entities.License
	err := r.db.WithContext(ctx).Where("serial_number = ?", serialNumber).First(&license).Error
	if err != nil {
		return nil, err
	}
	return &license, nil
}

// Update updates an existing license
func (r *LicenseRepositoryImpl) Update(ctx context.Context, license *entities.License) error {
	return r.db.WithContext(ctx).Save(license).Error
}

// Delete deletes a license (soft delete)
func (r *LicenseRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.License{}, "id = ?", id).Error
}

// List retrieves a list of licenses with pagination
func (r *LicenseRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.License, error) {
	var licenses []*entities.License
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&licenses).Error
	return licenses, err
}
