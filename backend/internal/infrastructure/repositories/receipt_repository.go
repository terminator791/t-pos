package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// ReceiptRepositoryImpl implements ReceiptRepository interface
type ReceiptRepositoryImpl struct {
	db *gorm.DB
}

// NewReceiptRepository creates a new receipt repository
func NewReceiptRepository(db *gorm.DB) *ReceiptRepositoryImpl {
	return &ReceiptRepositoryImpl{db: db}
}

// Create creates a new receipt record
func (r *ReceiptRepositoryImpl) Create(ctx context.Context, receipt *entities.Receipt) error {
	return r.db.WithContext(ctx).Create(receipt).Error
}

// GetByID retrieves a receipt record by ID
func (r *ReceiptRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Receipt, error) {
	var receipt entities.Receipt
	err := r.db.WithContext(ctx).First(&receipt, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetByShopID retrieves receipt records by shop ID
func (r *ReceiptRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Receipt, error) {
	var receipts []*entities.Receipt
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&receipts).Error
	return receipts, err
}

// GetByPaymentID retrieves a receipt record by payment ID
func (r *ReceiptRepositoryImpl) GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Receipt, error) {
	var receipt entities.Receipt
	err := r.db.WithContext(ctx).Where("payments_id = ?", paymentID).First(&receipt).Error
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// Update updates an existing receipt record
func (r *ReceiptRepositoryImpl) Update(ctx context.Context, receipt *entities.Receipt) error {
	return r.db.WithContext(ctx).Save(receipt).Error
}

// Delete deletes a receipt record (soft delete)
func (r *ReceiptRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Receipt{}, "id = ?", id).Error
}

// List retrieves a list of receipt records with pagination
func (r *ReceiptRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Receipt, error) {
	var receipts []*entities.Receipt
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&receipts).Error
	return receipts, err
}

// ListByShopIDs retrieves receipts filtered by accessible shop IDs for multi-tenant access
func (r *ReceiptRepositoryImpl) ListByShopIDs(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Receipt, error) {
	var receipts []*entities.Receipt
	if len(shopIDs) == 0 {
		return receipts, nil // Return empty slice if no accessible shops
	}
	
	err := r.db.WithContext(ctx).
		Where("shop_id IN (?)", shopIDs).
		Limit(limit).
		Offset(offset).
		Find(&receipts).Error
		
	return receipts, err
}

// GetByShopIDs retrieves all receipts for specified shop IDs (no pagination)
func (r *ReceiptRepositoryImpl) GetByShopIDs(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Receipt, error) {
	var receipts []*entities.Receipt
	if len(shopIDs) == 0 {
		return receipts, nil // Return empty slice if no accessible shops
	}
	
	err := r.db.WithContext(ctx).
		Where("shop_id IN (?)", shopIDs).
		Find(&receipts).Error
		
	return receipts, err
}
