package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/pkg/response"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentRepo repositories.PaymentRepository
	roleRepo    repositories.RoleRepository
	shopRepo    repositories.ShopRepository
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentRepo repositories.PaymentRepository, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) *PaymentHandler {
	return &PaymentHandler{
		paymentRepo: paymentRepo,
		roleRepo:    roleRepo,
		shopRepo:    shopRepo,
	}
}

// ListPayments handles GET /payments - with domain-specific filtering
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	limit, offset := parsePaginationFromContext(c)

	// Get domain access info to apply filtering
	domainAccess, err := auth.GetUserDomainAccess(c, h.roleRepo, h.shopRepo)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to get user access info", err.Error())
		return
	}

	var payments []*entities.Payment

	// Apply domain-specific filtering
	if domainAccess.HasGlobalAccess {
		// Super admin and admin can see all payments
		payments, err = h.paymentRepo.List(c.Request.Context(), limit, offset)
	} else {
		// Filter by accessible shop IDs for tenant users
		shopFilter := domainAccess.GetShopFilter()
		if len(shopFilter) == 0 {
			// User has no accessible shops
			payments = []*entities.Payment{}
			err = nil
		} else {
			payments, err = h.paymentRepo.ListByShopIDs(c.Request.Context(), shopFilter, limit, offset)
		}
	}

	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve payments", err.Error())
		return
	}

	response.SuccessOK(c, "Payments retrieved successfully", map[string]interface{}{
		"payments": payments,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListPaymentsByShop handles GET /payments/shop/:shopId
func (h *PaymentHandler) ListPaymentsByShop(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	limit, offset := parsePaginationFromContext(c)

	payments, err := h.paymentRepo.GetByShopID(c.Request.Context(), shopID)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve payments", err.Error())
		return
	}

	response.SuccessOK(c, "Payments retrieved successfully", map[string]interface{}{
		"payments": payments,
		"shop_id":  shopID,
		"limit":    limit,
		"offset":   offset,
	})
}

// ListPaymentsByShopAndStatus handles GET /payments/shop/:shopId/status/:status
func (h *PaymentHandler) ListPaymentsByShopAndStatus(c *gin.Context) {
	shopID, err := uuid.Parse(c.Param("shopId"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid shop ID", err.Error())
		return
	}

	status := entities.PaymentStatus(c.Param("status"))
	limit, offset := parsePaginationFromContext(c)

	payments, err := h.paymentRepo.GetByShopIDAndStatus(c.Request.Context(), shopID, status)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve payments", err.Error())
		return
	}

	response.SuccessOK(c, "Payments retrieved successfully", map[string]interface{}{
		"payments": payments,
		"shop_id":  shopID,
		"status":   status,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetPayment handles GET /payments/:id
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	paymentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ErrorBadRequest(c, "Invalid payment ID", err.Error())
		return
	}

	payment, err := h.paymentRepo.GetByID(c.Request.Context(), paymentID)
	if err != nil {
		response.ErrorNotFound(c, "Payment not found", err.Error())
		return
	}

	response.SuccessOK(c, "Payment retrieved successfully", payment)
}
