package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// OrderUseCase handles order-related business logic
type OrderUseCase struct {
	orderRepo    repositories.OrderRepository
	productRepo  repositories.ProductRepository
	customerRepo repositories.CustomerRepository
	userRepo     repositories.UserRepository
}

// NewOrderUseCase creates a new OrderUseCase
func NewOrderUseCase(
	orderRepo repositories.OrderRepository,
	productRepo repositories.ProductRepository,
	customerRepo repositories.CustomerRepository,
	userRepo repositories.UserRepository,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:    orderRepo,
		productRepo:  productRepo,
		customerRepo: customerRepo,
		userRepo:     userRepo,
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	UserID         uuid.UUID                `json:"user_id"`
	CustomerID     *uint                     `json:"customer_id"`
	Items          []CreateOrderItemRequest  `json:"items"`
	PaymentMethod  string                    `json:"payment_method"`
	DiscountAmount float64                   `json:"discount_amount"`
	TaxRate        float64                   `json:"tax_rate"`
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// CreateOrder creates a new order
func (uc *OrderUseCase) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*entities.Order, error) {
	// Validate user exists
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Validate customer exists if provided
	if req.CustomerID != nil {
		_, err := uc.customerRepo.GetByID(ctx, *req.CustomerID)
		if err != nil {
			return nil, errors.New("invalid customer ID")
		}
	}

	// Validate items
	if len(req.Items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	// Create order
	order := &entities.Order{
		OrderNumber:    uc.generateOrderNumber(),
		UserID:         req.UserID,
		CustomerID:     req.CustomerID,
		PaymentMethod:  req.PaymentMethod,
		DiscountAmount: req.DiscountAmount,
		Status:         "pending",
	}

	var orderItems []entities.OrderItem
	var subtotal float64

	// Process each item
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, errors.New("item quantity must be greater than 0")
		}

		product, err := uc.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID: %d", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		orderItem := entities.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Sale,
		}
		orderItem.CalculateTotal()

		orderItems = append(orderItems, orderItem)
		subtotal += orderItem.TotalPrice
	}

	// Calculate totals
	order.Subtotal = subtotal
	order.TaxAmount = subtotal * req.TaxRate
	order.CalculateTotal()

	// Save order
	err = uc.orderRepo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Save order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	order.OrderItems = orderItems

	return order, nil
}

// GetOrder retrieves an order by ID
func (uc *OrderUseCase) GetOrder(ctx context.Context, id uint) (*entities.Order, error) {
	return uc.orderRepo.GetByID(ctx, id)
}

// GetOrderByNumber retrieves an order by order number
func (uc *OrderUseCase) GetOrderByNumber(ctx context.Context, orderNumber string) (*entities.Order, error) {
	return uc.orderRepo.GetByOrderNumber(ctx, orderNumber)
}

// ListOrders retrieves a list of orders
func (uc *OrderUseCase) ListOrders(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	return uc.orderRepo.List(ctx, limit, offset)
}

// GetTodaysOrders retrieves today's orders
func (uc *OrderUseCase) GetTodaysOrders(ctx context.Context) ([]*entities.Order, error) {
	return uc.orderRepo.GetTodaysOrders(ctx)
}

// generateOrderNumber generates a unique order number
func (uc *OrderUseCase) generateOrderNumber() string {
	now := time.Now()
	return fmt.Sprintf("ORD-%d%02d%02d-%d", 
		now.Year(), now.Month(), now.Day(), now.Unix()%10000)
}