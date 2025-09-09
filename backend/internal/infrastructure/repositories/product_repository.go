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

// GetByBarcode retrieves a product by barcode
func (r *ProductRepositoryImpl) GetByBarcode(ctx context.Context, barcode string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("barcode = ?", barcode).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetByShopID retrieves products by shop ID
func (r *ProductRepositoryImpl) GetByShopID(ctx context.Context, shopID uint) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("shop_id = ?", shopID).Find(&products).Error
	return products, err
}

// GetByCategory retrieves products by category ID
func (r *ProductRepositoryImpl) GetByCategory(ctx context.Context, categoryID uint) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("cat_id = ?", categoryID).Find(&products).Error
	return products, err
}

// GetLowStockProducts retrieves products with low stock for a specific shop
func (r *ProductRepositoryImpl) GetLowStockProducts(ctx context.Context, shopID uint) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("shop_id = ? AND stock <= ?", shopID, 10).Find(&products).Error
	return products, err
}

// Update updates an existing product
func (r *ProductRepositoryImpl) Update(ctx context.Context, product *entities.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// UpdateStock updates product stock
func (r *ProductRepositoryImpl) UpdateStock(ctx context.Context, productID uint, quantity int) error {
	return r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", productID).Update("stock", quantity).Error
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

// Search searches for products by name or barcode within a shop
func (r *ProductRepositoryImpl) Search(ctx context.Context, query string, shopID uint) ([]*entities.Product, error) {
	var products []*entities.Product
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).Preload("Category").Where("shop_id = ? AND (name ILIKE ? OR barcode ILIKE ?)", shopID, searchQuery, searchQuery).Find(&products).Error
	return products, err
}