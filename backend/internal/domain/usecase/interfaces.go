package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entity"
)

// AuthUsecase defines authentication and authorization operations
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (*entity.User, string, error)
	Register(ctx context.Context, user *entity.User, password string) error
	RefreshToken(ctx context.Context, token string) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	VerifyToken(ctx context.Context, token string) (*entity.User, error)
}

// UserUsecase defines user management operations
type UserUsecase interface {
	CreateUser(ctx context.Context, user *entity.User) error
	GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, page, limit int) ([]*entity.User, int64, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	ResetPassword(ctx context.Context, email string) error
}

// CustomerUsecase defines customer management operations
type CustomerUsecase interface {
	CreateCustomer(ctx context.Context, customer *entity.Customer) error
	GetCustomer(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	UpdateCustomer(ctx context.Context, customer *entity.Customer) error
	DeleteCustomer(ctx context.Context, id uuid.UUID) error
	ListCustomers(ctx context.Context, page, limit int) ([]*entity.Customer, int64, error)
	SearchCustomers(ctx context.Context, query string, page, limit int) ([]*entity.Customer, int64, error)
	GetCustomerOrders(ctx context.Context, customerID uuid.UUID, page, limit int) ([]*entity.Order, int64, error)
}

// CategoryUsecase defines category management operations
type CategoryUsecase interface {
	CreateCategory(ctx context.Context, category *entity.Category) error
	GetCategory(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	UpdateCategory(ctx context.Context, category *entity.Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListCategories(ctx context.Context, page, limit int) ([]*entity.Category, int64, error)
	GetCategoryTree(ctx context.Context) ([]*entity.Category, error)
	GetChildCategories(ctx context.Context, parentID uuid.UUID) ([]*entity.Category, error)
}

// ProductUsecase defines product management operations
type ProductUsecase interface {
	CreateProduct(ctx context.Context, product *entity.Product) error
	GetProduct(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetProductByBarcode(ctx context.Context, barcode string) (*entity.Product, error)
	UpdateProduct(ctx context.Context, product *entity.Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	ListProducts(ctx context.Context, page, limit int) ([]*entity.Product, int64, error)
	SearchProducts(ctx context.Context, query string, page, limit int) ([]*entity.Product, int64, error)
	GetProductsByCategory(ctx context.Context, categoryID uuid.UUID, page, limit int) ([]*entity.Product, int64, error)
	GetLowStockProducts(ctx context.Context, page, limit int) ([]*entity.Product, int64, error)
	UpdateProductStock(ctx context.Context, productID uuid.UUID, quantity int) error
	BulkUpdateStock(ctx context.Context, updates []StockUpdate) error
}

// StockUpdate represents a stock update operation
type StockUpdate struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

// OrderUsecase defines order management operations
type OrderUsecase interface {
	CreateOrder(ctx context.Context, order *entity.Order, items []*entity.OrderItem) error
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	UpdateOrder(ctx context.Context, order *entity.Order) error
	CancelOrder(ctx context.Context, orderID uuid.UUID, reason string) error
	RefundOrder(ctx context.Context, orderID uuid.UUID, amount float64, reason string) error
	ListOrders(ctx context.Context, page, limit int) ([]*entity.Order, int64, error)
	GetOrdersByStatus(ctx context.Context, status entity.OrderStatus, page, limit int) ([]*entity.Order, int64, error)
	GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, page, limit int) ([]*entity.Order, int64, error)
	GetOrdersByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Order, int64, error)
	GetOrdersByDateRange(ctx context.Context, startDate, endDate string, page, limit int) ([]*entity.Order, int64, error)
	AddOrderItem(ctx context.Context, orderID uuid.UUID, item *entity.OrderItem) error
	RemoveOrderItem(ctx context.Context, orderItemID uuid.UUID) error
	UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error
	ProcessOrder(ctx context.Context, orderID uuid.UUID) error
	CompleteOrder(ctx context.Context, orderID uuid.UUID) error
}

// PaymentUsecase defines payment processing operations
type PaymentUsecase interface {
	ProcessPayment(ctx context.Context, payment *entity.Payment) error
	GetPayment(ctx context.Context, id uuid.UUID) (*entity.Payment, error)
	GetOrderPayments(ctx context.Context, orderID uuid.UUID) ([]*entity.Payment, error)
	RefundPayment(ctx context.Context, paymentID uuid.UUID, amount float64, reason string) error
	ListPayments(ctx context.Context, page, limit int) ([]*entity.Payment, int64, error)
	GetPaymentsByMethod(ctx context.Context, method entity.PaymentMethod, page, limit int) ([]*entity.Payment, int64, error)
	GetPaymentsByDateRange(ctx context.Context, startDate, endDate string, page, limit int) ([]*entity.Payment, int64, error)
	GetPaymentStats(ctx context.Context, startDate, endDate string) (*PaymentStats, error)
	ValidatePayment(ctx context.Context, payment *entity.Payment) error
}

// PaymentStats represents payment statistics
type PaymentStats struct {
	TotalAmount       float64                    `json:"total_amount"`
	TransactionCount  int64                      `json:"transaction_count"`
	AverageAmount     float64                    `json:"average_amount"`
	ByMethod          map[entity.PaymentMethod]float64 `json:"by_method"`
	ByStatus          map[entity.PaymentStatus]int64   `json:"by_status"`
}

// AnalyticsUsecase defines business analytics operations
type AnalyticsUsecase interface {
	GetSalesReport(ctx context.Context, startDate, endDate string) (*SalesReport, error)
	GetProductReport(ctx context.Context, startDate, endDate string) (*ProductReport, error)
	GetCustomerReport(ctx context.Context, startDate, endDate string) (*CustomerReport, error)
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetRevenueByPeriod(ctx context.Context, period string, startDate, endDate string) ([]RevenueData, error)
	GetTopProducts(ctx context.Context, limit int, startDate, endDate string) ([]*ProductSales, error)
	GetTopCustomers(ctx context.Context, limit int, startDate, endDate string) ([]*CustomerSales, error)
}

// SalesReport represents sales analytics
type SalesReport struct {
	TotalSales       float64                    `json:"total_sales"`
	TotalOrders      int64                      `json:"total_orders"`
	AverageOrderValue float64                   `json:"average_order_value"`
	ByStatus         map[entity.OrderStatus]int64 `json:"by_status"`
	ByPaymentMethod  map[entity.PaymentMethod]float64 `json:"by_payment_method"`
	DailySales       []DailySales               `json:"daily_sales"`
}

// ProductReport represents product analytics
type ProductReport struct {
	TotalProducts     int64                    `json:"total_products"`
	LowStockProducts  int64                    `json:"low_stock_products"`
	TopSellingProducts []*ProductSales         `json:"top_selling_products"`
	ByCategory        map[string]int64         `json:"by_category"`
}

// CustomerReport represents customer analytics
type CustomerReport struct {
	TotalCustomers     int64                    `json:"total_customers"`
	NewCustomers       int64                    `json:"new_customers"`
	TopCustomers       []*CustomerSales         `json:"top_customers"`
	AverageOrderValue  float64                  `json:"average_order_value"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TodaySales        float64 `json:"today_sales"`
	TodayOrders       int64   `json:"today_orders"`
	TotalCustomers    int64   `json:"total_customers"`
	TotalProducts     int64   `json:"total_products"`
	LowStockProducts  int64   `json:"low_stock_products"`
	PendingOrders     int64   `json:"pending_orders"`
}

// DailySales represents daily sales data
type DailySales struct {
	Date   string  `json:"date"`
	Sales  float64 `json:"sales"`
	Orders int64   `json:"orders"`
}

// RevenueData represents revenue by period
type RevenueData struct {
	Period string  `json:"period"`
	Revenue float64 `json:"revenue"`
	Orders  int64   `json:"orders"`
}

// ProductSales represents product sales data
type ProductSales struct {
	Product     *entity.Product `json:"product"`
	TotalSales  float64         `json:"total_sales"`
	UnitsSold   int64           `json:"units_sold"`
	Revenue     float64         `json:"revenue"`
}

// CustomerSales represents customer sales data
type CustomerSales struct {
	Customer    *entity.Customer `json:"customer"`
	TotalOrders int64            `json:"total_orders"`
	TotalSpent  float64          `json:"total_spent"`
	LastOrder   *time.Time       `json:"last_order"`
}