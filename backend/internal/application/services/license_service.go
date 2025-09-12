package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"gorm.io/gorm"
)

// LicenseService handles license business logic
type LicenseService struct {
	licenseRepo    *repositories.LicenseRepositoryImpl
	licenseLogRepo *repositories.LicenseLogRepositoryImpl
	userRepo       *repositories.UserRepositoryImpl
	db             *gorm.DB
}

// NewLicenseService creates a new license service
func NewLicenseService(
	licenseRepo *repositories.LicenseRepositoryImpl,
	licenseLogRepo *repositories.LicenseLogRepositoryImpl,
	userRepo *repositories.UserRepositoryImpl,
	db *gorm.DB,
) *LicenseService {
	return &LicenseService{
		licenseRepo:    licenseRepo,
		licenseLogRepo: licenseLogRepo,
		userRepo:       userRepo,
		db:             db,
	}
}

// CreateLicenseRequest represents a license creation request
type CreateLicenseRequest struct {
	SerialNumber string `json:"serial_number" binding:"required"`
}

// GetLicense retrieves a license by ID
func (s *LicenseService) GetLicense(ctx context.Context, id uuid.UUID) (*entities.License, error) {
	return s.licenseRepo.GetByID(ctx, id)
}

// GetAllLicenses retrieves all licenses with pagination
func (s *LicenseService) GetAllLicenses(ctx context.Context, limit, offset int) ([]*entities.License, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.licenseRepo.List(ctx, limit, offset)
}

// GetLicenseBySerialNumber retrieves a license by serial number
func (s *LicenseService) GetLicenseBySerialNumber(ctx context.Context, serialNumber string) (*entities.License, error) {
	return s.licenseRepo.GetBySerialNumber(ctx, serialNumber)
}

// CreateLicense creates a new license and license log
func (s *LicenseService) CreateLicense(ctx context.Context, req CreateLicenseRequest, userID uuid.UUID) (*entities.License, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// Check if license with serial number already exists
	existingLicense, err := s.licenseRepo.GetBySerialNumber(ctx, req.SerialNumber)
	if err == nil && existingLicense != nil {
		tx.Rollback()
		return nil, fmt.Errorf("license with serial number %s already exists", req.SerialNumber)
	}

	// Create new license
	license := &entities.License{
		SerialNumber: req.SerialNumber,
	}

	licenseRepoWithTx := repositories.NewLicenseRepository(tx)
	if err := licenseRepoWithTx.Create(ctx, license); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create license: %w", err)
	}

	// Create license log
	now := time.Now()
	licenseLog := &entities.LicenseLog{
		UserID:       &userID,
		LicenseID:    &license.ID,
		GenerateDate: &now,
	}

	licenseLogRepoWithTx := repositories.NewLicenseLogRepository(tx)
	if err := licenseLogRepoWithTx.Create(ctx, licenseLog); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create license log: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return license, nil
}

// DeleteLicense deletes a license and its associated license logs
func (s *LicenseService) DeleteLicense(ctx context.Context, id uuid.UUID) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Check if license exists
	_, err := s.licenseRepo.GetByID(ctx, id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("license not found: %w", err)
	}

	// Delete associated license logs
	licenseLogRepoWithTx := repositories.NewLicenseLogRepository(tx)
	if err := licenseLogRepoWithTx.DeleteByLicenseID(ctx, id); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete license logs: %w", err)
	}

	// Delete license
	licenseRepoWithTx := repositories.NewLicenseRepository(tx)
	if err := licenseRepoWithTx.Delete(ctx, id); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete license: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
