package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	Create(ctx context.Context, order *entities.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*entities.Order, error)
	Update(ctx context.Context, order *entities.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Order, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]*entities.Order, error)
	GetByCustomer(ctx context.Context, customerID uuid.UUID) ([]*entities.Order, error)
	GetByDateRange(ctx context.Context, startDate, endDate string) ([]*entities.Order, error)
	GetTodaysOrders(ctx context.Context) ([]*entities.Order, error)
}