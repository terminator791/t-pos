package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a system user (admin, cashier, manager)
type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	FirstName string         `json:"first_name" gorm:"not null"`
	LastName  string         `json:"last_name" gorm:"not null"`
	Role      UserRole       `json:"role" gorm:"type:varchar(20);not null;default:'cashier'"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time     `json:"last_login"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// UserRole represents user roles in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleCashier  UserRole = "cashier"
	RoleEmployee UserRole = "employee"
)

// Customer represents a customer in the POS system
type Customer struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FirstName   string         `json:"first_name" gorm:"not null"`
	LastName    string         `json:"last_name" gorm:"not null"`
	Email       string         `json:"email" gorm:"uniqueIndex"`
	Phone       string         `json:"phone" gorm:"uniqueIndex"`
	Address     string         `json:"address"`
	City        string         `json:"city"`
	State       string         `json:"state"`
	ZipCode     string         `json:"zip_code"`
	Country     string         `json:"country" gorm:"default:'US'"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Orders []Order `json:"orders" gorm:"foreignKey:CustomerID"`
}

// Category represents product categories
type Category struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null;uniqueIndex"`
	Description string         `json:"description"`
	ParentID    *uuid.UUID     `json:"parent_id" gorm:"type:uuid"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Parent     *Category `json:"parent" gorm:"foreignKey:ParentID"`
	Children   []Category `json:"children" gorm:"foreignKey:ParentID"`
	Products   []Product  `json:"products" gorm:"foreignKey:CategoryID"`
}

// Product represents items for sale
type Product struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SKU          string         `json:"sku" gorm:"uniqueIndex;not null"`
	Name         string         `json:"name" gorm:"not null"`
	Description  string         `json:"description"`
	CategoryID   uuid.UUID      `json:"category_id" gorm:"type:uuid;not null"`
	Price        float64        `json:"price" gorm:"type:decimal(10,2);not null"`
	Cost         float64        `json:"cost" gorm:"type:decimal(10,2)"`
	Barcode      string         `json:"barcode" gorm:"uniqueIndex"`
	ImageURL     string         `json:"image_url"`
	Stock        int            `json:"stock" gorm:"default:0"`
	MinStock     int            `json:"min_stock" gorm:"default:0"`
	MaxStock     int            `json:"max_stock" gorm:"default:0"`
	Unit         string         `json:"unit" gorm:"default:'pcs'"`
	Weight       float64        `json:"weight" gorm:"type:decimal(8,3)"`
	Dimensions   string         `json:"dimensions"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	IsTaxable    bool           `json:"is_taxable" gorm:"default:true"`
	TaxRate      float64        `json:"tax_rate" gorm:"type:decimal(5,4);default:0"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Category    Category    `json:"category" gorm:"foreignKey:CategoryID"`
	OrderItems  []OrderItem `json:"order_items" gorm:"foreignKey:ProductID"`
}

// Order represents a sales transaction
type Order struct {
	ID              uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderNumber     string         `json:"order_number" gorm:"uniqueIndex;not null"`
	CustomerID      *uuid.UUID     `json:"customer_id" gorm:"type:uuid"`
	UserID          uuid.UUID      `json:"user_id" gorm:"type:uuid;not null"`
	Status          OrderStatus    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	PaymentStatus   PaymentStatus  `json:"payment_status" gorm:"type:varchar(20);not null;default:'pending'"`
	Subtotal        float64        `json:"subtotal" gorm:"type:decimal(10,2);not null"`
	TaxAmount       float64        `json:"tax_amount" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount  float64        `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	ShippingAmount  float64        `json:"shipping_amount" gorm:"type:decimal(10,2);default:0"`
	Total           float64        `json:"total" gorm:"type:decimal(10,2);not null"`
	Notes           string         `json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Customer   *Customer   `json:"customer" gorm:"foreignKey:CustomerID"`
	User       User        `json:"user" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
	Payments   []Payment   `json:"payments" gorm:"foreignKey:OrderID"`
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// PaymentStatus represents the payment status of an order
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusPaid       PaymentStatus = "paid"
	PaymentStatusPartial    PaymentStatus = "partial"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusRefunded   PaymentStatus = "refunded"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
)

// OrderItem represents individual items in an order
type OrderItem struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID    uuid.UUID      `json:"order_id" gorm:"type:uuid;not null"`
	ProductID  uuid.UUID      `json:"product_id" gorm:"type:uuid;not null"`
	Quantity   int            `json:"quantity" gorm:"not null"`
	UnitPrice  float64        `json:"unit_price" gorm:"type:decimal(10,2);not null"`
	TotalPrice float64        `json:"total_price" gorm:"type:decimal(10,2);not null"`
	Notes      string         `json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Order   Order   `json:"order" gorm:"foreignKey:OrderID"`
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}

// Payment represents payment transactions
type Payment struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID       uuid.UUID      `json:"order_id" gorm:"type:uuid;not null"`
	PaymentMethod PaymentMethod  `json:"payment_method" gorm:"type:varchar(20);not null"`
	Amount        float64        `json:"amount" gorm:"type:decimal(10,2);not null"`
	Status        PaymentStatus  `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	TransactionID string         `json:"transaction_id" gorm:"uniqueIndex"`
	Reference     string         `json:"reference"`
	Notes         string         `json:"notes"`
	ProcessedAt   *time.Time     `json:"processed_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relationships
	Order Order `json:"order" gorm:"foreignKey:OrderID"`
}

// PaymentMethod represents different payment methods
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodCard         PaymentMethod = "card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDigitalWallet PaymentMethod = "digital_wallet"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)