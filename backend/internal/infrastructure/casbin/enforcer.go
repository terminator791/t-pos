package casbin

import (
	"fmt"
	"log"
	"sync"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// EnforcerService manages the Casbin enforcer singleton
type EnforcerService struct {
	enforcer *casbin.Enforcer
	adapter  *gormadapter.Adapter
	mu       sync.RWMutex
}

var (
	enforcerInstance *EnforcerService
	once             sync.Once
)

// GetEnforcerService returns the singleton enforcer instance
func GetEnforcerService() *EnforcerService {
	if enforcerInstance == nil {
		panic("enforcer service not initialized, call NewEnforcerService first")
	}
	return enforcerInstance
}

// NewEnforcerService creates and initializes the Casbin enforcer singleton
func NewEnforcerService(db *gorm.DB, modelPath string) (*EnforcerService, error) {
	var err error
	once.Do(func() {
		// Create GORM adapter
		adapter, adapterErr := gormadapter.NewAdapterByDB(db)
		if adapterErr != nil {
			err = fmt.Errorf("failed to create gorm adapter: %w", adapterErr)
			return
		}

		// Create enforcer with model and adapter
		enforcer, enforcerErr := casbin.NewEnforcer(modelPath, adapter)
		if enforcerErr != nil {
			err = fmt.Errorf("failed to create casbin enforcer: %w", enforcerErr)
			return
		}

		// Load policies from database
		if loadErr := enforcer.LoadPolicy(); loadErr != nil {
			err = fmt.Errorf("failed to load policies: %w", loadErr)
			return
		}

		// Enable auto-save
		enforcer.EnableAutoSave(true)

		enforcerInstance = &EnforcerService{
			enforcer: enforcer,
			adapter:  adapter,
		}

		log.Println("Casbin enforcer initialized successfully")
	})

	return enforcerInstance, err
}

// Enforce checks if a user has permission to access a resource
func (e *EnforcerService) Enforce(user, domain, object, action string) (bool, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.enforcer.Enforce(user, domain, object, action)
}

// AddPolicy adds a new policy rule
func (e *EnforcerService) AddPolicy(role, domain, object, action string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.AddPolicy(role, domain, object, action)
}

// RemovePolicy removes a policy rule
func (e *EnforcerService) RemovePolicy(role, domain, object, action string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.RemovePolicy(role, domain, object, action)
}

// AddRoleForUser adds a role for a user in a specific domain
func (e *EnforcerService) AddRoleForUser(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.AddRoleForUserInDomain(user, role, domain)
}

// RemoveRoleForUser removes a role for a user in a specific domain
func (e *EnforcerService) RemoveRoleForUser(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.DeleteRoleForUserInDomain(user, role, domain)
}

// GetRolesForUser gets all roles for a user in a specific domain
func (e *EnforcerService) GetRolesForUser(user, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.enforcer.GetRolesForUserInDomain(user, domain)
}

// GetUsersForRole gets all users with a specific role in a domain
func (e *EnforcerService) GetUsersForRole(role, domain string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.enforcer.GetUsersForRoleInDomain(role, domain)
}

// LoadPolicy reloads policies from database
func (e *EnforcerService) LoadPolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.LoadPolicy()
}

// SavePolicy saves current policies to database
func (e *EnforcerService) SavePolicy() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.SavePolicy()
}

// GetAllPolicies returns all policy rules
func (e *EnforcerService) GetAllPolicies() [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, _ := e.enforcer.GetPolicy()
	return policies
}

// GetAllRoles returns all role assignments
func (e *EnforcerService) GetAllRoles() [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	roles, _ := e.enforcer.GetGroupingPolicy()
	return roles
}

// AddPolicies adds multiple policy rules in batch
func (e *EnforcerService) AddPolicies(policies [][]string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.AddPolicies(policies)
}

// AddPoliciesForRole adds multiple policies for a specific role and domain
func (e *EnforcerService) AddPoliciesForRole(role, domain string, permissions [][]string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Convert permissions to full policy format
	var policies [][]string
	for _, perm := range permissions {
		if len(perm) >= 2 { // object, action
			policy := []string{role, domain, perm[0], perm[1]}
			policies = append(policies, policy)
		}
	}

	if len(policies) > 0 {
		_, err := e.enforcer.AddPolicies(policies)
		return err
	}

	return nil
}

// AddGroupingPolicy adds a grouping policy (user-role assignment with domain)
func (e *EnforcerService) AddGroupingPolicy(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.AddGroupingPolicy(user, role, domain)
}

// RemoveGroupingPolicy removes a grouping policy
func (e *EnforcerService) RemoveGroupingPolicy(user, role, domain string) (bool, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.enforcer.RemoveGroupingPolicy(user, role, domain)
}

// GetGroupingPolicy returns all grouping policies
func (e *EnforcerService) GetGroupingPolicy() [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, _ := e.enforcer.GetGroupingPolicy()
	return policies
}

// GetPolicy returns all policies
func (e *EnforcerService) GetPolicy() [][]string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	policies, _ := e.enforcer.GetPolicy()
	return policies
}
