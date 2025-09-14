package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/dto"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"gorm.io/gorm"
)

// CategoryUseCase handles category-related business logic
type CategoryUseCase struct {
	db           *gorm.DB
	categoryRepo repositories.CategoryRepository
	shopRepo     repositories.ShopRepository
}

// NewCategoryUseCase creates a new CategoryUseCase
func NewCategoryUseCase(db *gorm.DB, categoryRepo repositories.CategoryRepository, shopRepo repositories.ShopRepository) *CategoryUseCase {
	return &CategoryUseCase{
		db:           db,
		categoryRepo: categoryRepo,
		shopRepo:     shopRepo,
	}
}

// CreateCategory creates a new category
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, category *entities.Category) error {
	if category.Name == "" {
		return errors.New("category name is required")
	}

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if shop exists
	_, err := uc.shopRepo.GetByID(ctx, category.ShopID)
	if err != nil {
		tx.Rollback()
		return errors.New("invalid shop ID")
	}

	// Create category
	err = tx.WithContext(ctx).Create(category).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetCategory retrieves a category by ID
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id uuid.UUID) (*entities.Category, error) {
	return uc.categoryRepo.GetByID(ctx, id)
}

// GetCategoryForDTO retrieves a category by ID as DTO
func (uc *CategoryUseCase) GetCategoryForDTO(ctx context.Context, id uuid.UUID) (*dto.CategoryListDTO, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.CategoryToListDTO(category), nil
}

// GetCategoriesByShop retrieves categories by shop ID
func (uc *CategoryUseCase) GetCategoriesByShop(ctx context.Context, shopID uuid.UUID) ([]*entities.Category, error) {
	return uc.categoryRepo.GetByShopID(ctx, shopID)
}

// UpdateCategory updates an existing category
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, category *entities.Category) error {
	if category.ID == uuid.Nil {
		return errors.New("category ID is required")
	}

	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if category exists with row lock
	var existing entities.Category
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", category.ID).First(&existing).Error
	if err != nil {
		tx.Rollback()
		return errors.New("category not found")
	}

	// Update category
	err = tx.WithContext(ctx).Save(category).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// DeleteCategory deletes a category
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Start database transaction
	tx := uc.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Ensure rollback on error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if category exists with row lock
	var existing entities.Category
	err := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE").Where("id = ?", id).First(&existing).Error
	if err != nil {
		tx.Rollback()
		return errors.New("category not found")
	}

	// Delete category
	err = tx.WithContext(ctx).Delete(&entities.Category{}, "id = ?", id).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// ListCategories retrieves a list of categories
func (uc *CategoryUseCase) ListCategories(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	return uc.categoryRepo.List(ctx, limit, offset)
}

// ListCategoriesFiltered retrieves a list of categories filtered by accessible shop IDs
func (uc *CategoryUseCase) ListCategoriesFiltered(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Category, error) {
	if len(shopIDs) == 0 {
		// If no shop IDs provided, return empty list
		return []*entities.Category{}, nil
	}
	return uc.categoryRepo.ListByShopIDs(ctx, shopIDs, limit, offset)
}

// ListCategoriesForDTO retrieves a list of categories as DTOs
func (uc *CategoryUseCase) ListCategoriesForDTO(ctx context.Context, limit, offset int) ([]*dto.CategoryListDTO, error) {
	categories, err := uc.categoryRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return dto.CategoriesToListDTO(categories), nil
}

// ListCategoriesFilteredForDTO retrieves a list of categories filtered by accessible shop IDs as DTOs
func (uc *CategoryUseCase) ListCategoriesFilteredForDTO(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*dto.CategoryListDTO, error) {
	if len(shopIDs) == 0 {
		// If no shop IDs provided, return empty list
		return []*dto.CategoryListDTO{}, nil
	}
	categories, err := uc.categoryRepo.ListByShopIDs(ctx, shopIDs, limit, offset)
	if err != nil {
		return nil, err
	}
	return dto.CategoriesToListDTO(categories), nil
}
