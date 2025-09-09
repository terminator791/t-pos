package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// ShopRepositoryImpl implements ShopRepository interface
type ShopRepositoryImpl struct {
	db *gorm.DB
}

// NewShopRepository creates a new shop repository
func NewShopRepository(db *gorm.DB) *ShopRepositoryImpl {
	return &ShopRepositoryImpl{db: db}
}

// Create creates a new shop
func (r *ShopRepositoryImpl) Create(ctx context.Context, shop *entities.Shop) error {
	return r.db.WithContext(ctx).Create(shop).Error
}

// GetByID retrieves a shop by ID
func (r *ShopRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Shop, error) {
	var shop entities.Shop
	err := r.db.WithContext(ctx).First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByLicenseID retrieves shops by license ID
func (r *ShopRepositoryImpl) GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.Shop, error) {
	var shops []*entities.Shop
	err := r.db.WithContext(ctx).Where("license_id = ?", licenseID).Find(&shops).Error
	return shops, err
}

// GetFirstByLicenseID retrieves the first shop by license ID
func (r *ShopRepositoryImpl) GetFirstByLicenseID(ctx context.Context, licenseID uuid.UUID) (*entities.Shop, error) {
	var shop entities.Shop
	err := r.db.WithContext(ctx).Where("license_id = ?", licenseID).First(&shop).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

// GetByOwnerID retrieves shops by owner ID
func (r *ShopRepositoryImpl) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entities.Shop, error) {
	var shops []*entities.Shop
	err := r.db.WithContext(ctx).Where("user_id = ?", ownerID).Find(&shops).Error
	return shops, err
}

// Update updates an existing shop
func (r *ShopRepositoryImpl) Update(ctx context.Context, shop *entities.Shop) error {
	return r.db.WithContext(ctx).Save(shop).Error
}

// Delete deletes a shop (soft delete)
func (r *ShopRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Shop{}, id).Error
}

// List retrieves a list of shops with pagination
func (r *ShopRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Shop, error) {
	var shops []*entities.Shop
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&shops).Error
	return shops, err
}
