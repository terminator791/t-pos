package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	Create(ctx context.Context, transaction *entities.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Transaction, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Transaction, error)
	GetByShopIDAndStatus(ctx context.Context, shopID uuid.UUID, status entities.TransactionStatus) ([]*entities.Transaction, error)
	GetByCashierID(ctx context.Context, cashierID uuid.UUID) ([]*entities.Transaction, error)
	GetByStatus(ctx context.Context, status entities.TransactionStatus) ([]*entities.Transaction, error)
	GetTodaysTransactions(ctx context.Context, shopID uuid.UUID) ([]*entities.Transaction, error)
	Update(ctx context.Context, transaction *entities.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Transaction, error)
}
