package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// LicenseLogRepositoryImpl implements LicenseLogRepository interface
type LicenseLogRepositoryImpl struct {
	db *gorm.DB
}

// NewLicenseLogRepository creates a new license log repository
func NewLicenseLogRepository(db *gorm.DB) *LicenseLogRepositoryImpl {
	return &LicenseLogRepositoryImpl{db: db}
}

// Create creates a new license log
func (r *LicenseLogRepositoryImpl) Create(ctx context.Context, licenseLog *entities.LicenseLog) error {
	return r.db.WithContext(ctx).Create(licenseLog).Error
}

// GetByID retrieves a license log by ID
func (r *LicenseLogRepositoryImpl) GetByID(ctx context.Context, id uint) (*entities.LicenseLog, error) {
	var licenseLog entities.LicenseLog
	err := r.db.WithContext(ctx).Preload("User").Preload("License").First(&licenseLog, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &licenseLog, nil
}

// GetByLicenseID retrieves license logs by license ID
func (r *LicenseLogRepositoryImpl) GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.LicenseLog, error) {
	var licenseLogs []*entities.LicenseLog
	err := r.db.WithContext(ctx).Preload("User").Preload("License").Where("license_id = ?", licenseID).Find(&licenseLogs).Error
	return licenseLogs, err
}

// GetByUserID retrieves license logs by user ID
func (r *LicenseLogRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.LicenseLog, error) {
	var licenseLogs []*entities.LicenseLog
	err := r.db.WithContext(ctx).Preload("User").Preload("License").Where("user_id = ?", userID).Find(&licenseLogs).Error
	return licenseLogs, err
}

// Delete deletes a license log (soft delete)
func (r *LicenseLogRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.LicenseLog{}, "id = ?", id).Error
}

// DeleteByLicenseID deletes all license logs for a license (soft delete)
func (r *LicenseLogRepositoryImpl) DeleteByLicenseID(ctx context.Context, licenseID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("license_id = ?", licenseID).Delete(&entities.LicenseLog{}).Error
}

// List retrieves a list of license logs with pagination
func (r *LicenseLogRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.LicenseLog, error) {
	var licenseLogs []*entities.LicenseLog
	err := r.db.WithContext(ctx).Preload("User").Preload("License").Limit(limit).Offset(offset).Find(&licenseLogs).Error
	return licenseLogs, err
}
