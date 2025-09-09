package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// UserDomainRepository defines the interface for user domain data access
type UserDomainRepository interface {
	Create(ctx context.Context, userDomain *entities.UserDomain) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserDomain, error)
	GetByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) (*entities.UserDomain, error)
	GetByDomain(ctx context.Context, domain string) ([]*entities.UserDomain, error)
	Update(ctx context.Context, userDomain *entities.UserDomain) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) error
	List(ctx context.Context, limit, offset int) ([]*entities.UserDomain, error)
}
