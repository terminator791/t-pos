package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// HistoryHandler handles history-related HTTP requests
type HistoryHandler struct {
	historyRepo repositories.HistoryRepository
	roleRepo    repositories.RoleRepository
	shopRepo    repositories.ShopRepository
}

// NewHistoryHandler creates a new HistoryHandler
func NewHistoryHandler(historyRepo repositories.HistoryRepository, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *HistoryHandler {
	return &HistoryHandler{
		historyRepo: historyRepo,
		roleRepo:    roleRepo,
		shopRepo:    shopRepo,
	}
}

// ListHistories handles GET /histories - with domain-specific filtering
func (h *HistoryHandler) ListHistories(c *gin.Context) {
	limit, offset := parsePaginationFromContext(c)

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
		return
	}

	var histories []*entities.History

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all histories
		histories, err = h.historyRepo.List(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		if len(shopFilter) == 0 {
			// User has no accessible shops
			histories = []*entities.History{}
			err = nil
		} else {
			histories, err = h.historyRepo.ListByShopIDs(c.Request.Context(), shopFilter, limit, offset)
		}
	}

	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve histories", err.Error())
		return
	}

	response.SuccessOK(c, "Histories retrieved successfully", map[string]interface{}{
		"histories": histories,
		"limit":     limit,
		"offset":    offset,
	})
}

// ListHistoriesByShop handles GET /histories/shop/:shopId
func (h *HistoryHandler) ListHistoriesByShop(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	limit, offset := parsePaginationFromContext(c)

	histories, err := h.historyRepo.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve histories", err.Error())
		return
	}

	response.SuccessOK(c, "Histories retrieved successfully", map[string]interface{}{
		"histories": histories,
		"shop_id":   shopID,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetHistory handles GET /histories/:id
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	historyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid history ID", err.Error())
		return
	}

	history, err := h.historyRepo.GetByID(c.Request.Context(), historyID)
	if err != nil {
		response.ErrorNotFound(c, "History not found", err.Error())
		return
	}

	response.SuccessOK(c, "History retrieved successfully", history)
}
