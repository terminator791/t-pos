package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
	"github.com/terminator791/t-pos/pkg/response"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerService *services.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// GetCustomer handles GET /api/v1/customers/:id
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid customer ID", err.Error())
		return
	}

	customer, err := h.customerService.GetCustomer(c.Request.Context(), id)
	if err != nil {
		response.ErrorNotFound(c, "Customer not found", err.Error())
		return
	}

	response.SuccessOK(c, "Customer retrieved successfully", customer)
}

// GetAllCustomers handles GET /api/v1/customers
func (h *CustomerHandler) GetAllCustomers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid limit parameter", err.Error())
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid offset parameter", err.Error())
		return
	}

	customers, err := h.customerService.GetAllCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorInternalServer(c, "Failed to retrieve customers", err.Error())
		return
	}

	data := gin.H{
		"customers": customers,
		"count":     len(customers),
		"limit":     limit,
		"offset":    offset,
	}
	response.SuccessOK(c, "Customers retrieved successfully", data)
}

// CreateCustomer handles POST /api/v1/customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req services.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorBadRequest(c, "Invalid request body", err.Error())
		return
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), req)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to create customer", err.Error())
		return
	}

	response.SuccessCreated(c, "Customer created successfully", customer)
}

// DeleteCustomer handles DELETE /api/v1/customers/:id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.ErrorBadRequest(c, "Invalid customer ID", err.Error())
		return
	}

	err = h.customerService.DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		response.ErrorBadRequest(c, "Failed to delete customer", err.Error())
		return
	}

	response.SuccessOK(c, "Customer deleted successfully", nil)
}