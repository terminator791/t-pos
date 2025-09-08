package usecases

import (
	"context"
	"errors"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// ProductUseCase handles product-related business logic
type ProductUseCase struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

// NewProductUseCase creates a new ProductUseCase
func NewProductUseCase(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) *ProductUseCase {
	return &ProductUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

// CreateProduct creates a new product
func (uc *ProductUseCase) CreateProduct(ctx context.Context, product *entities.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.SKU == "" {
		return errors.New("product SKU is required")
	}
	if product.Price <= 0 {
		return errors.New("product price must be greater than 0")
	}

	// Check if category exists if provided
	if product.CategoryID != nil {
		_, err := uc.categoryRepo.GetByID(ctx, *product.CategoryID)
		if err != nil {
			return errors.New("invalid category ID")
		}
	}

	return uc.productRepo.Create(ctx, product)
}

// GetProduct retrieves a product by ID
func (uc *ProductUseCase) GetProduct(ctx context.Context, id uint) (*entities.Product, error) {
	return uc.productRepo.GetByID(ctx, id)
}

// GetProductBySKU retrieves a product by SKU
func (uc *ProductUseCase) GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	return uc.productRepo.GetBySKU(ctx, sku)
}

// GetProductByBarcode retrieves a product by barcode
func (uc *ProductUseCase) GetProductByBarcode(ctx context.Context, barcode string) (*entities.Product, error) {
	return uc.productRepo.GetByBarcode(ctx, barcode)
}

// UpdateProduct updates an existing product
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, product *entities.Product) error {
	if product.ID == 0 {
		return errors.New("product ID is required")
	}

	existing, err := uc.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	return uc.productRepo.Update(ctx, product)
}

// DeleteProduct deletes a product
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id uint) error {
	existing, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	return uc.productRepo.Delete(ctx, id)
}

// ListProducts retrieves a list of products
func (uc *ProductUseCase) ListProducts(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	return uc.productRepo.List(ctx, limit, offset)
}

// GetLowStockProducts retrieves products with low stock
func (uc *ProductUseCase) GetLowStockProducts(ctx context.Context) ([]*entities.Product, error) {
	return uc.productRepo.GetLowStockProducts(ctx)
}

// SearchProducts searches for products by name or SKU
func (uc *ProductUseCase) SearchProducts(ctx context.Context, query string) ([]*entities.Product, error) {
	return uc.productRepo.Search(ctx, query)
}