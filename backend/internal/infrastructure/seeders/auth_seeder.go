package seeders

import (
	"context"
	"fmt"
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

	// Define base policies for super_admin and admin only
	// owner_business and cashier policies will be assigned dynamically when users are created
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
		{"admin", "*", "/api/v1/categories", "DELETE"},
		{"admin", "*", "/api/v1/categories/*", "GET"},
		{"admin", "*", "/api/v1/categories/*", "POST"},
		{"admin", "*", "/api/v1/categories/*", "PUT"},
		{"admin", "*", "/api/v1/categories/*", "DELETE"},

		// Shops
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
		{"admin", "*", "/api/v1/users", "GET"},
		{"admin", "*", "/api/v1/users/*", "GET"},
		{"admin", "*", "/api/v1/users/*", "PUT"},

		// Roles
		{"admin", "*", "/api/v1/roles", "GET"},
		{"admin", "*", "/api/v1/roles/*", "GET"},
		{"admin", "*", "/api/v1/roles/name/*", "GET"},

		// Sync
		{"admin", "*", "/api/v1/sync", "GET"},
		{"admin", "*", "/api/v1/sync", "POST"},
		{"admin", "*", "/api/v1/sync", "PUT"},
		{"admin", "*", "/api/v1/sync", "DELETE"},
		{"admin", "*", "/api/v1/sync/*", "GET"},
		{"admin", "*", "/api/v1/sync/*", "POST"},
		{"admin", "*", "/api/v1/sync/*", "PUT"},
		{"admin", "*", "/api/v1/sync/*", "DELETE"},

		// ================== AUTH ENDPOINTS FOR ALL ROLES ==================
		{"admin", "*", "/api/v1/auth/*", "GET"},
		{"admin", "*", "/api/v1/auth/*", "POST"},
		{"admin", "*", "/api/v1/auth/*", "PUT"},
		{"admin", "*", "/api/v1/auth/*", "DELETE"},

		// NOTE: owner_business and cashier policies will be assigned dynamically
		// in auth handlers with specific domains (license.SerialNumber for owner_business, 
		// shop.Domain for cashier) to enforce proper multi-tenancy isolation
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

// AssignPoliciesForRole assigns policies for a specific role and domain dynamically
func (s *AuthSeeder) AssignPoliciesForRole(roleName, domain string) error {
	ctx := context.Background()

	// Define role-specific policies
	var policies []struct {
		roleName string
		domain   string
		object   string
		action   string
	}

	switch roleName {
	case "owner_business":
		// Owner business gets full access to all endpoints within their license domain
		policies = []struct {
			roleName string
			domain   string
			object   string
			action   string
		}{
			// Products
			{roleName, domain, "/api/v1/products", "GET"},
			{roleName, domain, "/api/v1/products", "POST"},
			{roleName, domain, "/api/v1/products", "PUT"},
			{roleName, domain, "/api/v1/products", "DELETE"},
			{roleName, domain, "/api/v1/products/*", "GET"},
			{roleName, domain, "/api/v1/products/*", "POST"},
			{roleName, domain, "/api/v1/products/*", "PUT"},
			{roleName, domain, "/api/v1/products/*", "DELETE"},

			// Categories
			{roleName, domain, "/api/v1/categories", "GET"},
			{roleName, domain, "/api/v1/categories", "POST"},
			{roleName, domain, "/api/v1/categories", "PUT"},
			{roleName, domain, "/api/v1/categories", "DELETE"},
			{roleName, domain, "/api/v1/categories/*", "GET"},
			{roleName, domain, "/api/v1/categories/*", "POST"},
			{roleName, domain, "/api/v1/categories/*", "PUT"},
			{roleName, domain, "/api/v1/categories/*", "DELETE"},

			// Shops
			{roleName, domain, "/api/v1/shops", "GET"},
			{roleName, domain, "/api/v1/shops", "POST"},
			{roleName, domain, "/api/v1/shops", "PUT"},
			{roleName, domain, "/api/v1/shops", "DELETE"},
			{roleName, domain, "/api/v1/shops/*", "GET"},
			{roleName, domain, "/api/v1/shops/*", "POST"},
			{roleName, domain, "/api/v1/shops/*", "PUT"},
			{roleName, domain, "/api/v1/shops/*", "DELETE"},
			// Shop license-specific endpoints
			{roleName, domain, "/api/v1/shops/license/*", "GET"},
			{roleName, domain, "/api/v1/shops/license/*", "POST"},
			{roleName, domain, "/api/v1/shops/license/*", "PUT"},
			{roleName, domain, "/api/v1/shops/license/*", "DELETE"},

			// Transactions (both individual and shop-specific)
			{roleName, domain, "/api/v1/transactions", "GET"},
			{roleName, domain, "/api/v1/transactions", "POST"},
			{roleName, domain, "/api/v1/transactions", "PUT"},
			{roleName, domain, "/api/v1/transactions", "DELETE"},
			{roleName, domain, "/api/v1/transactions/*", "GET"},
			{roleName, domain, "/api/v1/transactions/*", "POST"},
			{roleName, domain, "/api/v1/transactions/*", "PUT"},
			{roleName, domain, "/api/v1/transactions/*", "DELETE"},
			// Shop-specific transaction endpoints
			{roleName, domain, "/api/v1/transactions/shop/*", "GET"},
			{roleName, domain, "/api/v1/transactions/shop/*", "POST"},
			{roleName, domain, "/api/v1/transactions/shop/*", "PUT"},
			{roleName, domain, "/api/v1/transactions/shop/*", "DELETE"},

			// Carts
			{roleName, domain, "/api/v1/carts", "GET"},
			{roleName, domain, "/api/v1/carts", "POST"},
			{roleName, domain, "/api/v1/carts", "PUT"},
			{roleName, domain, "/api/v1/carts", "DELETE"},
			{roleName, domain, "/api/v1/carts/*", "GET"},
			{roleName, domain, "/api/v1/carts/*", "POST"},
			{roleName, domain, "/api/v1/carts/*", "PUT"},
			{roleName, domain, "/api/v1/carts/*", "DELETE"},

			// Expenses
			{roleName, domain, "/api/v1/expenses", "GET"},
			{roleName, domain, "/api/v1/expenses", "POST"},
			{roleName, domain, "/api/v1/expenses", "PUT"},
			{roleName, domain, "/api/v1/expenses", "DELETE"},
			{roleName, domain, "/api/v1/expenses/*", "GET"},
			{roleName, domain, "/api/v1/expenses/*", "POST"},
			{roleName, domain, "/api/v1/expenses/*", "PUT"},
			{roleName, domain, "/api/v1/expenses/*", "DELETE"},
			// Shop-specific expense endpoints
			{roleName, domain, "/api/v1/expenses/shop/*", "GET"},
			{roleName, domain, "/api/v1/expenses/shop/*", "POST"},
			{roleName, domain, "/api/v1/expenses/shop/*", "PUT"},
			{roleName, domain, "/api/v1/expenses/shop/*", "DELETE"},

			// Payments
			{roleName, domain, "/api/v1/payments", "GET"},
			{roleName, domain, "/api/v1/payments", "POST"},
			{roleName, domain, "/api/v1/payments", "PUT"},
			{roleName, domain, "/api/v1/payments", "DELETE"},
			{roleName, domain, "/api/v1/payments/*", "GET"},
			{roleName, domain, "/api/v1/payments/*", "POST"},
			{roleName, domain, "/api/v1/payments/*", "PUT"},
			{roleName, domain, "/api/v1/payments/*", "DELETE"},
			// Shop-specific payment endpoints
			{roleName, domain, "/api/v1/payments/shop/*", "GET"},
			{roleName, domain, "/api/v1/payments/shop/*", "POST"},
			{roleName, domain, "/api/v1/payments/shop/*", "PUT"},
			{roleName, domain, "/api/v1/payments/shop/*", "DELETE"},

			// Histories
			{roleName, domain, "/api/v1/histories", "GET"},
			{roleName, domain, "/api/v1/histories/*", "GET"},
			// Shop-specific history endpoints
			{roleName, domain, "/api/v1/histories/shop/*", "GET"},

			// Receipts
			{roleName, domain, "/api/v1/receipts", "GET"},
			{roleName, domain, "/api/v1/receipts/*", "GET"},
			// Shop-specific receipt endpoints
			{roleName, domain, "/api/v1/receipts/shop/*", "GET"},
			{roleName, domain, "/api/v1/receipts/shop/*", "POST"},

			// Transaction Products
			{roleName, domain, "/api/v1/transaction-products", "GET"},
			{roleName, domain, "/api/v1/transaction-products/*", "GET"},
			// Shop-specific transaction product endpoints
			{roleName, domain, "/api/v1/transaction-products/shop/*", "GET"},

			// Customers
			{roleName, domain, "/api/v1/customers", "GET"},
			{roleName, domain, "/api/v1/customers", "POST"},
			{roleName, domain, "/api/v1/customers", "PUT"},
			{roleName, domain, "/api/v1/customers", "DELETE"},
			{roleName, domain, "/api/v1/customers/*", "GET"},
			{roleName, domain, "/api/v1/customers/*", "POST"},
			{roleName, domain, "/api/v1/customers/*", "PUT"},
			{roleName, domain, "/api/v1/customers/*", "DELETE"},

			// Auth endpoints
			{roleName, domain, "/api/v1/auth/*", "GET"},
			{roleName, domain, "/api/v1/auth/*", "POST"},
			{roleName, domain, "/api/v1/auth/*", "PUT"},
			{roleName, domain, "/api/v1/auth/*", "DELETE"},

			// Sync
			{roleName, domain, "/api/v1/sync", "GET"},
			{roleName, domain, "/api/v1/sync", "POST"},
			{roleName, domain, "/api/v1/sync", "PUT"},
			{roleName, domain, "/api/v1/sync", "DELETE"},
			{roleName, domain, "/api/v1/sync/*", "GET"},
			{roleName, domain, "/api/v1/sync/*", "POST"},
			{roleName, domain, "/api/v1/sync/*", "PUT"},
			{roleName, domain, "/api/v1/sync/*", "DELETE"},
		}

	case "cashier":
		// Cashier gets POS operations access within their shop domain
		policies = []struct {
			roleName string
			domain   string
			object   string
			action   string
		}{
			// Products - Read only for POS
			{roleName, domain, "/api/v1/products", "GET"},
			{roleName, domain, "/api/v1/products/*", "GET"},

			// Categories - Read only
			{roleName, domain, "/api/v1/categories", "GET"},
			{roleName, domain, "/api/v1/categories/*", "GET"},

			// Shops - Read only for their shop
			{roleName, domain, "/api/v1/shops", "GET"},
			{roleName, domain, "/api/v1/shops/*", "GET"},

			// Transactions - Full CRUD for their shop (both individual and shop-specific endpoints)
			{roleName, domain, "/api/v1/transactions", "GET"},
			{roleName, domain, "/api/v1/transactions", "POST"},
			{roleName, domain, "/api/v1/transactions/*", "GET"},
			{roleName, domain, "/api/v1/transactions/*", "POST"},
			{roleName, domain, "/api/v1/transactions/*", "PUT"},
			{roleName, domain, "/api/v1/transactions/*", "DELETE"},
			// Shop-specific transaction endpoints
			{roleName, domain, "/api/v1/transactions/shop/*", "GET"},
			{roleName, domain, "/api/v1/transactions/shop/*", "POST"},
			{roleName, domain, "/api/v1/transactions/shop/*", "PUT"},
			{roleName, domain, "/api/v1/transactions/shop/*", "DELETE"},

			// Carts - Full CRUD for their shop
			{roleName, domain, "/api/v1/carts", "GET"},
			{roleName, domain, "/api/v1/carts", "POST"},
			{roleName, domain, "/api/v1/carts", "PUT"},
			{roleName, domain, "/api/v1/carts", "DELETE"},
			{roleName, domain, "/api/v1/carts/*", "GET"},
			{roleName, domain, "/api/v1/carts/*", "POST"},
			{roleName, domain, "/api/v1/carts/*", "PUT"},
			{roleName, domain, "/api/v1/carts/*", "DELETE"},

			// Expenses - Shop-specific access
			{roleName, domain, "/api/v1/expenses/*", "GET"},
			{roleName, domain, "/api/v1/expenses/*", "POST"},
			{roleName, domain, "/api/v1/expenses/shop/*", "GET"},
			{roleName, domain, "/api/v1/expenses/shop/*", "POST"},

			// Payments - Can create payments for their transactions
			{roleName, domain, "/api/v1/payments", "GET"},
			{roleName, domain, "/api/v1/payments", "POST"},
			{roleName, domain, "/api/v1/payments/*", "GET"},
			{roleName, domain, "/api/v1/payments/*", "POST"},
			// Shop-specific payment endpoints
			{roleName, domain, "/api/v1/payments/shop/*", "GET"},
			{roleName, domain, "/api/v1/payments/shop/*", "POST"},

			// Histories - Shop-specific access
			{roleName, domain, "/api/v1/histories/*", "GET"},
			{roleName, domain, "/api/v1/histories/shop/*", "GET"},

			// Receipts - Can generate receipts for their transactions
			{roleName, domain, "/api/v1/receipts", "GET"},
			{roleName, domain, "/api/v1/receipts", "POST"},
			{roleName, domain, "/api/v1/receipts/*", "GET"},
			{roleName, domain, "/api/v1/receipts/*", "POST"},
			// Shop-specific receipt endpoints
			{roleName, domain, "/api/v1/receipts/shop/*", "GET"},
			{roleName, domain, "/api/v1/receipts/shop/*", "POST"},

			// Transaction Products - Can access for their transactions
			{roleName, domain, "/api/v1/transaction-products", "GET"},
			{roleName, domain, "/api/v1/transaction-products", "POST"},
			{roleName, domain, "/api/v1/transaction-products/*", "GET"},
			{roleName, domain, "/api/v1/transaction-products/*", "POST"},
			// Shop-specific transaction product endpoints
			{roleName, domain, "/api/v1/transaction-products/shop/*", "GET"},
			{roleName, domain, "/api/v1/transaction-products/shop/*", "POST"},

			// Customers - Read only
			{roleName, domain, "/api/v1/customers", "GET"},
			{roleName, domain, "/api/v1/customers/*", "GET"},

			// Auth endpoints for cashier
			{roleName, domain, "/api/v1/auth/cashier", "GET"},
			{roleName, domain, "/api/v1/auth/cashier", "POST"},
			{roleName, domain, "/api/v1/auth/cashier/*", "GET"},
			{roleName, domain, "/api/v1/auth/cashier/*", "POST"},
			{roleName, domain, "/api/v1/auth/logout", "POST"},
			{roleName, domain, "/api/v1/auth/refresh", "POST"},
			{roleName, domain, "/api/v1/auth/profile", "GET"},
			{roleName, domain, "/api/v1/auth/permissions", "GET"},
			{roleName, domain, "/api/v1/auth/pin", "GET"},
			{roleName, domain, "/api/v1/auth/pin", "POST"},
			{roleName, domain, "/api/v1/auth/pin", "PUT"},
			{roleName, domain, "/api/v1/auth/pin", "DELETE"},
		}

	default:
		return fmt.Errorf("unsupported role for dynamic policy assignment: %s", roleName)
	}

	// Get role entity
	role, err := s.roleRepo.GetByName(ctx, roleName)
	if err != nil {
		return fmt.Errorf("failed to find role %s: %w", roleName, err)
	}

	// Check if policies already exist for this role and domain to avoid duplicates
	existingPolicies, err := s.policyRepo.GetByRoleAndDomain(ctx, role.ID, domain)
	if err != nil {
		log.Printf("Warning: Failed to check existing policies for role %s, domain %s: %v", roleName, domain, err)
	} else if len(existingPolicies) > 0 {
		log.Printf("Policies already exist for role %s, domain %s (%d existing). Skipping creation.", roleName, domain, len(existingPolicies))
		return nil
	}

	// Create and assign policies
	var policyEntities []*entities.Policy
	var casbinPolicies [][]string

	// Prepare all policies for bulk operations
	for _, p := range policies {
		// Create policy entity for database
		policy := &entities.Policy{
			RoleID:   &role.ID,
			Subject:  p.roleName,
			Domain:   p.domain,
			Object:   p.object,
			Action:   p.action,
			Effect:   "allow",
			IsActive: true,
		}
		policyEntities = append(policyEntities, policy)

		// Create Casbin policy format [role, domain, object, action]
		casbinPolicy := []string{p.roleName, p.domain, p.object, p.action}
		casbinPolicies = append(casbinPolicies, casbinPolicy)
	}

	// Bulk create policies in database
	if err := s.policyRepo.CreateBatch(ctx, policyEntities); err != nil {
		return fmt.Errorf("failed to create policies in batch for role %s, domain %s: %w", roleName, domain, err)
	}

	// Bulk add policies to Casbin
	if _, err := s.enforcerService.AddPolicies(casbinPolicies); err != nil {
		log.Printf("Failed to add Casbin policies in batch for role %s, domain %s: %v", roleName, domain, err)
		// Note: We could implement rollback here, but for now just log the error
	}

	log.Printf("Created %d domain-specific policies in batch for role %s, domain %s", len(policies), roleName, domain)

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
