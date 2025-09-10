package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// ExpenseRepositoryImpl implements ExpenseRepository interface
type ExpenseRepositoryImpl struct {
	db *gorm.DB
}

// NewExpenseRepository creates a new expense repository
func NewExpenseRepository(db *gorm.DB) *ExpenseRepositoryImpl {
	return &ExpenseRepositoryImpl{db: db}
}

// Create creates a new expense
func (r *ExpenseRepositoryImpl) Create(ctx context.Context, expense *entities.Expense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

// GetByID retrieves an expense by ID
func (r *ExpenseRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Expense, error) {
	var expense entities.Expense
	err := r.db.WithContext(ctx).Preload("Shop").First(&expense, id).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

// GetByShopID retrieves expenses by shop ID
func (r *ExpenseRepositoryImpl) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]*entities.Expense, error) {
	var expenses []*entities.Expense
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&expenses).Error
	return expenses, err
}

// Update updates an existing expense
func (r *ExpenseRepositoryImpl) Update(ctx context.Context, expense *entities.Expense) error {
	return r.db.WithContext(ctx).Save(expense).Error
}

// Delete deletes an expense (soft delete)
func (r *ExpenseRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Expense{}, id).Error
}

// List retrieves a list of expenses with pagination
func (r *ExpenseRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Expense, error) {
	var expenses []*entities.Expense
	err := r.db.WithContext(ctx).Preload("Shop").Limit(limit).Offset(offset).Find(&expenses).Error
	return expenses, err
}