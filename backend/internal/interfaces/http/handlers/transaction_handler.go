package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/usecases"
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
	ShopID       uuid.UUID                     `json:"shop_id" binding:"required"`
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

	transaction, err := h.transactionUseCase.CreateTransaction(c.Request.Context(), &usecases.CreateTransactionRequest{
		ShopID:       req.ShopID,
		CashierID:    cashierID,
		CustomerName: req.CustomerName,
		Items:        convertToUseCaseItems(req.Items),
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
func convertToUseCaseItems(items []CreateTransactionItemRequest) []usecases.CreateTransactionItem {
	result := make([]usecases.CreateTransactionItem, len(items))
	for i, item := range items {
		result[i] = usecases.CreateTransactionItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return result
}