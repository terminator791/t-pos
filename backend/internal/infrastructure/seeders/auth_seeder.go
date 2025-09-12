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
		{"admin", "*", "/api/v1/products/*", "GET"},
		{"admin", "*", "/api/v1/products/*", "POST"},
		{"admin", "*", "/api/v1/products/*", "PUT"},
		{"admin", "*", "/api/v1/products/*", "DELETE"},

		{"admin", "*", "/api/v1/products", "GET"},
		{"admin", "*", "/api/v1/products", "POST"},
		{"admin", "*", "/api/v1/products", "PUT"},
		{"admin", "*", "/api/v1/products", "DELETE"},

		// Categories
		{"admin", "*", "/api/v1/categories", "GET"},
		{"admin", "*", "/api/v1/categories", "POST"},
		{"admin", "*", "/api/v1/categories", "PUT"},
		{"admin", "*", "/api/v1/categories", "DELETE`"},

		{"admin", "*", "/api/v1/categories/*", "GET"},
		{"admin", "*", "/api/v1/categories/*", "POST"},
		{"admin", "*", "/api/v1/categories/*", "PUT"},
		{"admin", "*", "/api/v1/categories/*", "DELETE"},

		// shops
		{"admin", "*", "/api/v1/shops", "GET"},
		{"admin", "*", "/api/v1/shops", "POST"},
		{"admin", "*", "/api/v1/shops", "PUT"},
		{"admin", "*", "/api/v1/shops", "DELETE"},

		{"admin", "*", "/api/v1/shops/*", "GET"},
		{"admin", "*", "/api/v1/shops/*", "POST"},
		{"admin", "*", "/api/v1/shops/*", "PUT"},
		{"admin", "*", "/api/v1/shops/*", "DELETE"},

		// Carts
		{"admin", "*", "/api/v1/carts", "GET"},
		{"admin", "*", "/api/v1/carts", "POST"},
		{"admin", "*", "/api/v1/carts", "PUT"},
		{"admin", "*", "/api/v1/carts", "DELETE"},

		{"admin", "*", "/api/v1/carts/*", "GET"},
		{"admin", "*", "/api/v1/carts/*", "POST"},
		{"admin", "*", "/api/v1/carts/*", "PUT"},
		{"admin", "*", "/api/v1/carts/*", "DELETE"},

		// Transactions
		{"admin", "*", "/api/v1/transactions", "GET"},
		{"admin", "*", "/api/v1/transactions", "POST"},
		{"admin", "*", "/api/v1/transactions", "PUT"},
		{"admin", "*", "/api/v1/transactions", "DELETE"},

		{"admin", "*", "/api/v1/transactions/*", "GET"},
		{"admin", "*", "/api/v1/transactions/*", "POST"},
		{"admin", "*", "/api/v1/transactions/*", "PUT"},
		{"admin", "*", "/api/v1/transactions/*", "DELETE"},

		// Expenses
		{"admin", "*", "/api/v1/expenses", "GET"},
		{"admin", "*", "/api/v1/expenses", "POST"},
		{"admin", "*", "/api/v1/expenses", "PUT"},
		{"admin", "*", "/api/v1/expenses", "DELETE"},

		{"admin", "*", "/api/v1/expenses/*", "GET"},
		{"admin", "*", "/api/v1/expenses/*", "POST"},
		{"admin", "*", "/api/v1/expenses/*", "PUT"},
		{"admin", "*", "/api/v1/expenses/*", "DELETE"},

		// Payments
		{"admin", "*", "/api/v1/payments", "GET"},
		{"admin", "*", "/api/v1/payments", "POST"},
		{"admin", "*", "/api/v1/payments", "PUT"},
		{"admin", "*", "/api/v1/payments", "DELETE"},

		{"admin", "*", "/api/v1/payments/*", "GET"},
		{"admin", "*", "/api/v1/payments/*", "POST"},
		{"admin", "*", "/api/v1/payments/*", "PUT"},
		{"admin", "*", "/api/v1/payments/*", "DELETE"},

		// Histories
		{"admin", "*", "/api/v1/histories", "GET"},
		{"admin", "*", "/api/v1/histories", "POST"},
		{"admin", "*", "/api/v1/histories", "PUT"},
		{"admin", "*", "/api/v1/histories", "DELETE"},

		{"admin", "*", "/api/v1/histories/*", "GET"},
		{"admin", "*", "/api/v1/histories/*", "POST"},
		{"admin", "*", "/api/v1/histories/*", "PUT"},
		{"admin", "*", "/api/v1/histories/*", "DELETE"},

		// Receipts
		{"admin", "*", "/api/v1/receipts", "GET"},
		{"admin", "*", "/api/v1/receipts", "POST"},
		{"admin", "*", "/api/v1/receipts", "PUT"},
		{"admin", "*", "/api/v1/receipts", "DELETE"},

		{"admin", "*", "/api/v1/receipts/*", "GET"},
		{"admin", "*", "/api/v1/receipts/*", "POST"},
		{"admin", "*", "/api/v1/receipts/*", "PUT"},
		{"admin", "*", "/api/v1/receipts/*", "DELETE"},

		// Transaction Products
		{"admin", "*", "/api/v1/transaction-products", "GET"},
		{"admin", "*", "/api/v1/transaction-products", "POST"},
		{"admin", "*", "/api/v1/transaction-products", "PUT"},
		{"admin", "*", "/api/v1/transaction-products", "DELETE"},

		{"admin", "*", "/api/v1/transaction-products/*", "GET"},
		{"admin", "*", "/api/v1/transaction-products/*", "POST"},
		{"admin", "*", "/api/v1/transaction-products/*", "PUT"},
		{"admin", "*", "/api/v1/transaction-products/*", "DELETE"},

		// Licenses
		{"admin", "*", "/api/v1/licenses", "GET"},
		{"admin", "*", "/api/v1/licenses", "POST"},
		{"admin", "*", "/api/v1/licenses", "PUT"},
		{"admin", "*", "/api/v1/licenses", "DELETE"},

		{"admin", "*", "/api/v1/licenses/*", "GET"},
		{"admin", "*", "/api/v1/licenses/*", "POST"},
		{"admin", "*", "/api/v1/licenses/*", "PUT"},
		{"admin", "*", "/api/v1/licenses/*", "DELETE"},

		// Customers
		{"admin", "*", "/api/v1/customers", "GET"},
		{"admin", "*", "/api/v1/customers", "POST"},
		{"admin", "*", "/api/v1/customers", "PUT"},
		{"admin", "*", "/api/v1/customers", "DELETE"},

		{"admin", "*", "/api/v1/customers/*", "GET"},
		{"admin", "*", "/api/v1/customers/*", "POST"},
		{"admin", "*", "/api/v1/customers/*", "PUT"},
		{"admin", "*", "/api/v1/customers/*", "DELETE"},

		// User Management

		// Roles

		// ACL Management

		// Sync
		{"admin", "*", "/api/v1/sync", "GET"},
		{"admin", "*", "/api/v1/sync", "POST"},
		{"admin", "*", "/api/v1/sync", "PUT"},
		{"admin", "*", "/api/v1/sync", "DELETE"},

		{"admin", "*", "/api/v1/sync/*", "GET"},
		{"admin", "*", "/api/v1/sync/*", "POST"},
		{"admin", "*", "/api/v1/sync/*", "PUT"},
		{"admin", "*", "/api/v1/sync/*", "DELETE"},

		// ================== OWNER BUSINESS - License-based shop management ==================
		// Products - Full CRUD within license domain
		// Products
		{"owner_business", "*", "/api/v1/products/*", "GET"},
		{"owner_business", "*", "/api/v1/products/*", "POST"},
		{"owner_business", "*", "/api/v1/products/*", "PUT"},
		{"owner_business", "*", "/api/v1/products/*", "DELETE"},

		{"owner_business", "*", "/api/v1/products", "GET"},
		{"owner_business", "*", "/api/v1/products", "POST"},
		{"owner_business", "*", "/api/v1/products", "PUT"},
		{"owner_business", "*", "/api/v1/products", "DELETE"},

		// Categories
		{"owner_business", "*", "/api/v1/categories", "GET"},
		{"owner_business", "*", "/api/v1/categories", "POST"},
		{"owner_business", "*", "/api/v1/categories", "PUT"},
		{"owner_business", "*", "/api/v1/categories", "DELETE"},

		{"owner_business", "*", "/api/v1/categories/*", "GET"},
		{"owner_business", "*", "/api/v1/categories/*", "POST"},
		{"owner_business", "*", "/api/v1/categories/*", "PUT"},
		{"owner_business", "*", "/api/v1/categories/*", "DELETE"},

		// shops
		{"owner_business", "*", "/api/v1/shops", "GET"},
		{"owner_business", "*", "/api/v1/shops", "POST"},
		{"owner_business", "*", "/api/v1/shops", "PUT"},
		{"owner_business", "*", "/api/v1/shops", "DELETE"},

		{"owner_business", "*", "/api/v1/shops/*", "GET"},
		{"owner_business", "*", "/api/v1/shops/*", "POST"},
		{"owner_business", "*", "/api/v1/shops/*", "PUT"},
		{"owner_business", "*", "/api/v1/shops/*", "DELETE"},

		// Carts
		{"owner_business", "*", "/api/v1/carts", "GET"},
		{"owner_business", "*", "/api/v1/carts", "POST"},
		{"owner_business", "*", "/api/v1/carts", "PUT"},
		{"owner_business", "*", "/api/v1/carts", "DELETE"},

		{"owner_business", "*", "/api/v1/carts/*", "GET"},
		{"owner_business", "*", "/api/v1/carts/*", "POST"},
		{"owner_business", "*", "/api/v1/carts/*", "PUT"},
		{"owner_business", "*", "/api/v1/carts/*", "DELETE"},

		// Transactions
		{"owner_business", "*", "/api/v1/transactions", "GET"},
		{"owner_business", "*", "/api/v1/transactions", "POST"},
		{"owner_business", "*", "/api/v1/transactions", "PUT"},
		{"owner_business", "*", "/api/v1/transactions", "DELETE"},

		{"owner_business", "*", "/api/v1/transactions/*", "GET"},
		{"owner_business", "*", "/api/v1/transactions/*", "POST"},
		{"owner_business", "*", "/api/v1/transactions/*", "PUT"},
		{"owner_business", "*", "/api/v1/transactions/*", "DELETE"},

		// Expenses
		{"owner_business", "*", "/api/v1/expenses", "GET"},
		{"owner_business", "*", "/api/v1/expenses", "POST"},
		{"owner_business", "*", "/api/v1/expenses", "PUT"},
		{"owner_business", "*", "/api/v1/expenses", "DELETE"},

		{"owner_business", "*", "/api/v1/expenses/*", "GET"},
		{"owner_business", "*", "/api/v1/expenses/*", "POST"},
		{"owner_business", "*", "/api/v1/expenses/*", "PUT"},
		{"owner_business", "*", "/api/v1/expenses/*", "DELETE"},

		// Payments
		{"owner_business", "*", "/api/v1/payments", "GET"},
		{"owner_business", "*", "/api/v1/payments", "POST"},
		{"owner_business", "*", "/api/v1/payments", "PUT"},
		{"owner_business", "*", "/api/v1/payments", "DELETE"},

		{"owner_business", "*", "/api/v1/payments/*", "GET"},
		{"owner_business", "*", "/api/v1/payments/*", "POST"},
		{"owner_business", "*", "/api/v1/payments/*", "PUT"},
		{"owner_business", "*", "/api/v1/payments/*", "DELETE"},

		// Histories
		{"owner_business", "*", "/api/v1/histories", "GET"},
		{"owner_business", "*", "/api/v1/histories", "POST"},
		{"owner_business", "*", "/api/v1/histories", "PUT"},
		{"owner_business", "*", "/api/v1/histories", "DELETE"},

		{"owner_business", "*", "/api/v1/histories/*", "GET"},
		{"owner_business", "*", "/api/v1/histories/*", "POST"},
		{"owner_business", "*", "/api/v1/histories/*", "PUT"},
		{"owner_business", "*", "/api/v1/histories/*", "DELETE"},

		// Receipts
		{"owner_business", "*", "/api/v1/receipts", "GET"},
		{"owner_business", "*", "/api/v1/receipts", "POST"},
		{"owner_business", "*", "/api/v1/receipts", "PUT"},
		{"owner_business", "*", "/api/v1/receipts", "DELETE"},

		{"owner_business", "*", "/api/v1/receipts/*", "GET"},
		{"owner_business", "*", "/api/v1/receipts/*", "POST"},
		{"owner_business", "*", "/api/v1/receipts/*", "PUT"},
		{"owner_business", "*", "/api/v1/receipts/*", "DELETE"},

		// Transaction Products
		{"owner_business", "*", "/api/v1/transaction-products", "GET"},
		{"owner_business", "*", "/api/v1/transaction-products", "POST"},
		{"owner_business", "*", "/api/v1/transaction-products", "PUT"},
		{"owner_business", "*", "/api/v1/transaction-products", "DELETE"},

		{"owner_business", "*", "/api/v1/transaction-products/*", "GET"},
		{"owner_business", "*", "/api/v1/transaction-products/*", "POST"},
		{"owner_business", "*", "/api/v1/transaction-products/*", "PUT"},
		{"owner_business", "*", "/api/v1/transaction-products/*", "DELETE"},

		// Licenses

		// Customers

		// User Management

		// Roles

		// ACL Management

		// Sync
		{"owner_business", "*", "/api/v1/sync", "GET"},
		{"owner_business", "*", "/api/v1/sync", "POST"},
		{"owner_business", "*", "/api/v1/sync", "PUT"},
		{"owner_business", "*", "/api/v1/sync", "DELETE"},

		{"owner_business", "*", "/api/v1/sync/*", "GET"},
		{"owner_business", "*", "/api/v1/sync/*", "POST"},
		{"owner_business", "*", "/api/v1/sync/*", "PUT"},
		{"owner_business", "*", "/api/v1/sync/*", "DELETE"},

		// ================== CASHIER - Shop-specific POS operations ==================
		// Products
		{"cashier", "*", "/api/v1/products/*", "GET"},
		{"cashier", "*", "/api/v1/products/*", "POST"},
		{"cashier", "*", "/api/v1/products/*", "PUT"},
		{"cashier", "*", "/api/v1/products/*", "DELETE"},

		{"cashier", "*", "/api/v1/products", "GET"},
		{"cashier", "*", "/api/v1/products", "POST"},
		{"cashier", "*", "/api/v1/products", "PUT"},
		{"cashier", "*", "/api/v1/products", "DELETE"},

		// Categories
		{"cashier", "*", "/api/v1/categories", "GET"},
		{"cashier", "*", "/api/v1/categories", "POST"},
		{"cashier", "*", "/api/v1/categories", "PUT"},
		{"cashier", "*", "/api/v1/categories", "DELETE"},

		{"cashier", "*", "/api/v1/categories/*", "GET"},
		{"cashier", "*", "/api/v1/categories/*", "POST"},
		{"cashier", "*", "/api/v1/categories/*", "PUT"},
		{"cashier", "*", "/api/v1/categories/*", "DELETE"},

		// shops
		{"cashier", "*", "/api/v1/shops", "GET"},
		{"cashier", "*", "/api/v1/shops", "POST"},
		{"cashier", "*", "/api/v1/shops", "PUT"},
		{"cashier", "*", "/api/v1/shops", "DELETE"},

		{"cashier", "*", "/api/v1/shops/*", "GET"},
		{"cashier", "*", "/api/v1/shops/*", "POST"},
		{"cashier", "*", "/api/v1/shops/*", "PUT"},
		{"cashier", "*", "/api/v1/shops/*", "DELETE"},

		// Carts
		{"cashier", "*", "/api/v1/carts", "GET"},
		{"cashier", "*", "/api/v1/carts", "POST"},
		{"cashier", "*", "/api/v1/carts", "PUT"},
		{"cashier", "*", "/api/v1/carts", "DELETE"},

		{"cashier", "*", "/api/v1/carts/*", "GET"},
		{"cashier", "*", "/api/v1/carts/*", "POST"},
		{"cashier", "*", "/api/v1/carts/*", "PUT"},
		{"cashier", "*", "/api/v1/carts/*", "DELETE"},

		// Transactions
		{"cashier", "*", "/api/v1/transactions", "GET"},
		{"cashier", "*", "/api/v1/transactions", "POST"},
		{"cashier", "*", "/api/v1/transactions", "PUT"},
		{"cashier", "*", "/api/v1/transactions", "DELETE"},

		{"cashier", "*", "/api/v1/transactions/*", "GET"},
		{"cashier", "*", "/api/v1/transactions/*", "POST"},
		{"cashier", "*", "/api/v1/transactions/*", "PUT"},
		{"cashier", "*", "/api/v1/transactions/*", "DELETE"},

		// Expenses
		{"cashier", "*", "/api/v1/expenses", "GET"},
		{"cashier", "*", "/api/v1/expenses", "POST"},
		{"cashier", "*", "/api/v1/expenses", "PUT"},
		{"cashier", "*", "/api/v1/expenses", "DELETE"},

		{"cashier", "*", "/api/v1/expenses/*", "GET"},
		{"cashier", "*", "/api/v1/expenses/*", "POST"},
		{"cashier", "*", "/api/v1/expenses/*", "PUT"},
		{"cashier", "*", "/api/v1/expenses/*", "DELETE"},

		// Payments
		{"cashier", "*", "/api/v1/payments", "GET"},
		{"cashier", "*", "/api/v1/payments", "POST"},
		{"cashier", "*", "/api/v1/payments", "PUT"},
		{"cashier", "*", "/api/v1/payments", "DELETE"},

		{"cashier", "*", "/api/v1/payments/*", "GET"},
		{"cashier", "*", "/api/v1/payments/*", "POST"},
		{"cashier", "*", "/api/v1/payments/*", "PUT"},
		{"cashier", "*", "/api/v1/payments/*", "DELETE"},

		// Histories
		{"cashier", "*", "/api/v1/histories", "GET"},
		{"cashier", "*", "/api/v1/histories", "POST"},
		{"cashier", "*", "/api/v1/histories", "PUT"},
		{"cashier", "*", "/api/v1/histories", "DELETE"},

		{"cashier", "*", "/api/v1/histories/*", "GET"},
		{"cashier", "*", "/api/v1/histories/*", "POST"},
		{"cashier", "*", "/api/v1/histories/*", "PUT"},
		{"cashier", "*", "/api/v1/histories/*", "DELETE"},

		// Receipts
		{"cashier", "*", "/api/v1/receipts", "GET"},
		{"cashier", "*", "/api/v1/receipts", "POST"},
		{"cashier", "*", "/api/v1/receipts", "PUT"},
		{"cashier", "*", "/api/v1/receipts", "DELETE"},

		{"cashier", "*", "/api/v1/receipts/*", "GET"},
		{"cashier", "*", "/api/v1/receipts/*", "POST"},
		{"cashier", "*", "/api/v1/receipts/*", "PUT"},
		{"cashier", "*", "/api/v1/receipts/*", "DELETE"},

		// Transaction Products
		{"cashier", "*", "/api/v1/transaction-products", "GET"},
		{"cashier", "*", "/api/v1/transaction-products", "POST"},
		{"cashier", "*", "/api/v1/transaction-products", "PUT"},
		{"cashier", "*", "/api/v1/transaction-products", "DELETE"},

		{"cashier", "*", "/api/v1/transaction-products/*", "GET"},
		{"cashier", "*", "/api/v1/transaction-products/*", "POST"},
		{"cashier", "*", "/api/v1/transaction-products/*", "PUT"},
		{"cashier", "*", "/api/v1/transaction-products/*", "DELETE"},

		// Licenses

		// Customers

		// User Management

		// Roles

		// ACL Management

		// Sync
		{"owner_business", "*", "/api/v1/sync", "GET"},
		{"owner_business", "*", "/api/v1/sync", "POST"},
		{"owner_business", "*", "/api/v1/sync", "PUT"},
		{"owner_business", "*", "/api/v1/sync", "DELETE"},

		{"owner_business", "*", "/api/v1/sync/*", "GET"},
		{"owner_business", "*", "/api/v1/sync/*", "POST"},
		{"owner_business", "*", "/api/v1/sync/*", "PUT"},
		{"owner_business", "*", "/api/v1/sync/*", "DELETE"},

		// ================== AUTH ENDPOINTS FOR ALL ROLES ==================

		{"admin", "*", "/api/v1/auth/*", "GET"},
		{"admin", "*", "/api/v1/auth/*", "POST"},
		{"admin", "*", "/api/v1/auth/*", "PUT"},
		{"admin", "*", "/api/v1/auth/*", "DELETE"},

		{"owner_business", "*", "/api/v1/auth/*", "GET"},
		{"owner_business", "*", "/api/v1/auth/*", "POST"},
		{"owner_business", "*", "/api/v1/auth/*", "PUT"},
		{"owner_business", "*", "/api/v1/auth/*", "DELETE"},


		{"cashier", "*", "/api/v1/auth/cashier", "GET"},
		{"cashier", "*", "/api/v1/auth/cashier", "POST"},
		{"cashier", "*", "/api/v1/auth/cashier", "PUT"},
		{"cashier", "*", "/api/v1/auth/cashier", "DELETE"},

		{"cashier", "*", "/api/v1/auth/cashier/*", "GET"},
		{"cashier", "*", "/api/v1/auth/cashier/*", "POST"},
		{"cashier", "*", "/api/v1/auth/cashier/*", "PUT"},
		{"cashier", "*", "/api/v1/auth/cashier/*", "DELETE"},

		{"cashier", "*", "/api/v1/auth/logout", "POST"},
		{"cashier", "*", "/api/v1/auth/refresh", "POST"},
		{"cashier", "*", "/api/v1/auth/profile", "GET"},
		{"cashier", "*", "/api/v1/auth/permissions", "GET"},

		{"cashier", "*", "/api/v1/auth/pin", "GET"},
		{"cashier", "*", "/api/v1/auth/pin", "POST"},
		{"cashier", "*", "/api/v1/auth/pin", "PUT"},
		{"cashier", "*", "/api/v1/auth/pin", "DELETE"},
		
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
