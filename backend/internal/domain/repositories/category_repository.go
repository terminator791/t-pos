package repositories

import (
	"context"

	"github.com/terminator791/t-pos/internal/domain/entities"
)

// CategoryRepository defines the interface for category data access
type CategoryRepository interface {
	Create(ctx context.Context, category *entities.Category) error
	GetByID(ctx context.Context, id uint) (*entities.Category, error)
	GetByShopID(ctx context.Context, shopID uint) ([]*entities.Category, error)
	GetByName(ctx context.Context, name string, shopID uint) (*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.Category, error)
}