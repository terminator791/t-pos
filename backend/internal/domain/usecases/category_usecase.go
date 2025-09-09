package usecases

import (
	"context"
	"errors"

	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// CategoryUseCase handles category-related business logic
type CategoryUseCase struct {
	categoryRepo repositories.CategoryRepository
	shopRepo     repositories.ShopRepository
}

// NewCategoryUseCase creates a new CategoryUseCase
func NewCategoryUseCase(categoryRepo repositories.CategoryRepository, shopRepo repositories.ShopRepository) *CategoryUseCase {
	return &CategoryUseCase{
		categoryRepo: categoryRepo,
		shopRepo:     shopRepo,
	}
}

// CreateCategory creates a new category
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, category *entities.Category) error {
	if category.Name == "" {
		return errors.New("category name is required")
	}

	// Check if shop exists
	_, err := uc.shopRepo.GetByID(ctx, category.ShopID)
	if err != nil {
		return errors.New("invalid shop ID")
	}

	return uc.categoryRepo.Create(ctx, category)
}

// GetCategory retrieves a category by ID
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id uint) (*entities.Category, error) {
	return uc.categoryRepo.GetByID(ctx, id)
}

// GetCategoriesByShop retrieves categories by shop ID
func (uc *CategoryUseCase) GetCategoriesByShop(ctx context.Context, shopID uint) ([]*entities.Category, error) {
	return uc.categoryRepo.GetByShopID(ctx, shopID)
}

// UpdateCategory updates an existing category
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, category *entities.Category) error {
	if category.ID == 0 {
		return errors.New("category ID is required")
	}

	existing, err := uc.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("category not found")
	}

	return uc.categoryRepo.Update(ctx, category)
}

// DeleteCategory deletes a category
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id uint) error {
	existing, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("category not found")
	}

	return uc.categoryRepo.Delete(ctx, id)
}

// ListCategories retrieves a list of categories
func (uc *CategoryUseCase) ListCategories(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	return uc.categoryRepo.List(ctx, limit, offset)
}