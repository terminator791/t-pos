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
	expenseHandler *handlers.ExpenseHandler,
	paymentHandler *handlers.PaymentHandler,
	historyHandler *handlers.HistoryHandler,
	receiptHandler *handlers.ReceiptHandler,
	transactionProductHandler *handlers.TransactionProductHandler,
	aclHandler *handlers.ACLHandler,
	shopHandler *handlers.ShopHandler,
	syncHandler *handlers.SyncHandler,
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

			// Shop routes
			shops := protected.Group("/shops")
			{
				shops.POST("", shopHandler.CreateShop)
				shops.GET("", shopHandler.ListShops)
				shops.GET("/:id", shopHandler.GetShop)
				shops.PUT("/:id", shopHandler.UpdateShop)
				shops.DELETE("/:id", shopHandler.DeleteShop)
				shops.GET("/license/:licenseId", shopHandler.GetShopsByLicense)
				shops.GET("/owner/:ownerId", shopHandler.GetShopsByOwner)
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
				// List routes
				transactions.GET("", transactionHandler.ListTransactions) // super admin and admin only
				transactions.GET("/shop/:shopId", transactionHandler.ListTransactionsByShop)
				transactions.GET("/shop/:shopId/status/:status", transactionHandler.ListTransactionsByShopAndStatus)
				// Legacy routes
				transactions.GET("/shop/:shopId/today", checkoutHandler.GetTodaysTransactions)
			}

			// Expense routes
			expenses := protected.Group("/expenses")
			{
				expenses.GET("", expenseHandler.ListExpenses) // super admin and admin only
				expenses.GET("/shop/:shopId", expenseHandler.ListExpensesByShop)
				expenses.GET("/shop/:shopId/status/:status", expenseHandler.ListExpensesByShopAndStatus)
				expenses.GET("/:id", expenseHandler.GetExpense)
			}

			// Payment routes
			payments := protected.Group("/payments")
			{
				payments.GET("", paymentHandler.ListPayments) // super admin and admin only
				payments.GET("/shop/:shopId", paymentHandler.ListPaymentsByShop)
				payments.GET("/shop/:shopId/status/:status", paymentHandler.ListPaymentsByShopAndStatus)
				payments.GET("/:id", paymentHandler.GetPayment)
			}

			// History routes
			histories := protected.Group("/histories")
			{
				histories.GET("", historyHandler.ListHistories) // super admin and admin only
				histories.GET("/shop/:shopId", historyHandler.ListHistoriesByShop)
				histories.GET("/:id", historyHandler.GetHistory)
			}

			// Receipt routes
			receipts := protected.Group("/receipts")
			{
				receipts.GET("", receiptHandler.ListReceipts) // super admin and admin only
				receipts.GET("/shop/:shopId", receiptHandler.ListReceiptsByShop)
				receipts.GET("/:id", receiptHandler.GetReceipt)
			}

			// Transaction Product routes
			transactionProducts := protected.Group("/transaction-products")
			{
				transactionProducts.GET("", transactionProductHandler.ListTransactionProducts) // super admin and admin only
				transactionProducts.GET("/transaction/:transactionId", transactionProductHandler.ListTransactionProductsByTransaction)
				transactionProducts.GET("/shop/:shopId", transactionProductHandler.ListTransactionProductsByShop)
				transactionProducts.GET("/:id", transactionProductHandler.GetTransactionProduct)
			}

			// Checkout and transaction routes (existing, deprecated)
			// checkout := protected.Group("/checkout")
			// {
			// 	checkout.POST("", checkoutHandler.ProcessCheckout)
			// 	checkout.POST("/:transactionId/complete", checkoutHandler.CompletePayment)
			// 	checkout.POST("/:transactionId/cancel", checkoutHandler.CancelTransaction)
			// 	checkout.GET("/:transactionId", checkoutHandler.GetTransactionDetails)
			// }

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

			// ACL Management routes (super_admin only)
			acl := protected.Group("/acl")
			{
				// Policy management
				acl.GET("/policies", aclHandler.GetAllPolicies)
				acl.POST("/policies", aclHandler.AddPolicy)
				acl.DELETE("/policies", aclHandler.RemovePolicy)
				acl.GET("/policies/system", aclHandler.GetSystemPolicies)

				// Role assignment management
				acl.GET("/roles", aclHandler.GetAllRoles)
				acl.GET("/roles/system", aclHandler.GetSystemRoles)
				acl.GET("/users/:userId/roles", aclHandler.GetUserRoles)
				acl.GET("/roles/:role/users", aclHandler.GetRoleUsers)
				acl.POST("/users/roles", aclHandler.AddRoleForUser)
				acl.DELETE("/users/roles", aclHandler.RemoveRoleForUser)

				// Permission checking
				acl.POST("/check", aclHandler.CheckPermission)
				acl.POST("/reload", aclHandler.ReloadPolicies)
			}

			// Sync routes (owner_business and cashier roles)
			sync := protected.Group("/sync")
			{
				sync.POST("", syncHandler.ProcessSync)
				sync.GET("/info", syncHandler.GetSyncInfo)
				sync.GET("/health", syncHandler.Health)
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
