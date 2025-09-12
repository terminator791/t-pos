package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/repositories"
)

// DomainAccessInfo contains information about what domains/shops a user can access
type DomainAccessInfo struct {
	UserID      uuid.UUID
	Role        string
	Domain      string
	HasGlobalAccess bool
	LicenseID   *uuid.UUID
	ShopID      *uuid.UUID
	AccessibleShopIDs []uuid.UUID
}

// GetUserDomainAccess extracts domain access information from the current request context
func GetUserDomainAccess(c *gin.Context, roleRepo repositories.RoleRepository, shopRepo repositories.ShopRepository) (*DomainAccessInfo, error) {
	// Get user from context
	user, exists := GetUserFromContext(c)
	if !exists {
		return nil, fmt.Errorf("user not found in context")
	}

	// Get domain from context
	domain, _ := GetUserDomainFromContext(c)

	// Get role information
	var roleName string
	if user.RoleID != nil {
		role, err := roleRepo.GetByID(context.Background(), *user.RoleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user role: %w", err)
		}
		roleName = role.Name
	}

	access := &DomainAccessInfo{
		UserID:      user.ID,
		Role:        roleName,
		Domain:      domain,
		LicenseID:   user.LicenseID,
		ShopID:      user.ShopID,
	}

	// Determine access level based on role and domain
	switch {
	case domain == "*" || roleName == "super_admin" || roleName == "admin":
		// Global access
		access.HasGlobalAccess = true
		
	case roleName == "owner_business" && user.LicenseID != nil:
		// Owner business - can access all shops under their license
		shops, err := shopRepo.GetByLicenseID(context.Background(), *user.LicenseID)
		if err != nil {
			return nil, fmt.Errorf("failed to get shops for license: %w", err)
		}
		
		for _, shop := range shops {
			access.AccessibleShopIDs = append(access.AccessibleShopIDs, shop.ID)
		}
		
	case roleName == "cashier" && user.ShopID != nil:
		// Cashier - can only access their assigned shop
		access.AccessibleShopIDs = append(access.AccessibleShopIDs, *user.ShopID)
		
	default:
		return nil, fmt.Errorf("invalid user access configuration - role: %s, domain: %s", roleName, domain)
	}

	return access, nil
}

// CanAccessShop checks if the user can access a specific shop
func (access *DomainAccessInfo) CanAccessShop(shopID uuid.UUID) bool {
	if access.HasGlobalAccess {
		return true
	}
	
	for _, accessibleShopID := range access.AccessibleShopIDs {
		if accessibleShopID == shopID {
			return true
		}
	}
	
	return false
}

// CanAccessLicense checks if the user can access a specific license
func (access *DomainAccessInfo) CanAccessLicense(licenseID uuid.UUID) bool {
	if access.HasGlobalAccess {
		return true
	}
	
	// Owner business can access their own license
	if access.LicenseID != nil && *access.LicenseID == licenseID {
		return true
	}
	
	return false
}

// GetShopFilter returns shop IDs that should be used to filter queries
// Returns nil if user has global access (no filtering needed)
func (access *DomainAccessInfo) GetShopFilter() []uuid.UUID {
	if access.HasGlobalAccess {
		return nil // No filtering needed
	}
	
	return access.AccessibleShopIDs
}

// GetLicenseFilter returns license IDs that should be used to filter queries  
// Returns nil if user has global access (no filtering needed)
func (access *DomainAccessInfo) GetLicenseFilter() []uuid.UUID {
	if access.HasGlobalAccess {
		return nil // No filtering needed
	}
	
	if access.LicenseID != nil {
		return []uuid.UUID{*access.LicenseID}
	}
	
	return []uuid.UUID{} // No license access
}