package usecases

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// ProductUseCase handles product-related business logic
type ProductUseCase struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	shopRepo     repositories.ShopRepository
}

// NewProductUseCase creates a new ProductUseCase
func NewProductUseCase(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository, shopRepo repositories.ShopRepository) *ProductUseCase {
	return &ProductUseCase{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		shopRepo:     shopRepo,
	}
}

// CreateProduct creates a new product
func (uc *ProductUseCase) CreateProduct(ctx context.Context, product *entities.Product) error {
	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Sale <= 0 {
		return errors.New("product sale price must be greater than 0")
	}
	if product.Buy <= 0 {
		return errors.New("product buy price must be greater than 0")
	}

	// Check if shop exists
	_, err := uc.shopRepo.GetByID(ctx, product.ShopID)
	if err != nil {
		return errors.New("invalid shop ID")
	}

	// Check if category exists if provided
	if product.CatID != nil {
		_, err := uc.categoryRepo.GetByID(ctx, *product.CatID)
		if err != nil {
			return errors.New("invalid category ID")
		}
	}

	// Calculate profit
	product.CalculateProfit()

	return uc.productRepo.Create(ctx, product)
}

// GetProduct retrieves a product by ID
func (uc *ProductUseCase) GetProduct(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	return uc.productRepo.GetByID(ctx, id)
}

// GetProductByBarcode retrieves a product by barcode
func (uc *ProductUseCase) GetProductByBarcode(ctx context.Context, barcode string) (*entities.Product, error) {
	return uc.productRepo.GetByBarcode(ctx, barcode)
}

// GetProductsByShop retrieves products by shop ID
func (uc *ProductUseCase) GetProductsByShop(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error) {
	return uc.productRepo.GetByShopID(ctx, shopID)
}

// UpdateProduct updates an existing product
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, product *entities.Product) error {
	if product.ID == uuid.Nil {
		return errors.New("product ID is required")
	}

	existing, err := uc.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	// Calculate profit
	product.CalculateProfit()

	return uc.productRepo.Update(ctx, product)
}

// UpdateProductStock updates product stock
func (uc *ProductUseCase) UpdateProductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	existing, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	return uc.productRepo.UpdateStock(ctx, productID, quantity)
}

// DeleteProduct deletes a product
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id uuid.UUID) error {
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
func (uc *ProductUseCase) GetLowStockProducts(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error) {
	return uc.productRepo.GetLowStockProducts(ctx, shopID)
}

// SearchProducts searches for products by name or barcode within a shop
func (uc *ProductUseCase) SearchProducts(ctx context.Context, query string, shopID uuid.UUID) ([]*entities.Product, error) {
	return uc.productRepo.Search(ctx, query, shopID)
}

// ListProductsFiltered retrieves a list of products filtered by accessible shop IDs
func (uc *ProductUseCase) ListProductsFiltered(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Product, error) {
	if len(shopIDs) == 0 {
		// If no shop IDs provided, return empty list
		return []*entities.Product{}, nil
	}
	return uc.productRepo.ListByShopIDs(ctx, shopIDs, limit, offset)
}

// GetLowStockProductsFiltered retrieves low stock products filtered by accessible shop IDs
func (uc *ProductUseCase) GetLowStockProductsFiltered(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Product, error) {
	if len(shopIDs) == 0 {
		// If no shop IDs provided, return empty list
		return []*entities.Product{}, nil
	}
	return uc.productRepo.GetLowStockProductsByShopIDs(ctx, shopIDs)
}

// SearchProductsFiltered searches for products by name or barcode within accessible shops
func (uc *ProductUseCase) SearchProductsFiltered(ctx context.Context, query string, shopIDs []uuid.UUID) ([]*entities.Product, error) {
	if len(shopIDs) == 0 {
		// If no shop IDs provided, return empty list
		return []*entities.Product{}, nil
	}
	return uc.productRepo.SearchByShopIDs(ctx, query, shopIDs)
}
