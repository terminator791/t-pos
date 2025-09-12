package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Product, error)
	GetByBarcode(ctx context.Context, barcode string) (*entities.Product, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]*entities.Product, error)
	GetLowStockProducts(ctx context.Context, shopID uuid.UUID) ([]*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Product, error)
	Search(ctx context.Context, query string, shopID uuid.UUID) ([]*entities.Product, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
}
