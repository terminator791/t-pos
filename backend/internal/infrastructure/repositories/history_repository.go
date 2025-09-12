package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// HistoryRepositoryImpl implements HistoryRepository interface
type HistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewHistoryRepository creates a new history repository
func NewHistoryRepository(db *gorm.DB) *HistoryRepositoryImpl {
	return &HistoryRepositoryImpl{db: db}
}

// Create creates a new history record
func (r *HistoryRepositoryImpl) Create(ctx context.Context, history *entities.History) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetByID retrieves a history record by ID
func (r *HistoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.History, error) {
	var history entities.History
	err := r.db.WithContext(ctx).First(&history, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// GetByShopID retrieves history records by shop ID
func (r *HistoryRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.History, error) {
	var histories []*entities.History
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&histories).Error
	return histories, err
}

// GetByTransactionID retrieves a history record by transaction ID
func (r *HistoryRepositoryImpl) GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*entities.History, error) {
	var history entities.History
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&history).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// Update updates an existing history record
func (r *HistoryRepositoryImpl) Update(ctx context.Context, history *entities.History) error {
	return r.db.WithContext(ctx).Save(history).Error
}

// Delete deletes a history record (soft delete)
func (r *HistoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.History{}, "id = ?", id).Error
}

// List retrieves a list of history records with pagination
func (r *HistoryRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.History, error) {
	var histories []*entities.History
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&histories).Error
	return histories, err
}
