package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// PaymentRepository defines the interface for payment data access
type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*entities.Payment, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Payment, error)
	GetByShopIDAndStatus(ctx context.Context, shopID uuid.UUID, status entities.PaymentStatus) ([]*entities.Payment, error)
	GetByStatus(ctx context.Context, status entities.PaymentStatus) ([]*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
}
