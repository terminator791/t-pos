package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ShopRepository defines the interface for shop data access
type ShopRepository interface {
	Create(ctx context.Context, shop *entities.Shop) error
	GetByID(ctx context.Context, id uint) (*entities.Shop, error)
	GetByLicenseID(ctx context.Context, licenseID uuid.UUID) ([]*entities.Shop, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*entities.Shop, error)
	Update(ctx context.Context, shop *entities.Shop) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.Shop, error)
}
