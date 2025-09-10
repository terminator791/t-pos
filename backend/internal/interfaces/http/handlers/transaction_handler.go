package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/usecases"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	transactionUseCase *usecases.TransactionUseCase
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(transactionUseCase *usecases.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{
		transactionUseCase: transactionUseCase,
	}
}

// CreateTransactionRequest represents the request structure for creating a transaction
type CreateTransactionRequest struct {
	ShopID       *uuid.UUID                    `json:"shop_id,omitempty"` // Optional for owner business, ignored for cashiers
	CustomerName string                        `json:"customer_name" binding:"required"`
	Items        []CreateTransactionItemRequest `json:"items" binding:"required"`
	Discount     float64                       `json:"discount,omitempty"`
}

// CreateTransactionItemRequest represents an item in the transaction
type CreateTransactionItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1"`
}

// PayTransactionRequest represents the request structure for paying a transaction
type PayTransactionRequest struct {
	Amount float64 `json:"amount" binding:"required,min=0"`
}

// CreateTransaction handles POST /transactions
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Get cashier ID from context (set by auth middleware)
	cashierIDValue, exists := c.Get("user_id")
	if !exists {
		response.ErrorUnauthorized(c, "User not authenticated", nil)
		return
	}

	cashierID, ok := cashierIDValue.(uuid.UUID)
	if !ok {
		response.ErrorUnauthorized(c, "Invalid user ID", nil)
		return
	}

	// Get shop ID from user context or request
	// For cashiers: must use their assigned shop_id from context
	// For owner business: can specify shop_id in request or use context
	var shopID uuid.UUID
	userShopID, hasShopContext := auth.GetUserShopIDFromContext(c)
	
	if hasShopContext && userShopID != nil {
		// User has shop context (cashier), use it
		shopID = *userShopID
	} else if req.ShopID != nil {
		// No shop context but shop_id provided in request (owner business)
		shopID = *req.ShopID
	} else {
		// No shop context and no shop_id in request
		response.ErrorBadRequest(c, "Shop ID required: cashiers must be assigned to a shop, owner business must specify shop_id in request", nil)
		return
	}

	// Validate and convert request items
	usecaseItems, err := convertToUseCaseItems(req.Items)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction items", err.Error())
		return
	}

	transaction, err := h.transactionUseCase.CreateTransaction(c.Request.Context(), &usecases.CreateTransactionRequest{
		ShopID:       shopID,
		CashierID:    cashierID,
		CustomerName: req.CustomerName,
		Items:        usecaseItems,
		Discount:     req.Discount,
	})
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create transaction", err.Error())
		return
	}

	response.SuccessCreated(c, "Transaction created successfully", transaction)
}

// PayTransaction handles POST /transactions/:id/pay
func (h *TransactionHandler) PayTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction ID", err.Error())
		return
	}

	var req PayTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	result, err := h.transactionUseCase.PayTransaction(c.Request.Context(), transactionID, req.Amount)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to process payment", err.Error())
		return
	}

	response.SuccessOK(c, "Payment processed successfully", result)
}

// CancelTransaction handles POST /transactions/:id/cancel
func (h *TransactionHandler) CancelTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction ID", err.Error())
		return
	}

	err = h.transactionUseCase.CancelTransaction(c.Request.Context(), transactionID)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to cancel transaction", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction cancelled successfully", nil)
}

// GetTransaction handles GET /transactions/:id
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction ID", err.Error())
		return
	}

	transaction, err := h.transactionUseCase.GetTransaction(c.Request.Context(), transactionID)
	if err != nil {
		response.ErrorNotFound(c, "Transaction not found", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction retrieved successfully", transaction)
}

// convertToUseCaseItems converts request items to usecase items
func convertToUseCaseItems(items []CreateTransactionItemRequest) ([]usecases.CreateTransactionItem, error) {
	result := make([]usecases.CreateTransactionItem, len(items))
	for i, item := range items {
		// Validate that ProductID is not zero UUID
		if item.ProductID == uuid.Nil {
			return nil, errors.New("invalid product ID: cannot be empty or zero UUID")
		}
		
		result[i] = usecases.CreateTransactionItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return result, nil
}