package repositories

import (
	"context"
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
func (r *CustomerRepositoryImpl) GetByID(ctx context.Context, id uint) (*entities.Customer, error) {
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

// GetByPhone retrieves a customer by phone
func (r *CustomerRepositoryImpl) GetByPhone(ctx context.Context, phone string) (*entities.Customer, error) {
	var customer entities.Customer
	err := r.db.WithContext(ctx).Where("phone = ?", phone).First(&customer).Error
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
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entities.Customer{}, id).Error
}

// List retrieves a list of customers with pagination
func (r *CustomerRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Customer, error) {
	var customers []*entities.Customer
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&customers).Error
	return customers, err
}

// Search searches for customers by name, email, or phone
func (r *CustomerRepositoryImpl) Search(ctx context.Context, query string) ([]*entities.Customer, error) {
	var customers []*entities.Customer
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).Where(
		"first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery,
	).Find(&customers).Error
	return customers, err
}