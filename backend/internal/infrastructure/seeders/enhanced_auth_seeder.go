package seeders

import (
	"context"
	"fmt"
	"log"

	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
)

// EnhancedAuthSeeder handles seeding with Casbin grouping optimization
type EnhancedAuthSeeder struct {
	roleRepo        repositories.RoleRepository
	userRepo        repositories.UserRepository
	policyRepo      repositories.PolicyRepository
	enforcerService *casbin.EnforcerService
}

// NewEnhancedAuthSeeder creates a new enhanced auth seeder with grouping support
func NewEnhancedAuthSeeder(
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	policyRepo repositories.PolicyRepository,
	enforcerService *casbin.EnforcerService,
) *EnhancedAuthSeeder {
	return &EnhancedAuthSeeder{
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		policyRepo:      policyRepo,
		enforcerService: enforcerService,
	}
}

// SeedRoleGroupingsWithOptimization creates optimized role-based policies using Casbin grouping
func (s *EnhancedAuthSeeder) SeedRoleGroupingsWithOptimization() error {
	log.Println("Starting optimized role grouping seeding...")

	// Create base role policies (without domain-specific duplication)
	if err := s.createBaseRolePolicies(); err != nil {
		return fmt.Errorf("failed to create base role policies: %w", err)
	}

	// Create role groupings for existing users
	if err := s.createUserRoleGroupings(); err != nil {
		return fmt.Errorf("failed to create user role groupings: %w", err)
	}

	log.Println("Optimized role grouping seeding completed successfully")
	return nil
}

// createBaseRolePolicies creates optimized base policies for each role
func (s *EnhancedAuthSeeder) createBaseRolePolicies() error {
	ctx := context.Background()

	// Get all roles
	roles, err := s.roleRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}

	// Create optimized policies for each role
	for _, role := range roles {
		if err := s.createOptimizedRolePolicies(role.Name); err != nil {
			log.Printf("Error creating policies for role %s: %v", role.Name, err)
			continue
		}
		log.Printf("Created optimized policies for role: %s", role.Name)
	}

	return nil
}

// createOptimizedRolePolicies creates wildcard-optimized policies for a role
func (s *EnhancedAuthSeeder) createOptimizedRolePolicies(roleName string) error {
	// Use wildcard domain "*" for role-level policies
	// Users will be assigned to roles with specific domains through grouping
	baseDomain := "*"
	
	var policies [][]string

	switch roleName {
	case "super_admin":
		// Super admin gets global access with minimal policies
		policies = [][]string{
			{roleName, baseDomain, "/api/v1/*", "GET|POST|PUT|DELETE"},
		}

	case "admin":
		// Admin gets broad access with minimal policies
		policies = [][]string{
			{roleName, baseDomain, "/api/v1/*", "GET|POST|PUT|DELETE"},
		}

	case "owner_business":
		// Owner business gets comprehensive access with optimized wildcards
		policies = [][]string{
			// Products management
			{roleName, baseDomain, "/api/v1/products*", "GET|POST|PUT|DELETE"},
			
			// Categories management
			{roleName, baseDomain, "/api/v1/categories*", "GET|POST|PUT|DELETE"},
			
			// Shops management
			{roleName, baseDomain, "/api/v1/shops*", "GET|POST|PUT|DELETE"},
			
			// Transactions management
			{roleName, baseDomain, "/api/v1/transactions*", "GET|POST|PUT|DELETE"},
			
			// Carts management
			{roleName, baseDomain, "/api/v1/carts*", "GET|POST|PUT|DELETE"},
			
			// Expenses management
			{roleName, baseDomain, "/api/v1/expenses*", "GET|POST|PUT|DELETE"},
			
			// Payments management
			{roleName, baseDomain, "/api/v1/payments*", "GET|POST|PUT|DELETE"},
			
			// Histories management
			{roleName, baseDomain, "/api/v1/histories*", "GET|POST|PUT|DELETE"},
			
			// Receipts management
			{roleName, baseDomain, "/api/v1/receipts*", "GET|POST|PUT|DELETE"},
			
			// Transaction products management
			{roleName, baseDomain, "/api/v1/transaction-products*", "GET|POST|PUT|DELETE"},
			
			// User management
			{roleName, baseDomain, "/api/v1/auth/cashier/register", "POST"},
			{roleName, baseDomain, "/api/v1/users*", "GET|POST|PUT|DELETE"},
			
			// ACL management
			{roleName, baseDomain, "/api/v1/acl*", "GET|POST|PUT|DELETE"},
			
			// Sync operations
			{roleName, baseDomain, "/api/v1/sync*", "GET|POST"},
		}

	case "cashier":
		// Cashier gets limited access with optimized policies
		policies = [][]string{
			// Read-only products and categories
			{roleName, baseDomain, "/api/v1/products*", "GET"},
			{roleName, baseDomain, "/api/v1/categories*", "GET"},
			
			// Cart operations
			{roleName, baseDomain, "/api/v1/carts*", "GET|POST|PUT|DELETE"},
			
			// Transaction operations
			{roleName, baseDomain, "/api/v1/transactions*", "GET|POST|PUT"},
			{roleName, baseDomain, "/api/v1/transaction-products*", "GET|POST|PUT|DELETE"},
			
			// Payment operations
			{roleName, baseDomain, "/api/v1/payments*", "GET|POST|PUT"},
			
			// Receipt generation
			{roleName, baseDomain, "/api/v1/receipts*", "GET|POST"},
			
			// History viewing
			{roleName, baseDomain, "/api/v1/histories*", "GET"},
			
			// Limited sync operations
			{roleName, baseDomain, "/api/v1/sync*", "GET|POST"},
			
			// Profile management
			{roleName, baseDomain, "/api/v1/auth/profile", "GET"},
			{roleName, baseDomain, "/api/v1/auth/pin", "GET|POST|PUT|DELETE"},
		}
	}

	// Add the policies to Casbin
	for _, policy := range policies {
		if _, err := s.enforcerService.AddPolicy(policy[0], policy[1], policy[2], policy[3]); err != nil {
			log.Printf("Warning: Failed to add policy %v: %v", policy, err)
			// Continue with other policies instead of failing completely
		}
	}

	return nil
}

// createUserRoleGroupings creates user-to-role groupings with domain context
func (s *EnhancedAuthSeeder) createUserRoleGroupings() error {
	ctx := context.Background()

	// Get all users
	users, err := s.userRepo.List(ctx, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get users: %w", err)
	}

	var groupings [][]string

	for _, user := range users {
		if user.RoleID == nil {
			continue
		}

		// Get user role
		role, err := s.roleRepo.GetByID(ctx, *user.RoleID)
		if err != nil {
			log.Printf("Warning: Failed to get role for user %s: %v", user.ID, err)
			continue
		}

		// Determine domain based on role and user assignment
		domain := s.getUserDomain(user, role.Name)
		
		// Create user-to-role grouping: g, user_id, role_name, domain
		grouping := []string{user.ID.String(), role.Name, domain}
		groupings = append(groupings, grouping)
		
		log.Printf("Created grouping: user %s -> role %s in domain %s", user.ID, role.Name, domain)
	}

	// Batch add groupings for better performance
	if err := s.addGroupingsBatch(groupings); err != nil {
		return fmt.Errorf("failed to add user role groupings: %w", err)
	}

	return nil
}

// getUserDomain determines the appropriate domain for a user based on their role and assignments
func (s *EnhancedAuthSeeder) getUserDomain(user *entities.User, roleName string) string {
	switch roleName {
	case "super_admin", "admin":
		return "*" // Global access
		
	case "owner_business":
		if user.LicenseID != nil {
			// Use license serial as domain
			return fmt.Sprintf("LIC-%s", user.LicenseID.String()[:8])
		}
		return "*" // Fallback to global if no license
		
	case "cashier":
		if user.ShopID != nil {
			// Use shop-specific domain
			return fmt.Sprintf("shop-%s", user.ShopID.String())
		}
		return "*" // Fallback to global if no shop assignment
		
	default:
		return "*" // Unknown roles get global access
	}
}

// addGroupingsBatch adds multiple groupings in batch for better performance
func (s *EnhancedAuthSeeder) addGroupingsBatch(groupings [][]string) error {
	// Add groupings to Casbin enforcer
	for _, grouping := range groupings {
		if len(grouping) != 3 {
			log.Printf("Warning: Invalid grouping format: %v", grouping)
			continue
		}
		
		if _, err := s.enforcerService.AddGroupingPolicy(grouping[0], grouping[1], grouping[2]); err != nil {
			log.Printf("Warning: Failed to add grouping %v: %v", grouping, err)
			// Continue with other groupings instead of failing
		}
	}

	// Save to database
	if err := s.enforcerService.SavePolicy(); err != nil {
		return fmt.Errorf("failed to save grouping policies: %w", err)
	}

	log.Printf("Successfully added %d user role groupings", len(groupings))
	return nil
}

// OptimizePolicyPerformance performs various optimizations to improve policy lookup performance
func (s *EnhancedAuthSeeder) OptimizePolicyPerformance() error {
	log.Println("Starting policy performance optimization...")

	// Remove duplicate policies
	if err := s.removeDuplicatePolicies(); err != nil {
		log.Printf("Warning: Failed to remove duplicate policies: %v", err)
	}

	// Rebuild policy cache
	if err := s.enforcerService.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to reload policies: %w", err)
	}

	log.Println("Policy performance optimization completed")
	return nil
}

// removeDuplicatePolicies removes duplicate policies to improve performance
func (s *EnhancedAuthSeeder) removeDuplicatePolicies() error {
	// Get all policies
	policies := s.enforcerService.GetPolicy()
	
	// Track unique policies
	seen := make(map[string]bool)
	var duplicates [][]string
	
	for _, policy := range policies {
		// Create a key from policy components
		key := fmt.Sprintf("%s|%s|%s|%s", policy[0], policy[1], policy[2], policy[3])
		
		if seen[key] {
			duplicates = append(duplicates, policy)
		} else {
			seen[key] = true
		}
	}

	// Remove duplicates
	for _, duplicate := range duplicates {
		if _, err := s.enforcerService.RemovePolicy(duplicate[0], duplicate[1], duplicate[2], duplicate[3]); err != nil {
			log.Printf("Warning: Failed to remove duplicate policy %v: %v", duplicate, err)
		}
	}

	if len(duplicates) > 0 {
		log.Printf("Removed %d duplicate policies", len(duplicates))
		// Save the cleaned policies
		return s.enforcerService.SavePolicy()
	}

	return nil
}

// GetPerformanceMetrics returns metrics about the current policy setup
func (s *EnhancedAuthSeeder) GetPerformanceMetrics() map[string]interface{} {
	policies := s.enforcerService.GetPolicy()
	groupings := s.enforcerService.GetGroupingPolicy()
	
	return map[string]interface{}{
		"total_policies":          len(policies),
		"total_groupings":         len(groupings),
		"policy_to_grouping_ratio": float64(len(policies)) / float64(max(len(groupings), 1)),
		"estimated_lookup_time":   s.estimateLookupTime(len(policies), len(groupings)),
	}
}

// estimateLookupTime estimates policy lookup time based on policy count
func (s *EnhancedAuthSeeder) estimateLookupTime(policyCount, groupingCount int) string {
	// Simple heuristic: lookup time increases with policy count
	// With grouping, we expect better performance
	baseTime := float64(policyCount) * 0.1 // 0.1ms per policy
	
	if groupingCount > 0 {
		// Grouping reduces lookup time
		baseTime = baseTime * 0.5
	}
	
	if baseTime < 1 {
		return "< 1ms"
	} else if baseTime < 10 {
		return fmt.Sprintf("~%.1fms", baseTime)
	} else {
		return "> 10ms"
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}