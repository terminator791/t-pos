package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/application/services"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid customer ID",
			"message": err.Error(),
		})
		return
	}

	customer, err := h.customerService.GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Customer not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    customer,
		"message": "Customer retrieved successfully",
	})
}

// GetAllCustomers handles GET /api/v1/customers
func (h *CustomerHandler) GetAllCustomers(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid limit parameter",
			"message": err.Error(),
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid offset parameter",
			"message": err.Error(),
		})
		return
	}

	customers, err := h.customerService.GetAllCustomers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve customers",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    customers,
		"count":   len(customers),
		"limit":   limit,
		"offset":  offset,
		"message": "Customers retrieved successfully",
	})
}

// CreateCustomer handles POST /api/v1/customers
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req services.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to create customer",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    customer,
		"message": "Customer created successfully",
	})
}

// DeleteCustomer handles DELETE /api/v1/customers/:id
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid customer ID",
			"message": err.Error(),
		})
		return
	}

	err = h.customerService.DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to delete customer",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Customer deleted successfully",
	})
}