package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// StockHistoryRepository defines the interface for stock history data access
type StockHistoryRepository interface {
	Create(ctx context.Context, stockHistory *entities.StockHistory) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.StockHistory, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.StockHistory, error)
	Update(ctx context.Context, stockHistory *entities.StockHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.StockHistory, error)
}
