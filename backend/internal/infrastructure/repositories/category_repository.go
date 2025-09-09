package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// CategoryRepositoryImpl implements CategoryRepository interface
type CategoryRepositoryImpl struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepositoryImpl {
	return &CategoryRepositoryImpl{db: db}
}

// Create creates a new category
func (r *CategoryRepositoryImpl) Create(ctx context.Context, category *entities.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID retrieves a category by ID
func (r *CategoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	var category entities.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetByShopID retrieves categories by shop ID
func (r *CategoryRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Category, error) {
	var categories []*entities.Category
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&categories).Error
	return categories, err
}

// GetByName retrieves a category by name within a shop
func (r *CategoryRepositoryImpl) GetByName(ctx context.Context, name string, shopID uuid.UUID) (*entities.Category, error) {
	var category entities.Category
	err := r.db.WithContext(ctx).Where("name = ? AND shop_id = ?", name, shopID).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// Update updates an existing category
func (r *CategoryRepositoryImpl) Update(ctx context.Context, category *entities.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete deletes a category (soft delete)
func (r *CategoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Category{}, id).Error
}

// List retrieves a list of categories with pagination
func (r *CategoryRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	var categories []*entities.Category
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&categories).Error
	return categories, err
}