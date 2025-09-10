package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/pkg/response"
)

// ExpenseHandler handles expense-related HTTP requests
type ExpenseHandler struct {
	expenseRepo repositories.ExpenseRepository
}

// NewExpenseHandler creates a new ExpenseHandler
func NewExpenseHandler(expenseRepo repositories.ExpenseRepository) *ExpenseHandler {
	return &ExpenseHandler{
		expenseRepo: expenseRepo,
	}
}

// ListExpenses handles GET /expenses - super admin and admin only
func (h *ExpenseHandler) ListExpenses(c *gin.Context) {
	limit, offset := parsePaginationFromContext(c)

	expenses, err := h.expenseRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve expenses", err.Error())
		return
	}

	response.SuccessOK(c, "Expenses retrieved successfully", map[string]interface{}{
		"expenses": expenses,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListExpensesByShop handles GET /expenses/shop/:shopId
func (h *ExpenseHandler) ListExpensesByShop(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	limit, offset := parsePaginationFromContext(c)

	expenses, err := h.expenseRepo.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve expenses", err.Error())
		return
	}

	response.SuccessOK(c, "Expenses retrieved successfully", map[string]interface{}{
		"expenses": expenses,
		"shop_id":  shopID,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListExpensesByShopAndStatus handles GET /expenses/shop/:shopId/status/:status
func (h *ExpenseHandler) ListExpensesByShopAndStatus(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	status := entities.ExpenseStatus(c.Param("status"))
	limit, offset := parsePaginationFromContext(c)

	expenses, err := h.expenseRepo.GetByShopIDAndStatus(c.Request.Context(), shopID, status)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve expenses", err.Error())
		return
	}

	response.SuccessOK(c, "Expenses retrieved successfully", map[string]interface{}{
		"expenses": expenses,
		"shop_id":  shopID,
		"status":   status,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetExpense handles GET /expenses/:id
func (h *ExpenseHandler) GetExpense(c *gin.Context) {
	expenseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid expense ID", err.Error())
		return
	}

	expense, err := h.expenseRepo.GetByID(c.Request.Context(), expenseID)
	if err != nil {
		response.ErrorNotFound(c, "Expense not found", err.Error())
		return
	}

	response.SuccessOK(c, "Expense retrieved successfully", expense)
}

// parsePaginationFromContext parses limit and offset from query parameters
func parsePaginationFromContext(c *gin.Context) (int, int) {
	limit := 20 // default limit
	offset := 0 // default offset

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}
