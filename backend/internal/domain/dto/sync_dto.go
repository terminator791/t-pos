package dto

import (
	"time"

	"github.com/google/uuid"
)

// SyncCartDTO represents a cart entity for sync responses without relations
type SyncCartDTO struct {
	ID        uuid.UUID `json:"id"`
	ShopID    uuid.UUID `json:"shop_id"`
	ProductID uuid.UUID `json:"product_id"`
	UserID    uuid.UUID `json:"user_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncCategoryDTO represents a category entity for sync responses without relations
type SyncCategoryDTO struct {
	ID        uuid.UUID `json:"id"`
	ShopID    uuid.UUID `json:"shop_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncProductDTO represents a product entity for sync responses without relations
type SyncProductDTO struct {
	ID          uuid.UUID  `json:"id"`
	ShopID      uuid.UUID  `json:"shop_id"`
	CatID       *uuid.UUID `json:"cat_id"`
	Photo       *string    `json:"photo"`
	Name        string     `json:"name"`
	Barcode     *string    `json:"barcode"`
	Unit        *string    `json:"unit"`
	PPN         *float64   `json:"ppn"`
	Sale        float64    `json:"sale"`
	Buy         float64    `json:"buy"`
	Profit      *float64   `json:"profit"`
	Stock       int        `json:"stock"`
	IsSchedule  bool       `json:"is_schedule"`
	Schedule    *string    `json:"schedule"`
	Qty         *int       `json:"qty"`
	IsHaveStock bool       `json:"is_have_stock"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// SyncExpenseDTO represents an expense entity for sync responses without relations
type SyncExpenseDTO struct {
	ID        uuid.UUID `json:"id"`
	ShopID    uuid.UUID `json:"shop_id"`
	Nominal   float64   `json:"nominal"`
	Status    string    `json:"status"`
	Date      time.Time `json:"date"`
	Label     *string   `json:"label"`
	Desc      *string   `json:"desc"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncPaymentDTO represents a payment entity for sync responses without relations
type SyncPaymentDTO struct {
	ID            uuid.UUID  `json:"id"`
	ShopID        uuid.UUID  `json:"shop_id"`
	UserID        *uuid.UUID `json:"user_id"`
	TransactionID uuid.UUID  `json:"transaction_id"`
	Status        string     `json:"status"`
	Total         float64    `json:"total"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SyncTransactionDTO represents a transaction entity for sync responses without relations
type SyncTransactionDTO struct {
	ID                   uuid.UUID  `json:"id"`
	ShopID               uuid.UUID  `json:"shop_id"`
	CashierID            uuid.UUID  `json:"cashier_id"`
	CustomerName         *string    `json:"customer_name"`
	Discount             float64    `json:"discount"`
	DiscountPercentage   float64    `json:"discount_percentage"`
	AdditionalCost       float64    `json:"additional_cost"`
	Status               string     `json:"status"`
	TotalPrice           float64    `json:"total_price"`
	ProfitTransaction    *float64   `json:"profit_transaction"`
	CashierName          *string    `json:"cashier_name"`
	Change               *float64   `json:"change"`
	Amount               int64      `json:"amount"`
	InitialPaymentStatus *string    `json:"initial_payment_status"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// SyncReceiptDTO represents a receipt entity for sync responses without relations
type SyncReceiptDTO struct {
	ID         uuid.UUID `json:"id"`
	ShopID     uuid.UUID `json:"shop_id"`
	PaymentsID uuid.UUID `json:"payments_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SyncHistoryDTO represents a history entity for sync responses without relations
type SyncHistoryDTO struct {
	ID            uuid.UUID `json:"id"`
	ShopID        uuid.UUID `json:"shop_id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncShopDTO represents a shop entity for sync responses without relations
type SyncShopDTO struct {
	ID               uuid.UUID `json:"id"`
	LicenseID        uuid.UUID `json:"license_id"`
	UserID           uuid.UUID `json:"user_id"`
	Name             string    `json:"name"`
	Domain           string    `json:"domain"`
	Photo            *string   `json:"photo"`
	Address          *string   `json:"address"`
	Slogan           *string   `json:"slogan"`
	ProfitCalculate  int64     `json:"profit_calculate"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// SyncStockHistoryDTO represents a stock history entity for sync responses without relations
type SyncStockHistoryDTO struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	Stock     int       `json:"stock"`
	LastStock int       `json:"last_stock"`
	StockedAt time.Time `json:"stocked_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncTransactionProductDTO represents a transaction product entity for sync responses without relations
type SyncTransactionProductDTO struct {
	ID            uuid.UUID `json:"id"`
	TransactionID uuid.UUID `json:"transaction_id"`
	ProductID     uuid.UUID `json:"product_id"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	TotalPrice    float64   `json:"total_price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SyncUserDTO represents a user entity for sync responses without relations
type SyncUserDTO struct {
	ID              uuid.UUID  `json:"id"`
	LicenseID       *uuid.UUID `json:"license_id"`
	RoleID          *uuid.UUID `json:"role_id"`
	ShopID          *uuid.UUID `json:"shop_id"`
	Email           *string    `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	Username        *string    `json:"username"`
	Name            string     `json:"name"`
	InfoDevice      *string    `json:"info_device"`
	FCMToken        *string    `json:"fcm_token"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}