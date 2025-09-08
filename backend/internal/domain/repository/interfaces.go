package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entity"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.User, error)
	Count(ctx context.Context) (int64, error)
}

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	GetByPhone(ctx context.Context, phone string) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Customer, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*entity.Customer, error)
	Count(ctx context.Context) (int64, error)
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetByName(ctx context.Context, name string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Category, error)
	GetChildren(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error)
	GetRoot(ctx context.Context) ([]*entity.Category, error)
	Count(ctx context.Context) (int64, error)
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetByBarcode(ctx context.Context, barcode string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Product, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID, offset, limit int) ([]*entity.Product, error)
	Search(ctx context.Context, query string, offset, limit int) ([]*entity.Product, error)
	GetLowStock(ctx context.Context, offset, limit int) ([]*entity.Product, error)
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
	Count(ctx context.Context) (int64, error)
}

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Order, error)
	GetByCustomer(ctx context.Context, customerID uuid.UUID, offset, limit int) ([]*entity.Order, error)
	GetByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.Order, error)
	GetByStatus(ctx context.Context, status entity.OrderStatus, offset, limit int) ([]*entity.Order, error)
	GetByDateRange(ctx context.Context, startDate, endDate string, offset, limit int) ([]*entity.Order, error)
	Count(ctx context.Context) (int64, error)
	GetTotalSales(ctx context.Context, startDate, endDate string) (float64, error)
}

// OrderItemRepository defines the interface for order item data operations
type OrderItemRepository interface {
	Create(ctx context.Context, orderItem *entity.OrderItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.OrderItem, error)
	GetByOrder(ctx context.Context, orderID uuid.UUID) ([]*entity.OrderItem, error)
	Update(ctx context.Context, orderItem *entity.OrderItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOrder(ctx context.Context, orderID uuid.UUID) error
}

// PaymentRepository defines the interface for payment data operations
type PaymentRepository interface {
	Create(ctx context.Context, payment *entity.Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	GetByOrder(ctx context.Context, orderID uuid.UUID) ([]*entity.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*entity.Payment, error)
	Update(ctx context.Context, payment *entity.Payment) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Payment, error)
	GetByMethod(ctx context.Context, method entity.PaymentMethod, offset, limit int) ([]*entity.Payment, error)
	GetByDateRange(ctx context.Context, startDate, endDate string, offset, limit int) ([]*entity.Payment, error)
	Count(ctx context.Context) (int64, error)
	GetTotalByMethod(ctx context.Context, method entity.PaymentMethod, startDate, endDate string) (float64, error)
}