package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"gorm.io/gorm"
)

// userDomainRepository implements UserDomainRepository interface
type userDomainRepository struct {
	db *gorm.DB
}

// NewUserDomainRepository creates a new user domain repository
func NewUserDomainRepository(db *gorm.DB) repositories.UserDomainRepository {
	return &userDomainRepository{db: db}
}

// Create creates a new user domain
func (r *userDomainRepository) Create(ctx context.Context, userDomain *entities.UserDomain) error {
	return r.db.WithContext(ctx).Create(userDomain).Error
}

// GetByUserID gets all domains for a user
func (r *userDomainRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserDomain, error) {
	var userDomains []*entities.UserDomain
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userDomains).Error
	return userDomains, err
}

// GetByUserAndDomain gets user domain by user ID and domain
func (r *userDomainRepository) GetByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) (*entities.UserDomain, error) {
	var userDomain entities.UserDomain
	err := r.db.WithContext(ctx).Where("user_id = ? AND domain = ?", userID, domain).First(&userDomain).Error
	return &userDomain, err
}

// GetByDomain gets all users for a domain
func (r *userDomainRepository) GetByDomain(ctx context.Context, domain string) ([]*entities.UserDomain, error) {
	var userDomains []*entities.UserDomain
	err := r.db.WithContext(ctx).Where("domain = ?", domain).Find(&userDomains).Error
	return userDomains, err
}

// Update updates a user domain
func (r *userDomainRepository) Update(ctx context.Context, userDomain *entities.UserDomain) error {
	return r.db.WithContext(ctx).Save(userDomain).Error
}

// Delete deletes a user domain by ID
func (r *userDomainRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.UserDomain{}, id).Error
}

// DeleteByUserAndDomain deletes user domain by user ID and domain
func (r *userDomainRepository) DeleteByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND domain = ?", userID, domain).Delete(&entities.UserDomain{}).Error
}

// List gets user domains with pagination
func (r *userDomainRepository) List(ctx context.Context, limit, offset int) ([]*entities.UserDomain, error) {
	var userDomains []*entities.UserDomain
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&userDomains).Error
	return userDomains, err
}
