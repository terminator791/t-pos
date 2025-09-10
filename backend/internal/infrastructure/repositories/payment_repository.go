package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// PaymentRepositoryImpl implements PaymentRepository interface
type PaymentRepositoryImpl struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) *PaymentRepositoryImpl {
	return &PaymentRepositoryImpl{db: db}
}

// Create creates a new payment
func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *entities.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// GetByID retrieves a payment by ID
func (r *PaymentRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error) {
	var payment entities.Payment
	err := r.db.WithContext(ctx).
		Preload("Shop").
		Preload("Shop.License").
		Preload("Shop.Owner").
		Preload("Transaction").
		Preload("Transaction.Cashier").
		First(&payment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByTransactionID retrieves payments by transaction ID
func (r *PaymentRepositoryImpl) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Find(&payments).Error
	return payments, err
}

// GetByShopID retrieves payments by shop ID
func (r *PaymentRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&payments).Error
	return payments, err
}

// GetByStatus retrieves payments by status
func (r *PaymentRepositoryImpl) GetByStatus(ctx context.Context, status entities.PaymentStatus) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	err := r.db.WithContext(ctx).Where("status = ?", status).Find(&payments).Error
	return payments, err
}

// Update updates an existing payment
func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *entities.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// Delete deletes a payment (soft delete)
func (r *PaymentRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Payment{}, "id = ?", id).Error
}

// List retrieves a list of payments with pagination
func (r *PaymentRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Payment, error) {
	var payments []*entities.Payment
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}
