package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// TransactionProductRepositoryImpl implements TransactionProductRepository interface
type TransactionProductRepositoryImpl struct {
	db *gorm.DB
}

// NewTransactionProductRepository creates a new transaction product repository
func NewTransactionProductRepository(db *gorm.DB) *TransactionProductRepositoryImpl {
	return &TransactionProductRepositoryImpl{db: db}
}

// Create creates a new transaction product
func (r *TransactionProductRepositoryImpl) Create(ctx context.Context, transactionProduct *entities.TransactionProduct) error {
	return r.db.WithContext(ctx).Create(transactionProduct).Error
}

// GetByID retrieves a transaction product by ID
func (r *TransactionProductRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.TransactionProduct, error) {
	var transactionProduct entities.TransactionProduct
	err := r.db.WithContext(ctx).
		Preload("Transaction").
		Preload("Transaction.Shop").
		Preload("Transaction.Cashier").
		Preload("Product").
		Preload("Product.Category").
		First(&transactionProduct, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transactionProduct, nil
}

// GetByTransactionID retrieves transaction products by transaction ID
func (r *TransactionProductRepositoryImpl) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*entities.TransactionProduct, error) {
	var transactionProducts []*entities.TransactionProduct
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Product.Category").
		Where("transaction_id = ?", transactionID).Find(&transactionProducts).Error
	return transactionProducts, err
}

// GetByProductID retrieves transaction products by product ID
func (r *TransactionProductRepositoryImpl) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.TransactionProduct, error) {
	var transactionProducts []*entities.TransactionProduct
	err := r.db.WithContext(ctx).
		Preload("Transaction").
		Preload("Transaction.Shop").
		Where("product_id = ?", productID).Find(&transactionProducts).Error
	return transactionProducts, err
}

// GetByShopID retrieves transaction products by shop ID through transaction
func (r *TransactionProductRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.TransactionProduct, error) {
	var transactionProducts []*entities.TransactionProduct
	err := r.db.WithContext(ctx).
		Joins("JOIN transactions ON transaction_products.transaction_id = transactions.id").
		Preload("Transaction").
		Preload("Product").
		Preload("Product.Category").
		Where("transactions.shop_id = ?", shopID).Find(&transactionProducts).Error
	return transactionProducts, err
}

// Update updates an existing transaction product
func (r *TransactionProductRepositoryImpl) Update(ctx context.Context, transactionProduct *entities.TransactionProduct) error {
	return r.db.WithContext(ctx).Save(transactionProduct).Error
}

// Delete deletes a transaction product (soft delete)
func (r *TransactionProductRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.TransactionProduct{}, "id = ?", id).Error
}

// List retrieves a list of transaction products with pagination
func (r *TransactionProductRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.TransactionProduct, error) {
	var transactionProducts []*entities.TransactionProduct
	err := r.db.WithContext(ctx).
		Preload("Transaction").
		Preload("Product").
		Limit(limit).Offset(offset).Find(&transactionProducts).Error
	return transactionProducts, err
}
