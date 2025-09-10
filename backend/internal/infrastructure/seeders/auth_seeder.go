package seeders

import (
	"context"
	"log"

	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"gorm.io/gorm"
)

// AuthSeeder handles seeding of initial authentication data
type AuthSeeder struct {
	roleRepo        repositories.RoleRepository
	policyRepo      repositories.PolicyRepository
	enforcerService *casbin.EnforcerService
}

// NewAuthSeeder creates a new auth seeder
func NewAuthSeeder(
	roleRepo repositories.RoleRepository,
	policyRepo repositories.PolicyRepository,
	enforcerService *casbin.EnforcerService,
) *AuthSeeder {
	return &AuthSeeder{
		roleRepo:        roleRepo,
		policyRepo:      policyRepo,
		enforcerService: enforcerService,
	}
}

// SeedRoles creates initial system roles
func (s *AuthSeeder) SeedRoles() error {
	ctx := context.Background()

	roles := []entities.Role{
		{
			Name:        "super_admin",
			DisplayName: "Super Admin",
			Description: strPtr("Full system access across all shops"),
			IsActive:    true,
		},
		{
			Name:        "admin",
			DisplayName: "Admin",
			Description: strPtr("Administrative access within a shop"),
			IsActive:    true,
		},
		{
			Name:        "owner_business",
			DisplayName: "Owner Business",
			Description: strPtr("Business owner access to shop operations and management"),
			IsActive:    true,
		},
		{
			Name:        "cashier",
			DisplayName: "Cashier",
			Description: strPtr("Point of sale operations within a shop"),
			IsActive:    true,
		},
	}

	for _, role := range roles {
		// Check if role already exists
		existing, err := s.roleRepo.GetByName(ctx, role.Name)
		if err == gorm.ErrRecordNotFound {
			// Create the role
			if err := s.roleRepo.Create(ctx, &role); err != nil {
				log.Printf("Failed to create role %s: %v", role.Name, err)
				return err
			}
			log.Printf("Created role: %s", role.Name)
		} else if err != nil {
			log.Printf("Error checking role %s: %v", role.Name, err)
			return err
		} else {
			log.Printf("Role %s already exists", existing.Name)
		}
	}

	return nil
}

// SeedPolicies creates initial system policies
func (s *AuthSeeder) SeedPolicies() error {
	ctx := context.Background()

	// Define comprehensive policies for different roles with shop-based domain
	policies := []struct {
		roleName string
		domain   string
		object   string
		action   string
	}{
		// ================== SUPER ADMIN - Full system access ==================
		{"super_admin", "*", "/api/v1/*", "GET"},
		{"super_admin", "*", "/api/v1/*", "POST"},
		{"super_admin", "*", "/api/v1/*", "PUT"},
		{"super_admin", "*", "/api/v1/*", "DELETE"},

		// ================== ADMIN - Multi-shop management ==================
		// Products
		{"admin", "*", "/api/v1/products", "GET"},
		{"admin", "*", "/api/v1/products/*", "GET"},
		{"admin", "*", "/api/v1/products", "POST"},
		{"admin", "*", "/api/v1/products/*", "PUT"},
		{"admin", "*", "/api/v1/products/*", "DELETE"},
		{"admin", "*", "/api/v1/products/search", "GET"},
		{"admin", "*", "/api/v1/products/barcode/*", "GET"},
		{"admin", "*", "/api/v1/products/low-stock", "GET"},
		{"admin", "*", "/api/v1/products/upload", "POST"},

		// Categories
		{"admin", "*", "/api/v1/categories", "GET"},
		{"admin", "*", "/api/v1/categories/*", "GET"},
		{"admin", "*", "/api/v1/categories", "POST"},
		{"admin", "*", "/api/v1/categories/*", "PUT"},
		{"admin", "*", "/api/v1/categories/*", "DELETE"},

		// Carts
		{"admin", "*", "/api/v1/carts", "GET"},
		{"admin", "*", "/api/v1/carts/*", "GET"},
		{"admin", "*", "/api/v1/carts", "POST"},
		{"admin", "*", "/api/v1/carts/*", "PUT"},
		{"admin", "*", "/api/v1/carts/*", "DELETE"},
		{"admin", "*", "/api/v1/carts/all", "GET"},

		// Transactions
		{"admin", "*", "/api/v1/transactions", "GET"},
		{"admin", "*", "/api/v1/transactions/*", "GET"},
		{"admin", "*", "/api/v1/transactions", "POST"},
		{"admin", "*", "/api/v1/transactions/*/pay", "POST"},
		{"admin", "*", "/api/v1/transactions/*/cancel", "POST"},
		{"admin", "*", "/api/v1/transactions/shop/*", "GET"},
		{"admin", "*", "/api/v1/transactions/shop/*/status/*", "GET"},
		{"admin", "*", "/api/v1/transactions/shop/*/today", "GET"},

		// Expenses
		{"admin", "*", "/api/v1/expenses", "GET"},
		{"admin", "*", "/api/v1/expenses/*", "GET"},
		{"admin", "*", "/api/v1/expenses/shop/*", "GET"},
		{"admin", "*", "/api/v1/expenses/shop/*/status/*", "GET"},

		// Payments
		{"admin", "*", "/api/v1/payments", "GET"},
		{"admin", "*", "/api/v1/payments/*", "GET"},
		{"admin", "*", "/api/v1/payments/shop/*", "GET"},
		{"admin", "*", "/api/v1/payments/shop/*/status/*", "GET"},

		// Histories
		{"admin", "*", "/api/v1/histories", "GET"},
		{"admin", "*", "/api/v1/histories/*", "GET"},
		{"admin", "*", "/api/v1/histories/shop/*", "GET"},

		// Receipts
		{"admin", "*", "/api/v1/receipts", "GET"},
		{"admin", "*", "/api/v1/receipts/*", "GET"},
		{"admin", "*", "/api/v1/receipts/shop/*", "GET"},

		// Transaction Products
		{"admin", "*", "/api/v1/transaction-products", "GET"},
		{"admin", "*", "/api/v1/transaction-products/*", "GET"},
		{"admin", "*", "/api/v1/transaction-products/transaction/*", "GET"},
		{"admin", "*", "/api/v1/transaction-products/shop/*", "GET"},

		// Customers
		{"admin", "*", "/api/v1/customers", "GET"},
		{"admin", "*", "/api/v1/customers/*", "GET"},
		{"admin", "*", "/api/v1/customers", "POST"},
		{"admin", "*", "/api/v1/customers/*", "DELETE"},

		// User Management
		{"admin", "*", "/api/v1/users", "GET"},
		{"admin", "*", "/api/v1/users/*", "GET"},
		{"admin", "*", "/api/v1/users/*", "PUT"},

		// Roles
		{"admin", "*", "/api/v1/roles", "GET"},
		{"admin", "*", "/api/v1/roles/*", "GET"},
		{"admin", "*", "/api/v1/roles/name/*", "GET"},

		// ================== OWNER BUSINESS - License-based shop management ==================
		// Products - Full CRUD within license domain
		{"owner_business", "*", "/api/v1/products", "GET"},
		{"owner_business", "*", "/api/v1/products/*", "GET"},
		{"owner_business", "*", "/api/v1/products", "POST"},
		{"owner_business", "*", "/api/v1/products/*", "PUT"},
		{"owner_business", "*", "/api/v1/products/*", "DELETE"},
		{"owner_business", "*", "/api/v1/products/search", "GET"},
		{"owner_business", "*", "/api/v1/products/barcode/*", "GET"},
		{"owner_business", "*", "/api/v1/products/low-stock", "GET"},
		{"owner_business", "*", "/api/v1/products/upload", "POST"},

		// Categories - Full CRUD
		{"owner_business", "*", "/api/v1/categories", "GET"},
		{"owner_business", "*", "/api/v1/categories/*", "GET"},
		{"owner_business", "*", "/api/v1/categories", "POST"},
		{"owner_business", "*", "/api/v1/categories/*", "PUT"},
		{"owner_business", "*", "/api/v1/categories/*", "DELETE"},

		// Carts - Full CRUD
		{"owner_business", "*", "/api/v1/carts", "GET"},
		{"owner_business", "*", "/api/v1/carts/*", "GET"},
		{"owner_business", "*", "/api/v1/carts", "POST"},
		{"owner_business", "*", "/api/v1/carts/*", "PUT"},
		{"owner_business", "*", "/api/v1/carts/*", "DELETE"},
		{"owner_business", "*", "/api/v1/carts/all", "GET"},

		// Transactions - Full CRUD
		{"owner_business", "*", "/api/v1/transactions", "GET"},
		{"owner_business", "*", "/api/v1/transactions/*", "GET"},
		{"owner_business", "*", "/api/v1/transactions", "POST"},
		{"owner_business", "*", "/api/v1/transactions/*/pay", "POST"},
		{"owner_business", "*", "/api/v1/transactions/*/cancel", "POST"},
		{"owner_business", "*", "/api/v1/transactions/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/transactions/shop/*/status/*", "GET"},
		{"owner_business", "*", "/api/v1/transactions/shop/*/today", "GET"},

		// Expenses - Read only within license
		{"owner_business", "*", "/api/v1/expenses/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/expenses/shop/*/status/*", "GET"},
		{"owner_business", "*", "/api/v1/expenses/*", "GET"},

		// Payments - Read only within license
		{"owner_business", "*", "/api/v1/payments/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/payments/shop/*/status/*", "GET"},
		{"owner_business", "*", "/api/v1/payments/*", "GET"},

		// Histories - Read only within license
		{"owner_business", "*", "/api/v1/histories/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/histories/*", "GET"},

		// Receipts - Read only within license
		{"owner_business", "*", "/api/v1/receipts/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/receipts/*", "GET"},

		// Transaction Products - Read only within license
		{"owner_business", "*", "/api/v1/transaction-products/transaction/*", "GET"},
		{"owner_business", "*", "/api/v1/transaction-products/shop/*", "GET"},
		{"owner_business", "*", "/api/v1/transaction-products/*", "GET"},

		// Customers - Create and Delete (not cashier)
		{"owner_business", "*", "/api/v1/customers", "GET"},
		{"owner_business", "*", "/api/v1/customers/*", "GET"},
		{"owner_business", "*", "/api/v1/customers", "POST"},
		{"owner_business", "*", "/api/v1/customers/*", "DELETE"},

		// ================== CASHIER - Shop-specific POS operations ==================
		// Products - Full CRUD within shop
		{"cashier", "shop:*", "/api/v1/products", "GET"},
		{"cashier", "shop:*", "/api/v1/products/*", "GET"},
		{"cashier", "shop:*", "/api/v1/products", "POST"},
		{"cashier", "shop:*", "/api/v1/products/*", "PUT"},
		{"cashier", "shop:*", "/api/v1/products/*", "DELETE"},
		{"cashier", "shop:*", "/api/v1/products/search", "GET"},
		{"cashier", "shop:*", "/api/v1/products/barcode/*", "GET"},
		{"cashier", "shop:*", "/api/v1/products/low-stock", "GET"},
		{"cashier", "shop:*", "/api/v1/products/upload", "POST"},

		// Categories - Full CRUD within shop
		{"cashier", "shop:*", "/api/v1/categories", "GET"},
		{"cashier", "shop:*", "/api/v1/categories/*", "GET"},
		{"cashier", "shop:*", "/api/v1/categories", "POST"},
		{"cashier", "shop:*", "/api/v1/categories/*", "PUT"},
		{"cashier", "shop:*", "/api/v1/categories/*", "DELETE"},

		// Carts - Full CRUD within shop
		{"cashier", "shop:*", "/api/v1/carts", "GET"},
		{"cashier", "shop:*", "/api/v1/carts/*", "GET"},
		{"cashier", "shop:*", "/api/v1/carts", "POST"},
		{"cashier", "shop:*", "/api/v1/carts/*", "PUT"},
		{"cashier", "shop:*", "/api/v1/carts/*", "DELETE"},
		{"cashier", "shop:*", "/api/v1/carts/all", "GET"},

		// Transactions - Full CRUD within shop
		{"cashier", "shop:*", "/api/v1/transactions", "POST"},
		{"cashier", "shop:*", "/api/v1/transactions/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transactions/*/pay", "POST"},
		{"cashier", "shop:*", "/api/v1/transactions/*/cancel", "POST"},
		{"cashier", "shop:*", "/api/v1/transactions/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transactions/shop/*/status/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transactions/shop/*/today", "GET"},

		// Expenses - Read only within shop
		{"cashier", "shop:*", "/api/v1/expenses/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/expenses/shop/*/status/*", "GET"},
		{"cashier", "shop:*", "/api/v1/expenses/*", "GET"},

		// Payments - Read only within shop
		{"cashier", "shop:*", "/api/v1/payments/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/payments/shop/*/status/*", "GET"},
		{"cashier", "shop:*", "/api/v1/payments/*", "GET"},

		// Histories - Read only within shop
		{"cashier", "shop:*", "/api/v1/histories/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/histories/*", "GET"},

		// Receipts - Read only within shop
		{"cashier", "shop:*", "/api/v1/receipts/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/receipts/*", "GET"},

		// Transaction Products - Read only within shop
		{"cashier", "shop:*", "/api/v1/transaction-products/transaction/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transaction-products/shop/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transaction-products/*", "GET"},

		// Customers - Read only (cannot create/delete)
		{"cashier", "shop:*", "/api/v1/customers", "GET"},
		{"cashier", "shop:*", "/api/v1/customers/*", "GET"},

		// ================== AUTH ENDPOINTS FOR ALL ROLES ==================
		{"super_admin", "*", "/api/v1/auth/profile", "GET"},
		{"super_admin", "*", "/api/v1/auth/logout", "POST"},
		{"super_admin", "*", "/api/v1/auth/refresh", "POST"},
		{"super_admin", "*", "/api/v1/auth/permissions", "GET"},
		{"super_admin", "*", "/api/v1/auth/pin", "POST"},
		{"super_admin", "*", "/api/v1/auth/pin", "PUT"},
		{"super_admin", "*", "/api/v1/auth/pin", "DELETE"},
		
		{"admin", "*", "/api/v1/auth/profile", "GET"},
		{"admin", "*", "/api/v1/auth/logout", "POST"},
		{"admin", "*", "/api/v1/auth/refresh", "POST"},
		{"admin", "*", "/api/v1/auth/permissions", "GET"},
		{"admin", "*", "/api/v1/auth/pin", "POST"},
		{"admin", "*", "/api/v1/auth/pin", "PUT"},
		{"admin", "*", "/api/v1/auth/pin", "DELETE"},
		
		{"owner_business", "*", "/api/v1/auth/profile", "GET"},
		{"owner_business", "*", "/api/v1/auth/logout", "POST"},
		{"owner_business", "*", "/api/v1/auth/refresh", "POST"},
		{"owner_business", "*", "/api/v1/auth/permissions", "GET"},
		{"owner_business", "*", "/api/v1/auth/pin", "POST"},
		{"owner_business", "*", "/api/v1/auth/pin", "PUT"},
		{"owner_business", "*", "/api/v1/auth/pin", "DELETE"},
		
		{"cashier", "shop:*", "/api/v1/auth/profile", "GET"},
		{"cashier", "shop:*", "/api/v1/auth/logout", "POST"},
		{"cashier", "shop:*", "/api/v1/auth/refresh", "POST"},
		{"cashier", "shop:*", "/api/v1/auth/permissions", "GET"},
		{"cashier", "shop:*", "/api/v1/auth/pin", "POST"},
		{"cashier", "shop:*", "/api/v1/auth/pin", "PUT"},
		{"cashier", "shop:*", "/api/v1/auth/pin", "DELETE"},
	}

	for _, p := range policies {
		// Get role by name
		role, err := s.roleRepo.GetByName(ctx, p.roleName)
		if err != nil {
			log.Printf("Failed to find role %s: %v", p.roleName, err)
			continue
		}

		// Create policy entity
		policy := &entities.Policy{
			RoleID:   &role.ID,
			Subject:  p.roleName,
			Domain:   p.domain,
			Object:   p.object,
			Action:   p.action,
			Effect:   "allow",
			IsActive: true,
		}

		// Create policy in database
		if err := s.policyRepo.Create(ctx, policy); err != nil {
			log.Printf("Failed to create policy for role %s: %v", p.roleName, err)
			continue
		}

		// Add policy to Casbin
		if _, err := s.enforcerService.AddPolicy(p.roleName, p.domain, p.object, p.action); err != nil {
			log.Printf("Failed to add Casbin policy for role %s: %v", p.roleName, err)
			continue
		}

		log.Printf("Created policy: %s -> %s %s %s", p.roleName, p.domain, p.object, p.action)
	}

	return nil
}

// SeedAll runs all auth seeders
func (s *AuthSeeder) SeedAll() error {
	log.Println("Starting authentication data seeding...")

	if err := s.SeedRoles(); err != nil {
		return err
	}

	if err := s.SeedPolicies(); err != nil {
		return err
	}

	// Save policies to ensure they are persisted
	if err := s.enforcerService.SavePolicy(); err != nil {
		log.Printf("Failed to save Casbin policies: %v", err)
		return err
	}

	log.Println("Authentication data seeding completed successfully")
	return nil
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}