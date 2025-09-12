package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// HistoryRepository defines the interface for history data access
type HistoryRepository interface {
	Create(ctx context.Context, history *entities.History) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.History, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.History, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) (*entities.History, error)
	Update(ctx context.Context, history *entities.History) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.History, error)
}
