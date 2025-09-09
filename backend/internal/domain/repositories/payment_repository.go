package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// PaymentRepository defines the interface for payment data access
type PaymentRepository interface {
	Create(ctx context.Context, payment *entities.Payment) error
	GetByID(ctx context.Context, id uint) (*entities.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID uuid.UUID) ([]*entities.Payment, error)
	GetByShopID(ctx context.Context, shopID uint) ([]*entities.Payment, error)
	GetByStatus(ctx context.Context, status entities.PaymentStatus) ([]*entities.Payment, error)
	Update(ctx context.Context, payment *entities.Payment) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.Payment, error)
}
