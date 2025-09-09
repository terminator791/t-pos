package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// UserRepositoryImpl implements UserRepository interface
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// CreatePin creates or updates a PIN for a user by ID
func (r *UserRepositoryImpl) CreatePin(ctx context.Context, id uuid.UUID, pin string) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("pin", pin).Error
}

// UpdatePin updates the PIN for a user by ID
func (r *UserRepositoryImpl) UpdatePin(ctx context.Context, id uuid.UUID, pin string) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("pin", pin).Error
}

// DeletePin deletes the PIN for a user by ID (sets to NULL)
func (r *UserRepositoryImpl) DeletePin(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", id).Update("pin", nil).Error
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByLicenseID retrieves users by license ID
func (r *UserRepositoryImpl) GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Where("license_id = ?", licenseID).Find(&users).Error
	return users, err
}

// Update updates an existing user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user (soft delete)
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.User{}, "id = ?", id).Error
}

// List retrieves a list of users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}