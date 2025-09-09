package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// RoleRepositoryImpl implements RoleRepository interface
type RoleRepositoryImpl struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) *RoleRepositoryImpl {
	return &RoleRepositoryImpl{db: db}
}

// Create creates a new role
func (r *RoleRepositoryImpl) Create(ctx context.Context, role *entities.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID retrieves a role by ID
func (r *RoleRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	var role entities.Role
	err := r.db.WithContext(ctx).Preload("UserRoles").Preload("Policies").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepositoryImpl) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	var role entities.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// List retrieves a list of roles with pagination
func (r *RoleRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Role, error) {
	var roles []*entities.Role
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&roles).Error
	return roles, err
}

// Update updates an existing role
func (r *RoleRepositoryImpl) Update(ctx context.Context, role *entities.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete deletes a role (soft delete)
func (r *RoleRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Role{}, "id = ?", id).Error
}

// GetActiveRoles retrieves all active roles
func (r *RoleRepositoryImpl) GetActiveRoles(ctx context.Context) ([]*entities.Role, error) {
	var roles []*entities.Role
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&roles).Error
	return roles, err
}

// PolicyRepositoryImpl implements PolicyRepository interface
type PolicyRepositoryImpl struct {
	db *gorm.DB
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(db *gorm.DB) *PolicyRepositoryImpl {
	return &PolicyRepositoryImpl{db: db}
}

// Create creates a new policy
func (r *PolicyRepositoryImpl) Create(ctx context.Context, policy *entities.Policy) error {
	return r.db.WithContext(ctx).Create(policy).Error
}

// GetByID retrieves a policy by ID
func (r *PolicyRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Policy, error) {
	var policy entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").First(&policy, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// List retrieves a list of policies with pagination
func (r *PolicyRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Preload("Role").Find(&policies).Error
	return policies, err
}

// GetByRole retrieves policies by role ID
func (r *PolicyRepositoryImpl) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Where("role_id = ?", roleID).Preload("Role").Find(&policies).Error
	return policies, err
}

// GetByDomain retrieves policies by domain
func (r *PolicyRepositoryImpl) GetByDomain(ctx context.Context, domain string) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Where("domain = ?", domain).Preload("Role").Find(&policies).Error
	return policies, err
}

// Update updates an existing policy
func (r *PolicyRepositoryImpl) Update(ctx context.Context, policy *entities.Policy) error {
	return r.db.WithContext(ctx).Save(policy).Error
}

// Delete deletes a policy (soft delete)
func (r *PolicyRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Policy{}, "id = ?", id).Error
}

// GetActivePolicies retrieves all active policies
func (r *PolicyRepositoryImpl) GetActivePolicies(ctx context.Context) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Preload("Role").Find(&policies).Error
	return policies, err
}

// UserRoleRepositoryImpl implements UserRoleRepository interface
type UserRoleRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *gorm.DB) *UserRoleRepositoryImpl {
	return &UserRoleRepositoryImpl{db: db}
}

// Create creates a new user role
func (r *UserRoleRepositoryImpl) Create(ctx context.Context, userRole *entities.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// GetByID retrieves a user role by ID
func (r *UserRoleRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.UserRole, error) {
	var userRole entities.UserRole
	err := r.db.WithContext(ctx).Preload("User").Preload("Role").First(&userRole, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &userRole, nil
}

// GetByUserID retrieves user roles by user ID
func (r *UserRoleRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserRole, error) {
	var userRoles []*entities.UserRole
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Role").Find(&userRoles).Error
	return userRoles, err
}

// GetByUserAndDomain retrieves user roles by user ID and domain
func (r *UserRoleRepositoryImpl) GetByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) ([]*entities.UserRole, error) {
	var userRoles []*entities.UserRole
	err := r.db.WithContext(ctx).Where("user_id = ? AND domain = ?", userID, domain).Preload("Role").Find(&userRoles).Error
	return userRoles, err
}

// List retrieves a list of user roles with pagination
func (r *UserRoleRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.UserRole, error) {
	var userRoles []*entities.UserRole
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Preload("User").Preload("Role").Find(&userRoles).Error
	return userRoles, err
}

// Update updates an existing user role
func (r *UserRoleRepositoryImpl) Update(ctx context.Context, userRole *entities.UserRole) error {
	return r.db.WithContext(ctx).Save(userRole).Error
}

// Delete deletes a user role (soft delete)
func (r *UserRoleRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.UserRole{}, "id = ?", id).Error
}

// DeleteByUserAndRole deletes a user role by user ID, role ID and domain
func (r *UserRoleRepositoryImpl) DeleteByUserAndRole(ctx context.Context, userID, roleID uuid.UUID, domain string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ? AND domain = ?", userID, roleID, domain).Delete(&entities.UserRole{}).Error
}