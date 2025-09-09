package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	Create(ctx context.Context, role *entities.Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error)
	GetByName(ctx context.Context, name string) (*entities.Role, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Role, error)
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveRoles(ctx context.Context) ([]*entities.Role, error)
}

// PolicyRepository defines the interface for policy data access
type PolicyRepository interface {
	Create(ctx context.Context, policy *entities.Policy) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Policy, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Policy, error)
	GetByRole(ctx context.Context, roleID uuid.UUID) ([]*entities.Policy, error)
	GetByDomain(ctx context.Context, domain string) ([]*entities.Policy, error)
	Update(ctx context.Context, policy *entities.Policy) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActivePolicies(ctx context.Context) ([]*entities.Policy, error)
}

// UserRoleRepository defines the interface for user-role data access
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *entities.UserRole) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.UserRole, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserRole, error)
	GetByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) ([]*entities.UserRole, error)
	List(ctx context.Context, limit, offset int) ([]*entities.UserRole, error)
	Update(ctx context.Context, userRole *entities.UserRole) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserAndRole(ctx context.Context, userID, roleID uuid.UUID, domain string) error
}