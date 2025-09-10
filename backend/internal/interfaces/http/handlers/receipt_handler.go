package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/pkg/response"
)

// ReceiptHandler handles receipt-related HTTP requests
type ReceiptHandler struct {
	receiptRepo repositories.ReceiptRepository
}

// NewReceiptHandler creates a new ReceiptHandler
func NewReceiptHandler(receiptRepo repositories.ReceiptRepository) *ReceiptHandler {
	return &ReceiptHandler{
		receiptRepo: receiptRepo,
	}
}

// ListReceipts handles GET /receipts - super admin and admin only
func (h *ReceiptHandler) ListReceipts(c *gin.Context) {
	limit, offset := parsePaginationFromContext(c)

	receipts, err := h.receiptRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve receipts", err.Error())
		return
	}

	response.SuccessOK(c, "Receipts retrieved successfully", map[string]interface{}{
		"receipts": receipts,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListReceiptsByShop handles GET /receipts/shop/:shopId
func (h *ReceiptHandler) ListReceiptsByShop(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	limit, offset := parsePaginationFromContext(c)

	receipts, err := h.receiptRepo.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve receipts", err.Error())
		return
	}

	response.SuccessOK(c, "Receipts retrieved successfully", map[string]interface{}{
		"receipts": receipts,
		"shop_id":  shopID,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetReceipt handles GET /receipts/:id
func (h *ReceiptHandler) GetReceipt(c *gin.Context) {
	receiptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid receipt ID", err.Error())
		return
	}

	receipt, err := h.receiptRepo.GetByID(c.Request.Context(), receiptID)
	if err != nil {
		response.ErrorNotFound(c, "Receipt not found", err.Error())
		return
	}

	response.SuccessOK(c, "Receipt retrieved successfully", receipt)
}
