package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/usecases"
)

// CheckoutHandler handles checkout-related HTTP requests
type CheckoutHandler struct {
	checkoutUseCase *usecases.CheckoutUseCase
}

// NewCheckoutHandler creates a new checkout handler
func NewCheckoutHandler(checkoutUseCase *usecases.CheckoutUseCase) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutUseCase: checkoutUseCase,
	}
}

// ProcessCheckout processes a checkout request
func (h *CheckoutHandler) ProcessCheckout(c *gin.Context) {
	var req usecases.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.checkoutUseCase.ProcessCheckout(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// CompletePayment completes a payment
func (h *CheckoutHandler) CompletePayment(c *gin.Context) {
	transactionIDParam := c.Param("transactionId")
	transactionID, err := uuid.Parse(transactionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	err = h.checkoutUseCase.CompletePayment(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment completed successfully",
	})
}

// CancelTransaction cancels a transaction
func (h *CheckoutHandler) CancelTransaction(c *gin.Context) {
	transactionIDParam := c.Param("transactionId")
	transactionID, err := uuid.Parse(transactionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	err = h.checkoutUseCase.CancelTransaction(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction cancelled successfully",
	})
}

// GetTransactionDetails gets transaction details
func (h *CheckoutHandler) GetTransactionDetails(c *gin.Context) {
	transactionIDParam := c.Param("transactionId")
	transactionID, err := uuid.Parse(transactionIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	transaction, err := h.checkoutUseCase.GetTransactionWithDetails(c.Request.Context(), transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// GetTodaysTransactions gets today's transactions for a shop
func (h *CheckoutHandler) GetTodaysTransactions(c *gin.Context) {
	shopIDParam := c.Param("shopId")
	shopID, err := uuid.Parse(shopIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid shop ID"})
		return
	}

	transactions, err := h.checkoutUseCase.GetTodaysTransactions(c.Request.Context(), shopID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions,
	})
}
