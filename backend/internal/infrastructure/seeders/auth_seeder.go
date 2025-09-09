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
		// Super Admin - Full access across all shops
		{"super_admin", "*", "/api/v1/*", "GET|POST|PUT|DELETE"},
		{"super_admin", "*", "/api/v1/admin/*", "GET|POST|PUT|DELETE"},

		// Admin - Shop specific full access (domain represents specific shop)
		{"admin", "shop*", "/api/v1/*", "GET|POST|PUT|DELETE"},
		{"admin", "shop*", "/api/v1/products/*", "GET|POST|PUT|DELETE"},
		{"admin", "shop*", "/api/v1/checkout/*", "GET|POST|PUT|DELETE"},
		{"admin", "shop*", "/api/v1/transactions/*", "GET|POST|PUT|DELETE"},
		{"admin", "shop*", "/api/v1/users/*", "GET|POST|PUT"},

		// Owner Business - Full shop management access
		{"owner_business", "shop*", "/api/v1/products", "GET|POST"},
		{"owner_business", "shop*", "/api/v1/products/*", "GET|PUT|DELETE"},
		{"owner_business", "shop*", "/api/v1/checkout/*", "GET|POST"},
		{"owner_business", "shop*", "/api/v1/transactions/*", "GET"},
		{"owner_business", "shop*", "/api/v1/users", "GET"},
		{"owner_business", "shop*", "/api/v1/users/*", "GET|PUT"},
		{"owner_business", "shop*", "/api/v1/reports/*", "GET"},

		// Cashier - POS operations within specific shop
		{"cashier", "shop*", "/api/v1/products", "GET"},
		{"cashier", "shop*", "/api/v1/products/*", "GET"},
		{"cashier", "shop*", "/api/v1/products/search", "GET"},
		{"cashier", "shop*", "/api/v1/products/barcode/*", "GET"},
		{"cashier", "shop*", "/api/v1/checkout", "POST"},
		{"cashier", "shop*", "/api/v1/checkout/*", "POST"},
		{"cashier", "shop*", "/api/v1/transactions/*", "GET"},

		// Auth endpoints for all roles (within their shops)
		{"super_admin", "*", "/api/v1/auth/profile", "GET"},
		{"super_admin", "*", "/api/v1/auth/logout", "POST"},
		{"super_admin", "*", "/api/v1/auth/refresh", "POST"},
		{"super_admin", "*", "/api/v1/auth/permissions", "GET"},
		
		{"admin", "shop*", "/api/v1/auth/profile", "GET"},
		{"admin", "shop*", "/api/v1/auth/logout", "POST"},
		{"admin", "shop*", "/api/v1/auth/refresh", "POST"},
		{"admin", "shop*", "/api/v1/auth/permissions", "GET"},
		
		{"owner_business", "shop*", "/api/v1/auth/profile", "GET"},
		{"owner_business", "shop*", "/api/v1/auth/logout", "POST"},
		{"owner_business", "shop*", "/api/v1/auth/refresh", "POST"},
		{"owner_business", "shop*", "/api/v1/auth/permissions", "GET"},
		
		{"cashier", "shop*", "/api/v1/auth/profile", "GET"},
		{"cashier", "shop*", "/api/v1/auth/logout", "POST"},
		{"cashier", "shop*", "/api/v1/auth/refresh", "POST"},
		{"cashier", "shop*", "/api/v1/auth/permissions", "GET"},
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