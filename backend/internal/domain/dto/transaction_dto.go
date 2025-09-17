package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
)

// TransactionDTO represents a transaction without circular dependencies
type TransactionDTO struct {
	ID                   uuid.UUID                  `json:"id"`
	ShopID               uuid.UUID                  `json:"shop_id"`
	CashierID            uuid.UUID                  `json:"cashier_id"`
	CustomerName         *string                    `json:"customer_name"`
	Discount             float64                    `json:"discount"`
	DiscountPercentage   float64                    `json:"discount_percentage"`
	AdditionalCost       float64                    `json:"additional_cost"`
	Status               entities.TransactionStatus `json:"status"`
	TotalPrice           float64                    `json:"total_price"`
	ProfitTransaction    *float64                   `json:"profit_transaction"`
	CashierName          *string                    `json:"cashier_name"`
	Change               *float64                   `json:"change"`
	Amount               int64                      `json:"amount"`
	InitialPaymentStatus *string                    `json:"initial_payment_status"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
	Shop                 *ShopSummaryDTO            `json:"shop,omitempty"`
	Cashier              *UserSummaryDTO            `json:"cashier,omitempty"`
}

// PaymentDTO represents a payment without circular dependencies
type PaymentDTO struct {
	ID            uuid.UUID              `json:"id"`
	ShopID        uuid.UUID              `json:"shop_id"`
	UserID        *uuid.UUID             `json:"user_id"`
	TransactionID uuid.UUID              `json:"transaction_id"`
	Status        entities.PaymentStatus `json:"status"`
	Total         float64                `json:"total"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Shop          *ShopSummaryDTO        `json:"shop,omitempty"`
	Transaction   *TransactionSummaryDTO `json:"transaction,omitempty"`
}

// ExpenseDTO represents an expense without circular dependencies
type ExpenseDTO struct {
	ID        uuid.UUID              `json:"id"`
	ShopID    uuid.UUID              `json:"shop_id"`
	Nominal   float64                `json:"nominal"`
	Status    entities.ExpenseStatus `json:"status"`
	Date      time.Time              `json:"date"`
	Label     *string                `json:"label"`
	Desc      *string                `json:"desc"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Shop      *ShopSummaryDTO        `json:"shop,omitempty"`
}

// ShopSummaryDTO represents a shop summary without deep relations
type ShopSummaryDTO struct {
	ID              uuid.UUID          `json:"id"`
	LicenseID       uuid.UUID          `json:"license_id"`
	UserID          uuid.UUID          `json:"user_id"`
	Name            string             `json:"name"`
	Domain          string             `json:"domain"`
	Photo           *string            `json:"photo"`
	Address         *string            `json:"address"`
	Slogan          *string            `json:"slogan"`
	ProfitCalculate int64              `json:"profit_calculate"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	License         *LicenseSummaryDTO `json:"license,omitempty"`
	Owner           *UserSummaryDTO    `json:"owner,omitempty"`
}

// UserSummaryDTO represents a user summary without deep relations
type UserSummaryDTO struct {
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

// LicenseSummaryDTO represents a license summary
type LicenseSummaryDTO struct {
	ID           uuid.UUID `json:"id"`
	SerialNumber string    `json:"serial_number"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TransactionSummaryDTO represents a transaction summary for use in other DTOs
type TransactionSummaryDTO struct {
	ID                   uuid.UUID                  `json:"id"`
	ShopID               uuid.UUID                  `json:"shop_id"`
	CashierID            uuid.UUID                  `json:"cashier_id"`
	CustomerName         *string                    `json:"customer_name"`
	Discount             float64                    `json:"discount"`
	DiscountPercentage   float64                    `json:"discount_percentage"`
	AdditionalCost       float64                    `json:"additional_cost"`
	Status               entities.TransactionStatus `json:"status"`
	TotalPrice           float64                    `json:"total_price"`
	ProfitTransaction    *float64                   `json:"profit_transaction"`
	CashierName          *string                    `json:"cashier_name"`
	Change               *float64                   `json:"change"`
	Amount               int64                      `json:"amount"`
	InitialPaymentStatus *string                    `json:"initial_payment_status"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
	Cashier              *UserSummaryDTO            `json:"cashier,omitempty"`
}

// CreateTransactionResponseDTO represents the response after creating a transaction
type CreateTransactionResponseDTO struct {
	Transaction *TransactionDTO `json:"transaction"`
	Payment     *PaymentDTO     `json:"payment"`
	Expense     *ExpenseDTO     `json:"expense"`
}

// PayTransactionResponseDTO represents the response after paying a transaction
type PayTransactionResponseDTO struct {
	Transaction *TransactionDTO `json:"transaction"`
	Payment     *PaymentDTO     `json:"payment"`
	Change      float64         `json:"change"`
	Success     bool            `json:"success"`
}

// TransactionListResponseDTO represents the response for listing transactions
type TransactionListResponseDTO struct {
	Transactions []*TransactionDTO `json:"transactions"`
	ShopID       *uuid.UUID        `json:"shop_id,omitempty"`
	Status       *string           `json:"status,omitempty"`
	Limit        int               `json:"limit"`
	Offset       int               `json:"offset"`
}

// ConvertToTransactionDTO converts an entity to DTO
func ConvertToTransactionDTO(t *entities.Transaction) *TransactionDTO {
	if t == nil {
		return nil
	}

	dto := &TransactionDTO{
		ID:                   t.ID,
		ShopID:               t.ShopID,
		CashierID:            t.CashierID,
		CustomerName:         t.CustomerName,
		Discount:             t.Discount,
		DiscountPercentage:   t.DiscountPercentage,
		AdditionalCost:       t.AdditionalCost,
		Status:               t.Status,
		TotalPrice:           t.TotalPrice,
		ProfitTransaction:    t.ProfitTransaction,
		CashierName:          t.CashierName,
		Change:               t.Change,
		Amount:               t.Amount,
		InitialPaymentStatus: t.InitialPaymentStatus,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}

	// Convert shop if present
	if t.Shop.ID != uuid.Nil {
		dto.Shop = ConvertToShopSummaryDTO(&t.Shop)
	}

	// Convert cashier if present
	if t.Cashier.ID != uuid.Nil {
		dto.Cashier = ConvertToUserSummaryDTO(&t.Cashier)
	}

	return dto
}

// ConvertToPaymentDTO converts a payment entity to DTO
func ConvertToPaymentDTO(p *entities.Payment) *PaymentDTO {
	if p == nil {
		return nil
	}

	dto := &PaymentDTO{
		ID:            p.ID,
		ShopID:        p.ShopID,
		UserID:        p.UserID,
		TransactionID: p.TransactionID,
		Status:        p.Status,
		Total:         p.Total,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}

	// Convert shop if present
	if p.Shop.ID != uuid.Nil {
		dto.Shop = ConvertToShopSummaryDTO(&p.Shop)
	}

	// Convert transaction summary if present (avoid circular dependency)
	if p.Transaction.ID != uuid.Nil {
		dto.Transaction = ConvertToTransactionSummaryDTO(&p.Transaction)
	}

	return dto
}

// ConvertToExpenseDTO converts an expense entity to DTO
func ConvertToExpenseDTO(e *entities.Expense) *ExpenseDTO {
	if e == nil {
		return nil
	}

	dto := &ExpenseDTO{
		ID:        e.ID,
		ShopID:    e.ShopID,
		Nominal:   e.Nominal,
		Status:    e.Status,
		Date:      e.Date,
		Label:     e.Label,
		Desc:      e.Desc,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	// Convert shop if present
	if e.Shop.ID != uuid.Nil {
		dto.Shop = ConvertToShopSummaryDTO(&e.Shop)
	}

	return dto
}

// ConvertToShopSummaryDTO converts a shop entity to summary DTO
func ConvertToShopSummaryDTO(s *entities.Shop) *ShopSummaryDTO {
	if s == nil || s.ID == uuid.Nil {
		return nil
	}

	dto := &ShopSummaryDTO{
		ID:              s.ID,
		LicenseID:       s.LicenseID,
		UserID:          s.UserID,
		Name:            s.Name,
		Domain:          s.Domain,
		Photo:           s.Photo,
		Address:         s.Address,
		Slogan:          s.Slogan,
		ProfitCalculate: s.ProfitCalculate,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}

	// Convert license if present
	if s.License.ID != uuid.Nil {
		dto.License = ConvertToLicenseSummaryDTO(&s.License)
	}

	// Convert owner if present
	if s.Owner.ID != uuid.Nil {
		dto.Owner = ConvertToUserSummaryDTO(&s.Owner)
	}

	return dto
}

// ConvertToUserSummaryDTO converts a user entity to summary DTO
func ConvertToUserSummaryDTO(u *entities.User) *UserSummaryDTO {
	if u == nil || u.ID == uuid.Nil {
		return nil
	}

	return &UserSummaryDTO{
		ID:              u.ID,
		LicenseID:       u.LicenseID,
		RoleID:          u.RoleID,
		ShopID:          u.ShopID,
		Email:           u.Email,
		EmailVerifiedAt: u.EmailVerifiedAt,
		Username:        u.Username,
		Name:            u.Name,
		InfoDevice:      u.InfoDevice,
		FCMToken:        u.FCMToken,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// ConvertToLicenseSummaryDTO converts a license entity to summary DTO
func ConvertToLicenseSummaryDTO(l *entities.License) *LicenseSummaryDTO {
	if l == nil || l.ID == uuid.Nil {
		return nil
	}

	return &LicenseSummaryDTO{
		ID:           l.ID,
		SerialNumber: l.SerialNumber,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
}

// ConvertToTransactionSummaryDTO converts a transaction entity to summary DTO (no shop to avoid circular dependency)
func ConvertToTransactionSummaryDTO(t *entities.Transaction) *TransactionSummaryDTO {
	if t == nil || t.ID == uuid.Nil {
		return nil
	}

	dto := &TransactionSummaryDTO{
		ID:                   t.ID,
		ShopID:               t.ShopID,
		CashierID:            t.CashierID,
		CustomerName:         t.CustomerName,
		Discount:             t.Discount,
		DiscountPercentage:   t.DiscountPercentage,
		AdditionalCost:       t.AdditionalCost,
		Status:               t.Status,
		TotalPrice:           t.TotalPrice,
		ProfitTransaction:    t.ProfitTransaction,
		CashierName:          t.CashierName,
		Change:               t.Change,
		Amount:               t.Amount,
		InitialPaymentStatus: t.InitialPaymentStatus,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}

	// Convert cashier if present
	if t.Cashier.ID != uuid.Nil {
		dto.Cashier = ConvertToUserSummaryDTO(&t.Cashier)
	}

	return dto
}
