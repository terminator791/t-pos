package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// ReceiptHandler handles receipt-related HTTP requests
type ReceiptHandler struct {
	receiptRepo repositories.ReceiptRepository
	roleRepo    repositories.RoleRepository
	shopRepo    repositories.ShopRepository
}

// NewReceiptHandler creates a new ReceiptHandler
func NewReceiptHandler(receiptRepo repositories.ReceiptRepository, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *ReceiptHandler {
	return &ReceiptHandler{
		receiptRepo: receiptRepo,
		roleRepo:    roleRepo,
		shopRepo:    shopRepo,
	}
}

// ListReceipts handles GET /receipts - with domain-specific filtering
func (h *ReceiptHandler) ListReceipts(c *gin.Context) {
	limit, offset := parsePagination(c)

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
		return
	}

	var receipts []*entities.Receipt

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all receipts
		receipts, err = h.receiptRepo.List(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		if len(shopFilter) == 0 {
			// User has no accessible shops
			receipts = []*entities.Receipt{}
			err = nil
		} else {
			receipts, err = h.receiptRepo.ListByShopIDs(c.Request.Context(), shopFilter, limit, offset)
		}
	}

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

	limit, offset := parsePagination(c)

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
