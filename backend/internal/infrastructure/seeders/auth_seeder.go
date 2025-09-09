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
			DisplayName: "Super Administrator",
			Description: strPtr("Full system access across all domains"),
			IsActive:    true,
		},
		{
			Name:        "admin",
			DisplayName: "Administrator",
			Description: strPtr("Administrative access within a domain"),
			IsActive:    true,
		},
		{
			Name:        "manager",
			DisplayName: "Manager",
			Description: strPtr("Management access to shop operations"),
			IsActive:    true,
		},
		{
			Name:        "cashier",
			DisplayName: "Cashier",
			Description: strPtr("Point of sale operations"),
			IsActive:    true,
		},
		{
			Name:        "user",
			DisplayName: "User",
			Description: strPtr("Basic user access"),
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

	// Define policies for different roles
	policies := []struct {
		roleName string
		domain   string
		object   string
		action   string
	}{
		// Super Admin - Full access
		{"super_admin", "*", "/api/v1/*", "GET|POST|PUT|DELETE"},
		{"super_admin", "*", "/api/v1/admin/*", "GET|POST|PUT|DELETE"},

		// Admin - Domain specific full access
		{"admin", "*", "/api/v1/*", "GET|POST|PUT|DELETE"},
		{"admin", "*", "/api/v1/products/*", "GET|POST|PUT|DELETE"},
		{"admin", "*", "/api/v1/checkout/*", "GET|POST|PUT|DELETE"},
		{"admin", "*", "/api/v1/transactions/*", "GET|POST|PUT|DELETE"},
		{"admin", "*", "/api/v1/users/*", "GET|POST|PUT"},

		// Manager - Shop management access
		{"manager", "*", "/api/v1/products", "GET|POST"},
		{"manager", "*", "/api/v1/products/*", "GET|PUT"},
		{"manager", "*", "/api/v1/checkout/*", "GET|POST"},
		{"manager", "*", "/api/v1/transactions/*", "GET"},
		{"manager", "*", "/api/v1/users", "GET"},
		{"manager", "*", "/api/v1/users/*", "GET"},

		// Cashier - POS operations
		{"cashier", "*", "/api/v1/products", "GET"},
		{"cashier", "*", "/api/v1/products/*", "GET"},
		{"cashier", "*", "/api/v1/products/search", "GET"},
		{"cashier", "*", "/api/v1/products/barcode/*", "GET"},
		{"cashier", "*", "/api/v1/checkout", "POST"},
		{"cashier", "*", "/api/v1/checkout/*", "POST"},
		{"cashier", "*", "/api/v1/transactions/*", "GET"},

		// User - Basic read access
		{"user", "*", "/api/v1/products", "GET"},
		{"user", "*", "/api/v1/products/*", "GET"},
		{"user", "*", "/api/v1/auth/profile", "GET"},
		{"user", "*", "/api/v1/auth/logout", "POST"},
		{"user", "*", "/api/v1/auth/refresh", "POST"},
		{"user", "*", "/api/v1/auth/permissions", "GET"},
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