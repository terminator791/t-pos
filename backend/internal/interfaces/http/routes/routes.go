package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/terminator791/t-pos/internal/infrastructure/auth"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"github.com/terminator791/t-pos/internal/interfaces/http/handlers"
)

// SetupRoutes sets up all the API routes
func SetupRoutes(
	router *gin.Engine,
	productHandler *handlers.ProductHandler,
	checkoutHandler *handlers.CheckoutHandler,
	authHandler *handlers.AuthHandler,
	licenseHandler *handlers.LicenseHandler,
	customerHandler *handlers.CustomerHandler,
	userManagementHandler *handlers.UserManagementHandler,
	roleHandler *handlers.RoleHandler,
	categoryHandler *handlers.CategoryHandler,
	cartHandler *handlers.CartHandler,
	transactionHandler *handlers.TransactionHandler,
	authMiddleware *auth.AuthMiddleware,
	authzMiddleware *casbin.AuthzMiddleware,
) {
	// API version 1
	v1 := router.Group("/api/v1")
	{
		// Public authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Protected authentication routes
		authProtected := v1.Group("/auth")
		authProtected.Use(authMiddleware.RequireAuth())
		{
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.POST("/refresh", authHandler.RefreshToken)
			authProtected.GET("/profile", authHandler.Profile)
			authProtected.GET("/permissions", authHandler.GetPermissions)
			authProtected.POST("/pin", authHandler.CreatePin)
			authProtected.PUT("/pin", authHandler.UpdatePin)
			authProtected.DELETE("/pin", authHandler.DeletePin)
		}

		// Protected API routes
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		protected.Use(authzMiddleware.RequirePermission())
		{
			// Product routes
			products := protected.Group("/products")
			{
				products.POST("", productHandler.CreateProduct)
				products.POST("/upload", productHandler.CreateProductWithFile)
				products.GET("", productHandler.ListProducts)
				products.GET("/search", productHandler.SearchProducts)
				products.GET("/low-stock", productHandler.GetLowStockProducts)
				products.GET("/:id", productHandler.GetProduct)
				products.GET("/barcode/:barcode", productHandler.GetProductByBarcode)
				products.PUT("/:id", productHandler.UpdateProduct)
				products.DELETE("/:id", productHandler.DeleteProduct)
			}

			// Category routes
			categories := protected.Group("/categories")
			{
				categories.POST("", categoryHandler.CreateCategory)
				categories.GET("", categoryHandler.ListCategories)
				categories.GET("/:id", categoryHandler.GetCategory)
				categories.PUT("/:id", categoryHandler.UpdateCategory)
				categories.DELETE("/:id", categoryHandler.DeleteCategory)
			}

			// Cart routes
			carts := protected.Group("/carts")
			{
				carts.POST("", cartHandler.AddToCart)
				carts.GET("", cartHandler.GetUserCart)
				carts.GET("/all", cartHandler.ListAllCarts)
				carts.GET("/:id", cartHandler.GetCartItem)
				carts.PUT("/:id", cartHandler.UpdateCartQuantity)
				carts.DELETE("/:id", cartHandler.RemoveFromCart)
				carts.DELETE("", cartHandler.ClearCart)
			}

			// Transaction routes (new flow)
			transactions := protected.Group("/transactions")
			{
				transactions.POST("", transactionHandler.CreateTransaction)
				transactions.GET("/:id", transactionHandler.GetTransaction)
				transactions.POST("/:id/pay", transactionHandler.PayTransaction)
				transactions.POST("/:id/cancel", transactionHandler.CancelTransaction)
				// Legacy routes
				transactions.GET("/shop/:shopId/today", checkoutHandler.GetTodaysTransactions)
			}

			// Checkout and transaction routes (existing)
			checkout := protected.Group("/checkout")
			{
				checkout.POST("", checkoutHandler.ProcessCheckout)
				checkout.POST("/:transactionId/complete", checkoutHandler.CompletePayment)
				checkout.POST("/:transactionId/cancel", checkoutHandler.CancelTransaction)
				checkout.GET("/:transactionId", checkoutHandler.GetTransactionDetails)
			}

			// License routes
			licenses := protected.Group("/licenses")
			{
				licenses.GET("", licenseHandler.GetAllLicenses)
				licenses.GET("/:id", licenseHandler.GetLicense)
				licenses.POST("", licenseHandler.CreateLicense)
				licenses.DELETE("/:id", licenseHandler.DeleteLicense)
			}

			// Customer routes (users with cashier/owner_business roles)
			customers := protected.Group("/customers")
			{
				customers.GET("", customerHandler.GetAllCustomers)
				customers.GET("/:id", customerHandler.GetCustomer)
				customers.POST("", customerHandler.CreateCustomer)
				customers.DELETE("/:id", customerHandler.DeleteCustomer)
			}

			// User management routes (users with admin/super_admin roles)
			users := protected.Group("/users")
			{
				users.GET("", userManagementHandler.GetAllUsers)
				users.GET("/:id", userManagementHandler.GetUser)
				users.POST("", userManagementHandler.CreateUser)
				users.PUT("/:id", userManagementHandler.UpdateUserPassword)
				users.DELETE("/:id", userManagementHandler.DeleteUser)
			}

			// Role routes
			roles := protected.Group("/roles")
			{
				roles.GET("", roleHandler.GetAllRoles)
				roles.GET("/:id", roleHandler.GetRole)
				roles.GET("/name/:name", roleHandler.GetRoleByName)
			}
		}
	}

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "T-POS API is running",
		})
	})
}