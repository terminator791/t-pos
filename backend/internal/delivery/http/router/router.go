package router

import (
	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/delivery/http/middleware"
	"github.com/terminator791/t-pos/internal/infrastructure/config"
	"github.com/terminator791/t-pos/internal/infrastructure/logger"
)

// Router holds the router dependencies
type Router struct {
	config *config.Config
	logger logger.Logger
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, logger logger.Logger) *Router {
	return &Router{
		config: cfg,
		logger: logger,
	}
}

// SetupRoutes sets up all the routes
func (r *Router) SetupRoutes() *gin.Engine {
	// Set Gin mode
	gin.SetMode(r.config.Server.Mode)

	// Create Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(r.logger))
	router.Use(middleware.CORS())

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
			"message": "t-pos API is running",
		})
	})

	// API version 1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Login endpoint - coming soon"})
			})
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Register endpoint - coming soon"})
			})
		}

		// Protected routes (require authentication)
		protected := v1.Group("/")
		protected.Use(middleware.Auth()) // Will implement this middleware
		{
			// Users
			users := protected.Group("/users")
			{
				users.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List users - coming soon"})
				})
				users.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Create user - coming soon"})
				})
				users.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get user - coming soon"})
				})
				users.PUT("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Update user - coming soon"})
				})
				users.DELETE("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Delete user - coming soon"})
				})
			}

			// Customers
			customers := protected.Group("/customers")
			{
				customers.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List customers - coming soon"})
				})
				customers.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Create customer - coming soon"})
				})
				customers.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get customer - coming soon"})
				})
				customers.PUT("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Update customer - coming soon"})
				})
				customers.DELETE("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Delete customer - coming soon"})
				})
			}

			// Categories
			categories := protected.Group("/categories")
			{
				categories.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List categories - coming soon"})
				})
				categories.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Create category - coming soon"})
				})
				categories.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get category - coming soon"})
				})
				categories.PUT("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Update category - coming soon"})
				})
				categories.DELETE("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Delete category - coming soon"})
				})
			}

			// Products
			products := protected.Group("/products")
			{
				products.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List products - coming soon"})
				})
				products.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Create product - coming soon"})
				})
				products.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get product - coming soon"})
				})
				products.PUT("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Update product - coming soon"})
				})
				products.DELETE("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Delete product - coming soon"})
				})
				products.GET("/search", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Search products - coming soon"})
				})
				products.GET("/low-stock", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Low stock products - coming soon"})
				})
			}

			// Orders
			orders := protected.Group("/orders")
			{
				orders.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List orders - coming soon"})
				})
				orders.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Create order - coming soon"})
				})
				orders.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get order - coming soon"})
				})
				orders.PUT("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Update order - coming soon"})
				})
				orders.DELETE("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Cancel order - coming soon"})
				})
				orders.POST("/:id/items", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Add order item - coming soon"})
				})
				orders.DELETE("/:id/items/:item_id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Remove order item - coming soon"})
				})
			}

			// Payments
			payments := protected.Group("/payments")
			{
				payments.GET("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "List payments - coming soon"})
				})
				payments.POST("", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Process payment - coming soon"})
				})
				payments.GET("/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Get payment - coming soon"})
				})
				payments.POST("/:id/refund", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Refund payment - coming soon"})
				})
			}

			// Analytics
			analytics := protected.Group("/analytics")
			{
				analytics.GET("/dashboard", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Dashboard stats - coming soon"})
				})
				analytics.GET("/sales", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Sales report - coming soon"})
				})
				analytics.GET("/products", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Product report - coming soon"})
				})
				analytics.GET("/customers", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "Customer report - coming soon"})
				})
			}
		}
	}

	return router
}