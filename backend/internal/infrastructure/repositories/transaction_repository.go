package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// TransactionRepositoryImpl implements TransactionRepository interface
type TransactionRepositoryImpl struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{db: db}
}

// Create creates a new transaction
func (r *TransactionRepositoryImpl) Create(ctx context.Context, transaction *entities.Transaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Transaction, error) {
	var transaction entities.Transaction
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Cashier").
		// Preload("TransactionProducts").
		// Preload("TransactionProducts.Product").
		// Preload("Payments").
		// Preload("Payments.Shop").
		First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetByShopID retrieves transactions by shop ID
func (r *TransactionRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&transactions).Error
	return transactions, err
}

// GetByCashierID retrieves transactions by cashier ID
func (r *TransactionRepositoryImpl) GetByCashierID(ctx context.Context, cashierID uuid.UUID) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	err := r.db.WithContext(ctx).Where("cashier_id = ?", cashierID).Find(&transactions).Error
	return transactions, err
}

// GetByStatus retrieves transactions by status
func (r *TransactionRepositoryImpl) GetByStatus(ctx context.Context, status entities.TransactionStatus) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&transactions).Error
	return transactions, err
}

// GetTodaysTransactions retrieves today's transactions for a shop
func (r *TransactionRepositoryImpl) GetTodaysTransactions(ctx context.Context, shopID uuid.UUID) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	today := time.Now().Format("2006-01-02")
	err := r.db.WithContext(ctx).Where("shop_id = ? AND DATE(created_at) = ?", shopID, today).Find(&transactions).Error
	return transactions, err
}

// Update updates an existing transaction
func (r *TransactionRepositoryImpl) Update(ctx context.Context, transaction *entities.Transaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// Delete deletes a transaction (soft delete)
func (r *TransactionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Transaction{}, "id = ?", id).Error
}

// List retrieves a list of transactions with pagination
func (r *TransactionRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Transaction, error) {
	var transactions []*entities.Transaction
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&transactions).Error
	return transactions, err
}
