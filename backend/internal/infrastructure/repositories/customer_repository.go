package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// CustomerRepositoryImpl implements CustomerRepository interface
type CustomerRepositoryImpl struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) *CustomerRepositoryImpl {
	return &CustomerRepositoryImpl{db: db}
}

// Create creates a new customer
func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *entities.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

// GetByID retrieves a customer by ID
func (r *CustomerRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Customer, error) {
	var customer entities.Customer
	err := r.db.WithContext(ctx).First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// GetByEmail retrieves a customer by email
func (r *CustomerRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.Customer, error) {
	var customer entities.Customer
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// Update updates an existing customer
func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *entities.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

// Delete deletes a customer (soft delete)
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Customer{}, id).Error
}

// List retrieves a list of customers with pagination
func (r *CustomerRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Customer, error) {
	var customers []*entities.Customer
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&customers).Error
	return customers, err
}
