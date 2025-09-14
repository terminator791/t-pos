package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// TransactionProductHandler handles transaction product-related HTTP requests
type TransactionProductHandler struct {
	transactionProductRepo repositories.TransactionProductRepository
	roleRepo               repositories.RoleRepository
	shopRepo               repositories.ShopRepository
}

// NewTransactionProductHandler creates a new TransactionProductHandler
func NewTransactionProductHandler(transactionProductRepo repositories.TransactionProductRepository, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *TransactionProductHandler {
	return &TransactionProductHandler{
		transactionProductRepo: transactionProductRepo,
		roleRepo:               roleRepo,
		shopRepo:               shopRepo,
	}
}

// ListTransactionProducts handles GET /transaction-products - with domain-specific filtering
func (h *TransactionProductHandler) ListTransactionProducts(c *gin.Context) {
	limit, offset := parsePagination(c)

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
		return
	}

	var transactionProducts []*entities.TransactionProduct

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all transaction products
		transactionProducts, err = h.transactionProductRepo.List(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		if len(shopFilter) == 0 {
			// User has no accessible shops
			transactionProducts = []*entities.TransactionProduct{}
			err = nil
		} else {
			transactionProducts, err = h.transactionProductRepo.ListByShopIDs(c.Request.Context(), shopFilter, limit, offset)
		}
	}

	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve transaction products", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction products retrieved successfully", map[string]interface{}{
		"transaction_products": transactionProducts,
		"limit":                limit,
		"offset":               offset,
	})
}

// ListTransactionProductsByTransaction handles GET /transaction-products/transaction/:transactionId
func (h *TransactionProductHandler) ListTransactionProductsByTransaction(c *gin.Context) {
	transactionID, err := uuid.Parse(c.Param("transactionId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction ID", err.Error())
		return
	}

	limit, offset := parsePaginationFromContext(c)

	transactionProducts, err := h.transactionProductRepo.GetByTransactionID(c.Request.Context(), transactionID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve transaction products", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction products retrieved successfully", map[string]interface{}{
		"transaction_products": transactionProducts,
		"transaction_id":       transactionID,
		"limit":                limit,
		"offset":               offset,
	})
}

// ListTransactionProductsByShop handles GET /transaction-products/shop/:shopId
func (h *TransactionProductHandler) ListTransactionProductsByShop(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	limit, offset := parsePagination(c)

	transactionProducts, err := h.transactionProductRepo.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve transaction products", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction products retrieved successfully", map[string]interface{}{
		"transaction_products": transactionProducts,
		"shop_id":              shopID,
		"limit":                limit,
		"offset":               offset,
	})
}

// GetTransactionProduct handles GET /transaction-products/:id
func (h *TransactionProductHandler) GetTransactionProduct(c *gin.Context) {
	transactionProductID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid transaction product ID", err.Error())
		return
	}

	transactionProduct, err := h.transactionProductRepo.GetByID(c.Request.Context(), transactionProductID)
	if err != nil {
		response.ErrorNotFound(c, "Transaction product not found", err.Error())
		return
	}

	response.SuccessOK(c, "Transaction product retrieved successfully", transactionProduct)
}
