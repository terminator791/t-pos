package repositories

import (
	"context"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uint) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
	GetActiveUsers(ctx context.Context) ([]*entities.User, error)
}