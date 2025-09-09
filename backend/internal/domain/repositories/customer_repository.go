package repositories

import (
	"context"

	"github.com/terminator791/t-pos/internal/domain/entities"
)

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	Create(ctx context.Context, customer *entities.Customer) error
	GetByID(ctx context.Context, id uint) (*entities.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entities.Customer, error)
	Update(ctx context.Context, customer *entities.Customer) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.Customer, error)
}
