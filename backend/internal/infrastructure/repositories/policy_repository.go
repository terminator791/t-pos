package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

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

// GetAll retrieves all policies
func (r *PolicyRepositoryImpl) GetAll(ctx context.Context) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Find(&policies).Error
	return policies, err
}

// GetByRoleID retrieves policies by role ID
func (r *PolicyRepositoryImpl) GetByRoleID(ctx context.Context, roleID uuid.UUID) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Where("role_id = ?", roleID).Find(&policies).Error
	return policies, err
}

// GetByRole retrieves policies by role ID (alias for GetByRoleID)
func (r *PolicyRepositoryImpl) GetByRole(ctx context.Context, roleID uuid.UUID) ([]*entities.Policy, error) {
	return r.GetByRoleID(ctx, roleID)
}

// GetBySubject retrieves policies by subject
func (r *PolicyRepositoryImpl) GetBySubject(ctx context.Context, subject string) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Where("subject = ?", subject).Find(&policies).Error
	return policies, err
}

// GetByDomain retrieves policies by domain
func (r *PolicyRepositoryImpl) GetByDomain(ctx context.Context, domain string) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Where("domain = ?", domain).Find(&policies).Error
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

// List retrieves a list of policies with pagination
func (r *PolicyRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Limit(limit).Offset(offset).Find(&policies).Error
	return policies, err
}

// CreateBatch creates multiple policies in a single transaction
func (r *PolicyRepositoryImpl) CreateBatch(ctx context.Context, policies []*entities.Policy) error {
	return r.db.WithContext(ctx).CreateInBatches(policies, 100).Error
}

// GetActivePolicies retrieves all active policies
func (r *PolicyRepositoryImpl) GetActivePolicies(ctx context.Context) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Preload("Role").Find(&policies).Error
	return policies, err
}

// GetByRoleAndDomain retrieves policies by role ID and domain
func (r *PolicyRepositoryImpl) GetByRoleAndDomain(ctx context.Context, roleID uuid.UUID, domain string) ([]*entities.Policy, error) {
	var policies []*entities.Policy
	err := r.db.WithContext(ctx).Preload("Role").Where("role_id = ? AND domain = ?", roleID, domain).Find(&policies).Error
	return policies, err
}
