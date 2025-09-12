package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// ProductRepositoryImpl implements ProductRepository interface
type ProductRepositoryImpl struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) *ProductRepositoryImpl {
	return &ProductRepositoryImpl{db: db}
}

// Create creates a new product
func (r *ProductRepositoryImpl) Create(ctx context.Context, product *entities.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return err
	}

	// Reload with relationships
	return r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		First(product, product.ID).Error
}

// GetByID retrieves a product by ID
func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByBarcode retrieves a product by barcode
func (r *ProductRepositoryImpl) GetByBarcode(ctx context.Context, barcode string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("barcode = ?", barcode).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByShopID retrieves products by shop ID
func (r *ProductRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id = ?", shopID).Find(&products).Error
	return products, err
}

// GetByCategory retrieves products by category ID
func (r *ProductRepositoryImpl) GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("cat_id = ?", categoryID).Find(&products).Error
	return products, err
}

// GetLowStockProducts retrieves products with low stock for a specific shop
func (r *ProductRepositoryImpl) GetLowStockProducts(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id = ? AND stock <= ?", shopID, 10).Find(&products).Error
	return products, err
}

// Update updates an existing product
func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entities.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return err
	}

	// Reload with relationships
	return r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		First(product, product.ID).Error
}

// UpdateStock updates product stock
func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", productID).Update("stock", quantity).Error
}

// Delete deletes a product (soft delete)
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Product{}, id).Error
}

// List retrieves a list of products with pagination
func (r *ProductRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

// Search searches for products by name or barcode within a shop
func (r *ProductRepositoryImpl) Search(ctx context.Context, query string, shopID uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id = ? AND (name ILIKE ? OR barcode ILIKE ?)", shopID, searchQuery, searchQuery).Find(&products).Error
	return products, err
}

// GetByShopIDs retrieves products by multiple shop IDs
func (r *ProductRepositoryImpl) GetByShopIDs(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	if len(shopIDs) == 0 {
		return products, nil
	}
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id IN ?", shopIDs).Find(&products).Error
	return products, err
}

// ListByShopIDs retrieves a list of products filtered by shop IDs with pagination
func (r *ProductRepositoryImpl) ListByShopIDs(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	if len(shopIDs) == 0 {
		return products, nil
	}
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id IN ?", shopIDs).Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

// GetLowStockProductsByShopIDs retrieves low stock products filtered by multiple shop IDs
func (r *ProductRepositoryImpl) GetLowStockProductsByShopIDs(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	if len(shopIDs) == 0 {
		return products, nil
	}
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id IN ? AND stock <= 10 AND is_have_stock = true", shopIDs).Find(&products).Error
	return products, err
}

// SearchByShopIDs searches for products by name or barcode within multiple shops
func (r *ProductRepositoryImpl) SearchByShopIDs(ctx context.Context, query string, shopIDs []uuid.UUID) ([]*entities.Product, error) {
	var products []*entities.Product
	if len(shopIDs) == 0 {
		return products, nil
	}
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Category").
		Where("shop_id IN ? AND (name ILIKE ? OR barcode ILIKE ?)", shopIDs, searchQuery, searchQuery).Find(&products).Error
	return products, err
}

// ListForDTO retrieves a list of products with limited preloading for DTO conversion
func (r *ProductRepositoryImpl) ListForDTO(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).
		Preload("Shop").     // Only preload Shop without License/Owner
		Preload("Category"). // Only preload Category without Shop
		Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

// ListByShopIDsForDTO retrieves a list of products filtered by shop IDs with limited preloading
func (r *ProductRepositoryImpl) ListByShopIDsForDTO(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	if len(shopIDs) == 0 {
		return products, nil
	}
	err := r.db.WithContext(ctx).
		Preload("Shop").     // Only preload Shop without License/Owner
		Preload("Category"). // Only preload Category without Shop
		Where("shop_id IN ?", shopIDs).Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}
