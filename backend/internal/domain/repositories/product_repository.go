package repositories

import (
	"context"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id uint) (*entities.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)
	GetByBarcode(ctx context.Context, barcode string) (*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.Product, error)
	GetByCategory(ctx context.Context, categoryID uint) ([]*entities.Product, error)
	GetActiveProducts(ctx context.Context) ([]*entities.Product, error)
	GetLowStockProducts(ctx context.Context) ([]*entities.Product, error)
	Search(ctx context.Context, query string) ([]*entities.Product, error)
}