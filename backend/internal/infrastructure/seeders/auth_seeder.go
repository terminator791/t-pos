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

	// Define policies for different roles with shop-based domain
	policies := []struct {
		roleName string
		domain   string
		object   string
		action   string
	}{
		// Super Admin - Full access across all domains/shops
		{"super_admin", "*", "/api/v1/*", "GET"},
		{"super_admin", "*", "/api/v1/*", "POST"},
		{"super_admin", "*", "/api/v1/*", "PUT"},
		{"super_admin", "*", "/api/v1/*", "DELETE"},
		{"super_admin", "*", "/api/v1/licenses", "GET"},
		{"super_admin", "*", "/api/v1/licenses/*", "GET"},
		{"super_admin", "*", "/api/v1/licenses", "POST"},
		{"super_admin", "*", "/api/v1/licenses/*", "DELETE"},
		{"super_admin", "*", "/api/v1/users", "GET"},
		{"super_admin", "*", "/api/v1/users/*", "GET"},
		{"super_admin", "*", "/api/v1/users", "POST"},
		{"super_admin", "*", "/api/v1/users/*", "PUT"},
		{"super_admin", "*", "/api/v1/users/*", "DELETE"},
		{"super_admin", "*", "/api/v1/customers", "GET"},
		{"super_admin", "*", "/api/v1/customers/*", "GET"},
		{"super_admin", "*", "/api/v1/customers", "POST"},
		{"super_admin", "*", "/api/v1/customers/*", "DELETE"},

		// Admin - Full access across all shops but not system-wide resources like licenses
		{"admin", "*", "/api/v1/products", "GET"},
		{"admin", "*", "/api/v1/products/*", "GET"},
		{"admin", "*", "/api/v1/products", "POST"},
		{"admin", "*", "/api/v1/products/*", "PUT"},
		{"admin", "*", "/api/v1/products/*", "DELETE"},
		{"admin", "*", "/api/v1/products/search", "GET"},
		{"admin", "*", "/api/v1/products/barcode/*", "GET"},
		{"admin", "*", "/api/v1/products/low-stock", "GET"},
		{"admin", "*", "/api/v1/checkout", "POST"},
		{"admin", "*", "/api/v1/checkout/*", "POST"},
		{"admin", "*", "/api/v1/transactions/*", "GET"},
		{"admin", "*", "/api/v1/transactions/shop/*/today", "GET"},
		{"admin", "*", "/api/v1/customers", "GET"},
		{"admin", "*", "/api/v1/customers/*", "GET"},
		{"admin", "*", "/api/v1/customers", "POST"},
		{"admin", "*", "/api/v1/customers/*", "DELETE"},
		{"admin", "*", "/api/v1/users/cashier", "GET"},
		{"admin", "*", "/api/v1/users/cashier", "POST"},
		{"admin", "*", "/api/v1/users/cashier/*", "GET"},
		{"admin", "*", "/api/v1/users/cashier/*", "PUT"},
		{"admin", "*", "/api/v1/users/cashier/*", "DELETE"},

		// Owner Business - Full shop management access within specific shop
		{"owner_business", "shop:*", "/api/v1/products", "GET"},
		{"owner_business", "shop:*", "/api/v1/products/*", "GET"},
		{"owner_business", "shop:*", "/api/v1/products", "POST"},
		{"owner_business", "shop:*", "/api/v1/products/*", "PUT"},
		{"owner_business", "shop:*", "/api/v1/products/*", "DELETE"},
		{"owner_business", "shop:*", "/api/v1/products/search", "GET"},
		{"owner_business", "shop:*", "/api/v1/products/barcode/*", "GET"},
		{"owner_business", "shop:*", "/api/v1/products/low-stock", "GET"},
		{"owner_business", "shop:*", "/api/v1/checkout", "POST"},
		{"owner_business", "shop:*", "/api/v1/checkout/*", "POST"},
		{"owner_business", "shop:*", "/api/v1/transactions/*", "GET"},
		{"owner_business", "shop:*", "/api/v1/transactions/shop/*/today", "GET"},
		{"owner_business", "shop:*", "/api/v1/customers", "GET"},
		{"owner_business", "shop:*", "/api/v1/customers/*", "GET"},
		{"owner_business", "shop:*", "/api/v1/customers", "POST"},
		{"owner_business", "shop:*", "/api/v1/customers/*", "DELETE"},
		{"owner_business", "shop:*", "/api/v1/users/cashier", "GET"},
		{"owner_business", "shop:*", "/api/v1/users/cashier", "POST"},
		{"owner_business", "shop:*", "/api/v1/users/cashier/*", "GET"},
		{"owner_business", "shop:*", "/api/v1/users/cashier/*", "PUT"},
		{"owner_business", "shop:*", "/api/v1/users/cashier/*", "DELETE"},

		// Cashier - POS operations within specific shop
		{"cashier", "shop:*", "/api/v1/products", "GET"},
		{"cashier", "shop:*", "/api/v1/products/*", "GET"},
		{"cashier", "shop:*", "/api/v1/products", "POST"},
		{"cashier", "shop:*", "/api/v1/products/*", "PUT"},
		{"cashier", "shop:*", "/api/v1/products/*", "DELETE"},
		{"cashier", "shop:*", "/api/v1/products/search", "GET"},
		{"cashier", "shop:*", "/api/v1/products/barcode/*", "GET"},
		{"cashier", "shop:*", "/api/v1/products/low-stock", "GET"},
		{"cashier", "shop:*", "/api/v1/checkout", "POST"},
		{"cashier", "shop:*", "/api/v1/checkout/*", "POST"},
		{"cashier", "shop:*", "/api/v1/transactions/*", "GET"},
		{"cashier", "shop:*", "/api/v1/transactions/shop/*/today", "GET"},
		{"cashier", "shop:*", "/api/v1/customers", "GET"},
		{"cashier", "shop:*", "/api/v1/customers/*", "GET"},
		{"cashier", "shop:*", "/api/v1/customers", "POST"},
		{"cashier", "shop:*", "/api/v1/customers/*", "DELETE"},

		// Auth endpoints for all roles (within their domains)
		{"super_admin", "*", "/api/v1/auth/profile", "GET"},
		{"super_admin", "*", "/api/v1/auth/logout", "POST"},
		{"super_admin", "*", "/api/v1/auth/refresh", "POST"},
		{"super_admin", "*", "/api/v1/auth/permissions", "GET"},
		
		{"admin", "*", "/api/v1/auth/profile", "GET"},
		{"admin", "*", "/api/v1/auth/logout", "POST"},
		{"admin", "*", "/api/v1/auth/refresh", "POST"},
		{"admin", "*", "/api/v1/auth/permissions", "GET"},
		
		{"owner_business", "shop:*", "/api/v1/auth/profile", "GET"},
		{"owner_business", "shop:*", "/api/v1/auth/logout", "POST"},
		{"owner_business", "shop:*", "/api/v1/auth/refresh", "POST"},
		{"owner_business", "shop:*", "/api/v1/auth/permissions", "GET"},
		
		{"cashier", "shop:*", "/api/v1/auth/profile", "GET"},
		{"cashier", "shop:*", "/api/v1/auth/logout", "POST"},
		{"cashier", "shop:*", "/api/v1/auth/refresh", "POST"},
		{"cashier", "shop:*", "/api/v1/auth/permissions", "GET"},
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