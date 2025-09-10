package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// TransactionProductRepository defines the interface for transaction product data access
type TransactionProductRepository interface {
	Create(ctx context.Context, transactionProduct *entities.TransactionProduct) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.TransactionProduct, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*entities.TransactionProduct, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*entities.TransactionProduct, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.TransactionProduct, error)
	Update(ctx context.Context, transactionProduct *entities.TransactionProduct) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.TransactionProduct, error)
}
