package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// ExpenseRepository defines the interface for expense data access
type ExpenseRepository interface {
	Create(ctx context.Context, expense *entities.Expense) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Expense, error)
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Expense, error)
	GetByShopIDAndStatus(ctx context.Context, shopID uuid.UUID, status entities.ExpenseStatus) ([]*entities.Expense, error)
	GetByStatus(ctx context.Context, status entities.ExpenseStatus) ([]*entities.Expense, error)
	Update(ctx context.Context, expense *entities.Expense) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Expense, error)
}