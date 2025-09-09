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

// UserManagementService handles user management business logic for admin users
type UserManagementService struct {
	userRepo    *repositories.UserRepositoryImpl
	roleRepo    *repositories.RoleRepositoryImpl
	licenseRepo *repositories.LicenseRepositoryImpl
	db          *gorm.DB
}

// NewUserManagementService creates a new user management service
func NewUserManagementService(
	userRepo *repositories.UserRepositoryImpl,
	roleRepo *repositories.RoleRepositoryImpl,
	licenseRepo *repositories.LicenseRepositoryImpl,
	db *gorm.DB,
) *UserManagementService {
	return &UserManagementService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		licenseRepo: licenseRepo,
		db:          db,
	}
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username     string `json:"username" binding:"required"`
	SerialNumber string `json:"serial_number" binding:"required"`
	RoleID       string `json:"role_id" binding:"required"`
	Pin          string `json:"pin" binding:"required"`
}

// UpdateUserPasswordRequest represents a user password update request
type UpdateUserPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

// GetUser retrieves a user by ID (only admin and super_admin roles)
func (s *UserManagementService) GetUser(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if user is an admin user (admin or super_admin)
	if !s.isAdminRole(ctx, user.RoleID) {
		return nil, fmt.Errorf("user is not an admin user")
	}

	return user, nil
}

// GetAllUsers retrieves all admin users (users with admin and super_admin roles)
func (s *UserManagementService) GetAllUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get admin and super_admin role IDs
	adminRole, err := s.roleRepo.GetByName(ctx, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to get admin role: %w", err)
	}

	superAdminRole, err := s.roleRepo.GetByName(ctx, "super_admin")
	if err != nil {
		return nil, fmt.Errorf("failed to get super_admin role: %w", err)
	}

	// Query users with admin roles
	var users []*entities.User
	err = s.db.WithContext(ctx).
		Preload("Role").
		Preload("License").
		Where("role_id IN ?", []uuid.UUID{adminRole.ID, superAdminRole.ID}).
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, err
}

// CreateUser creates a new admin user (user with admin or super_admin role)
func (s *UserManagementService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
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

	// Check if role is valid for admin users (admin or super_admin)
	roleRepoWithTx := repositories.NewRoleRepository(tx)
	role, err := roleRepoWithTx.GetByID(ctx, roleID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("role not found: %w", err)
	}

	if role.Name != "admin" && role.Name != "super_admin" {
		tx.Rollback()
		return nil, fmt.Errorf("invalid role for user: %s. Only 'admin' and 'super_admin' are allowed", role.Name)
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
		Name:      req.Username, // Use username as default name
		Password:  string(hashedPin), // Use PIN as password for admin users
		Pin:       stringPtr(string(hashedPin)),
	}

	if err := userRepoWithTx.Create(ctx, user); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload user with relationships
	user, err = s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload created user: %w", err)
	}

	return user, nil
}

// UpdateUserPassword updates a user's password
func (s *UserManagementService) UpdateUserPassword(ctx context.Context, id uuid.UUID, req UpdateUserPasswordRequest) error {
	// Check if user exists and is an admin user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user is an admin user (admin or super_admin)
	if !s.isAdminRole(ctx, user.RoleID) {
		return fmt.Errorf("user is not an admin user")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.Password = string(hashedPassword)
	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes an admin user (soft delete)
func (s *UserManagementService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Check if user exists and is an admin user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user is an admin user (admin or super_admin)
	if !s.isAdminRole(ctx, user.RoleID) {
		return fmt.Errorf("user is not an admin user")
	}

	return s.userRepo.Delete(ctx, id)
}

// isAdminRole checks if a role ID belongs to an admin role (admin or super_admin)
func (s *UserManagementService) isAdminRole(ctx context.Context, roleID *uuid.UUID) bool {
	if roleID == nil {
		return false
	}

	role, err := s.roleRepo.GetByID(ctx, *roleID)
	if err != nil {
		return false
	}

	return role.Name == "admin" || role.Name == "super_admin"
}