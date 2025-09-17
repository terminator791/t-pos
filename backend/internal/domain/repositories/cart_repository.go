package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// CartRepository defines the interface for cart data access
type CartRepository interface {
	Create(ctx context.Context, cart *entities.Cart) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Cart, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Cart, error)
	GetByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entities.Cart, error)
	Update(ctx context.Context, cart *entities.Cart) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Cart, error)
	ListByShopIDs(ctx context.Context, shopIDs []uuid.UUID, limit, offset int) ([]*entities.Cart, error)
	GetByShopIDs(ctx context.Context, shopIDs []uuid.UUID) ([]*entities.Cart, error)
}
