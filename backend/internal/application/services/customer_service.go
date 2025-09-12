package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/infrastructure/repositories"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CustomerService handles customer business logic
type CustomerService struct {
	userRepo    *repositories.UserRepositoryImpl
	roleRepo    *repositories.RoleRepositoryImpl
	licenseRepo *repositories.LicenseRepositoryImpl
	db          *gorm.DB
}

// NewCustomerService creates a new customer service
func NewCustomerService(
	userRepo *repositories.UserRepositoryImpl,
	roleRepo *repositories.RoleRepositoryImpl,
	licenseRepo *repositories.LicenseRepositoryImpl,
	db *gorm.DB,
) *CustomerService {
	return &CustomerService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		licenseRepo: licenseRepo,
		db:          db,
	}
}

// CreateCustomerRequest represents a customer creation request
type CreateCustomerRequest struct {
	Username     string `json:"username" binding:"required"`
	SerialNumber string `json:"serial_number" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
	Pin          string `json:"pin" binding:"required"`
}

// GetCustomer retrieves a customer by ID (only cashier and owner_business roles)
func (s *CustomerService) GetCustomer(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user is a customer (cashier or owner_business)
	if !s.isCustomerRole(ctx, user.RoleID) {
		return nil, fmt.Errorf("user is not a customer")
	}

	return user, nil
}

// GetAllCustomers retrieves all customers (users with cashier and owner_business roles)
func (s *CustomerService) GetAllCustomers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get cashier and owner_business role IDs
	cashierRole, err := s.roleRepo.GetByName(ctx, "cashier")
	if err != nil {
		return nil, fmt.Errorf("failed to get cashier role: %w", err)
	}

	ownerBusinessRole, err := s.roleRepo.GetByName(ctx, "owner_business")
	if err != nil {
		return nil, fmt.Errorf("failed to get owner_business role: %w", err)
	}

	// Query users with customer roles
	var customers []*entities.User
	err = s.db.WithContext(ctx).
		Preload("Role").
		Preload("License").
		Where("role_id IN ?", []uuid.UUID{cashierRole.ID, ownerBusinessRole.ID}).
		Limit(limit).
		Offset(offset).
		Find(&customers).Error

	return customers, err
}

// CreateCustomer creates a new customer (user with cashier or owner_business role)
func (s *CustomerService) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*entities.User, error) {
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

	// Validate role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// Check if role is valid for customers (cashier or owner_business)
	roleRepoWithTx := repositories.NewRoleRepository(tx)
	role, err := roleRepoWithTx.GetByID(ctx, roleID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("role not found: %w", err)
	}

	if role.Name != "cashier" && role.Name != "owner_business" {
		tx.Rollback()
		return nil, fmt.Errorf("invalid role for customer: %s. Only 'cashier' and 'owner_business' are allowed", role.Name)
	}

	// Check if license exists
	licenseRepoWithTx := repositories.NewLicenseRepository(tx)
	license, err := licenseRepoWithTx.GetBySerialNumber(ctx, req.SerialNumber)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("license not found: %w", err)
	}

	// Check if username already exists
	userRepoWithTx := repositories.NewUserRepository(tx)
	existingUser, err := userRepoWithTx.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		tx.Rollback()
		return nil, fmt.Errorf("username %s already exists", req.Username)
	}

	// Hash PIN
	hashedPin, err := bcrypt.GenerateFromPassword([]byte(req.Pin), bcrypt.DefaultCost)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	// Create user
	user := &entities.User{
		LicenseID: &license.ID,
		RoleID:    &roleID,
		Username:  &req.Username,
		Name:      req.Username,      // Use username as default name
		Password:  string(hashedPin), // Use PIN as password for customers
		Pin:       stringPtr(string(hashedPin)),
	}

	if err := userRepoWithTx.Create(ctx, user); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload user with relationships
	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload created customer: %w", err)
	}

	return user, nil
}

// DeleteCustomer deletes a customer (soft delete)
func (s *CustomerService) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	// Check if user exists and is a customer
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("customer not found: %w", err)
	}

	// Check if user is a customer (cashier or owner_business)
	if !s.isCustomerRole(ctx, user.RoleID) {
		return fmt.Errorf("user is not a customer")
	}

	return s.userRepo.Delete(ctx, id)
}

// isCustomerRole checks if a role ID belongs to a customer role (cashier or owner_business)
func (s *CustomerService) isCustomerRole(ctx context.Context, roleID *uuid.UUID) bool {
	if roleID == nil {
		return false
	}

	role, err := s.roleRepo.GetByID(ctx, *roleID)
	if err != nil {
		return false
	}

	return role.Name == "cashier" || role.Name == "owner_business"
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
