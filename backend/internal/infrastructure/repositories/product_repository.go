package repositories

import (
	"context"
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
	return r.db.WithContext(ctx).Create(product).Error
}

// GetByID retrieves a product by ID
func (r *ProductRepositoryImpl) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepositoryImpl) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByBarcode retrieves a product by barcode
func (r *ProductRepositoryImpl) GetByBarcode(ctx context.Context, barcode string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("barcode = ?", barcode).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Update updates an existing product
func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entities.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete deletes a product (soft delete)
func (r *ProductRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Product{}, id).Error
}

// List retrieves a list of products with pagination
func (r *ProductRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Limit(limit).Offset(offset).Find(&products).Error
	return products, err
}

// GetByCategory retrieves products by category ID
func (r *ProductRepositoryImpl) GetByCategory(ctx context.Context, categoryID uint) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("category_id = ?", categoryID).Find(&products).Error
	return products, err
}

// GetActiveProducts retrieves all active products
func (r *ProductRepositoryImpl) GetActiveProducts(ctx context.Context) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("is_active = ?", true).Find(&products).Error
	return products, err
}

// GetLowStockProducts retrieves products with low stock
func (r *ProductRepositoryImpl) GetLowStockProducts(ctx context.Context) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("stock_quantity <= min_stock_level").Find(&products).Error
	return products, err
}

// Search searches for products by name or SKU
func (r *ProductRepositoryImpl) Search(ctx context.Context, query string) ([]*entities.Product, error) {
	var products []*entities.Product
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).Preload("Category").Where("name ILIKE ? OR sku ILIKE ?", searchQuery, searchQuery).Find(&products).Error
	return products, err
}