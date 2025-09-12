package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// StockHistoryRepositoryImpl implements StockHistoryRepository interface
type StockHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewStockHistoryRepository creates a new stock history repository
func NewStockHistoryRepository(db *gorm.DB) *StockHistoryRepositoryImpl {
	return &StockHistoryRepositoryImpl{db: db}
}

// Create creates a new stock history record
func (r *StockHistoryRepositoryImpl) Create(ctx context.Context, stockHistory *entities.StockHistory) error {
	return r.db.WithContext(ctx).Create(stockHistory).Error
}

// GetByID retrieves a stock history record by ID
func (r *StockHistoryRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.StockHistory, error) {
	var stockHistory entities.StockHistory
	err := r.db.WithContext(ctx).Preload("Product").First(&stockHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &stockHistory, nil
}

// GetByProductID retrieves stock history records by product ID
func (r *StockHistoryRepositoryImpl) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.StockHistory, error) {
	var stockHistories []*entities.StockHistory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("stocked_at DESC").Find(&stockHistories).Error
	return stockHistories, err
}

// Update updates a stock history record
func (r *StockHistoryRepositoryImpl) Update(ctx context.Context, stockHistory *entities.StockHistory) error {
	return r.db.WithContext(ctx).Save(stockHistory).Error
}

// Delete soft deletes a stock history record
func (r *StockHistoryRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.StockHistory{}, id).Error
}

// List retrieves stock history records with pagination
func (r *StockHistoryRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.StockHistory, error) {
	var stockHistories []*entities.StockHistory
	err := r.db.WithContext(ctx).Preload("Product").Limit(limit).Offset(offset).Find(&stockHistories).Error
	return stockHistories, err
}
