package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// OrderRepositoryImpl implements OrderRepository interface
type OrderRepositoryImpl struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{db: db}
}

// Create creates a new order
func (r *OrderRepositoryImpl) Create(ctx context.Context, order *entities.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// GetByID retrieves an order by ID
func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entities.Order, error) {
	var order entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Customer").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("Payments").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetByOrderNumber retrieves an order by order number
func (r *OrderRepositoryImpl) GetByOrderNumber(ctx context.Context, orderNumber string) (*entities.Order, error) {
	var order entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Customer").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("Payments").
		Where("order_number = ?", orderNumber).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update updates an existing order
func (r *OrderRepositoryImpl) Update(ctx context.Context, order *entities.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// Delete deletes an order (soft delete)
func (r *OrderRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Order{}, id).Error
}

// List retrieves a list of orders with pagination
func (r *OrderRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Customer").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	return orders, err
}

// GetByUser retrieves orders by user ID
func (r *OrderRepositoryImpl) GetByUser(ctx context.Context, userID uuid.UUID) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("OrderItems").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// GetByCustomer retrieves orders by customer ID
func (r *OrderRepositoryImpl) GetByCustomer(ctx context.Context, customerID uuid.UUID) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("OrderItems").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// GetByDateRange retrieves orders within a date range
func (r *OrderRepositoryImpl) GetByDateRange(ctx context.Context, startDate, endDate string) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Customer").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}

// GetTodaysOrders retrieves today's orders
func (r *OrderRepositoryImpl) GetTodaysOrders(ctx context.Context) ([]*entities.Order, error) {
	var orders []*entities.Order
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Customer").
		Preload("OrderItems").
		Where("DATE(created_at) = CURRENT_DATE").
		Order("created_at DESC").
		Find(&orders).Error
	return orders, err
}