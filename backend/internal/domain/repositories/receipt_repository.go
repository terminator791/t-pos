package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ReceiptRepository defines the interface for receipt data access
type ReceiptRepository interface {
	Create(ctx context.Context, receipt *entities.Receipt) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Receipt, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Receipt, error)
	GetByPaymentID(ctx context.Context, paymentID uuid.UUID) (*entities.Receipt, error)
	Update(ctx context.Context, receipt *entities.Receipt) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Receipt, error)
}