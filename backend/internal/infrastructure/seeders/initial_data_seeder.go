package seeders

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"github.com/terminator791/t-pos/internal/domain/repositories"
	"github.com/terminator791/t-pos/internal/infrastructure/casbin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// InitialDataSeeder handles seeding of initial application data
type InitialDataSeeder struct {
	licenseRepo     repositories.LicenseRepository
	userRepo        repositories.UserRepository
	roleRepo        repositories.RoleRepository
	shopRepo        repositories.ShopRepository
	categoryRepo    repositories.CategoryRepository
	productRepo     repositories.ProductRepository
	userDomainRepo  repositories.UserDomainRepository
	enforcerService *casbin.EnforcerService
}

// NewInitialDataSeeder creates a new initial data seeder
func NewInitialDataSeeder(
	licenseRepo repositories.LicenseRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	shopRepo repositories.ShopRepository,
	categoryRepo repositories.CategoryRepository,
	productRepo repositories.ProductRepository,
	userDomainRepo repositories.UserDomainRepository,
	enforcerService *casbin.EnforcerService,
) *InitialDataSeeder {
	return &InitialDataSeeder{
		licenseRepo:     licenseRepo,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		shopRepo:        shopRepo,
		categoryRepo:    categoryRepo,
		productRepo:     productRepo,
		userDomainRepo:  userDomainRepo,
		enforcerService: enforcerService,
	}
}

// SeedLicenses creates initial licenses
func (s *InitialDataSeeder) SeedLicenses() error {
	ctx := context.Background()

	licenses := []entities.License{
		{
			SerialNumber: "LIC-001-DEMO",
		},
		{
			SerialNumber: "LIC-002-DEMO",
		},
	}

	for _, license := range licenses {
		// Check if license already exists
		existing, err := s.licenseRepo.GetBySerialNumber(ctx, license.SerialNumber)
		if err == gorm.ErrRecordNotFound {
			// Create the license
			if err := s.licenseRepo.Create(ctx, &license); err != nil {
				log.Printf("Failed to create license %s: %v", license.SerialNumber, err)
				return err
			}
			log.Printf("Created license: %s", license.SerialNumber)
		} else if err != nil {
			log.Printf("Error checking license %s: %v", license.SerialNumber, err)
			return err
		} else {
			log.Printf("License %s already exists", existing.SerialNumber)
		}
	}

	return nil
}

// SeedUsers creates initial users (super admin and admin)
func (s *InitialDataSeeder) SeedUsers() error {
	ctx := context.Background()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return err
	}

	// Hash pin
	hashedPin, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash pin: %v", err)
		return err
	}

	users := []entities.User{
		{
			LicenseID: nil, // Super admin doesn't need license
			Email:     strPtr("superadmin@example.com"),
			Username:  strPtr("superadmin"),
			Name:      "Super Admin",
			Pin:       strPtr(string(hashedPin)),
			Password:  string(hashedPassword),
		},
		{
			LicenseID: nil, // Admin doesn't need license
			Email:     strPtr("admin@example.com"),
			Username:  strPtr("admin"),
			Name:      "Admin User",
			Pin:       strPtr(string(hashedPin)),
			Password:  string(hashedPassword),
		},
	}

	for _, user := range users {
		// Check if user already exists by email
		if user.Email != nil {
			existing, err := s.userRepo.GetByEmail(ctx, *user.Email)
			if err == gorm.ErrRecordNotFound {
				// Create the user
				if err := s.userRepo.Create(ctx, &user); err != nil {
					log.Printf("Failed to create user %s: %v", user.Name, err)
					return err
				}
				log.Printf("Created user: %s", user.Name)
			} else if err != nil {
				log.Printf("Error checking user %s: %v", user.Name, err)
				return err
			} else {
				log.Printf("User %s already exists", existing.Name)
			}
		}
	}

	return nil
}

// SeedUserRoles assigns roles to users with new single role system
func (s *InitialDataSeeder) SeedUserRoles() error {
	ctx := context.Background()

	// Get users
	superAdmin, err := s.userRepo.GetByEmail(ctx, "superadmin@example.com")
	if err != nil {
		log.Printf("Failed to get super admin user: %v", err)
		return err
	}

	admin, err := s.userRepo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Printf("Failed to get admin user: %v", err)
		return err
	}

	// Get roles
	superAdminRole, err := s.roleRepo.GetByName(ctx, "super_admin")
	if err != nil {
		log.Printf("Failed to get super_admin role: %v", err)
		return err
	}

	adminRole, err := s.roleRepo.GetByName(ctx, "admin")
	if err != nil {
		log.Printf("Failed to get admin role: %v", err)
		return err
	}

	// Update users with their roles directly
	superAdmin.RoleID = &superAdminRole.ID
	if err := s.userRepo.Update(ctx, superAdmin); err != nil {
		log.Printf("Failed to assign role to super admin: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to superadmin user")

	admin.RoleID = &adminRole.ID
	if err := s.userRepo.Update(ctx, admin); err != nil {
		log.Printf("Failed to assign role to admin: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin user")

	return nil
}

// SeedShops creates initial shops
func (s *InitialDataSeeder) SeedShops() error {
	ctx := context.Background()

	// Get the admin user
	admin, err := s.userRepo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Printf("Failed to get admin user: %v", err)
		return err
	}

	// Get the license for the shop
	license, err := s.licenseRepo.GetBySerialNumber(ctx, "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to get license: %v", err)
		return err
	}

	shops := []entities.Shop{
		{
			LicenseID: license.ID,
			UserID:    admin.ID,
			Name:      "Demo Shop",
			Address:   strPtr("Jl. Demo No. 123, Jakarta"),
			Slogan:    strPtr("Your Trusted Partner"),
		},
		{
			LicenseID: license.ID,
			UserID:    admin.ID,
			Name:      "Demo Shop 2",
			Address:   strPtr("Jl. Demo No. 123sda, Semarang"),
			Slogan:    strPtr("Your Trusted Partner"),
		},
	}

	for _, shop := range shops {
		// Check if shop already exists by name
		// Get all shops for the license and check names
		existingShops, err := s.shopRepo.GetByLicenseID(ctx, shop.LicenseID)
		if err != nil {
			log.Printf("Error checking shops: %v", err)
			return err
		}

		exists := false
		for _, existing := range existingShops {
			if existing.Name == shop.Name {
				exists = true
				break
			}
		}

		if !exists {
			// Create the shop
			if err := s.shopRepo.Create(ctx, &shop); err != nil {
				log.Printf("Failed to create shop %s: %v", shop.Name, err)
				return err
			}
			log.Printf("Created shop: %s", shop.Name)
		} else {
			log.Printf("Shop %s already exists", shop.Name)
		}
	}

	return nil
}

// SeedCategories creates initial categories
func (s *InitialDataSeeder) SeedCategories() error {
	ctx := context.Background()

	// Get the license
	license, err := s.licenseRepo.GetBySerialNumber(ctx, "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to get license: %v", err)
		return err
	}

	// Get shops for the license
	shops, err := s.shopRepo.GetByLicenseID(ctx, license.ID)
	if err != nil {
		log.Printf("Failed to get shops: %v", err)
		return err
	}

	if len(shops) == 0 {
		log.Printf("No shops found")
		return nil
	}

	shop := shops[0] // Use the first shop

	categories := []entities.Category{
		{
			ShopID: shop.ID,
			Name:   "Food & Beverages",
		},
		{
			ShopID: shop.ID,
			Name:   "Electronics",
		},
		{
			ShopID: shop.ID,
			Name:   "Clothing",
		},
		{
			ShopID: shop.ID,
			Name:   "Household",
		},
	}

	for _, category := range categories {
		// Check if category already exists by name and shop
		existing, err := s.categoryRepo.GetByName(ctx, category.Name, category.ShopID)
		if err == gorm.ErrRecordNotFound {
			// Create the category
			if err := s.categoryRepo.Create(ctx, &category); err != nil {
				log.Printf("Failed to create category %s: %v", category.Name, err)
				return err
			}
			log.Printf("Created category: %s", category.Name)
		} else if err != nil {
			log.Printf("Error checking category %s: %v", category.Name, err)
			return err
		} else {
			log.Printf("Category %s already exists", existing.Name)
		}
	}

	return nil
}

// SeedProducts creates initial products
func (s *InitialDataSeeder) SeedProducts() error {
	ctx := context.Background()

	// Get the demo shop
	license, err := s.licenseRepo.GetBySerialNumber(ctx, "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to get license: %v", err)
		return err
	}

	shops, err := s.shopRepo.GetByLicenseID(ctx, license.ID)
	if err != nil {
		log.Printf("Failed to get shops: %v", err)
		return err
	}

	if len(shops) == 0 {
		log.Printf("No shops found")
		return nil
	}

	shop := shops[0]

	// Get categories
	categories, err := s.categoryRepo.GetByShopID(ctx, shop.ID)
	if err != nil {
		log.Printf("Failed to get categories: %v", err)
		return err
	}

	if len(categories) == 0 {
		log.Printf("No categories found")
		return nil
	}

	// Map category names to IDs
	categoryMap := make(map[string]uuid.UUID)
	for _, cat := range categories {
		categoryMap[cat.Name] = cat.ID
	}

	// Get category IDs
	foodCatID := categoryMap["Food & Beverages"]
	electronicsCatID := categoryMap["Electronics"]
	clothingCatID := categoryMap["Clothing"]
	householdCatID := categoryMap["Household"]

	products := []entities.Product{
		{
			ShopID:      shop.ID,
			CatID:       &foodCatID,
			Name:        "Nasi Goreng Special",
			Barcode:     strPtr("1234567890123"),
			Unit:        strPtr("portion"),
			Sale:        25000,
			Buy:         15000,
			Stock:       50,
			IsHaveStock: true,
		},
		{
			ShopID:      shop.ID,
			CatID:       &foodCatID,
			Name:        "Ayam Bakar",
			Barcode:     strPtr("1234567890124"),
			Unit:        strPtr("portion"),
			Sale:        30000,
			Buy:         18000,
			Stock:       30,
			IsHaveStock: true,
		},
		{
			ShopID:      shop.ID,
			CatID:       &electronicsCatID,
			Name:        "Wireless Mouse",
			Barcode:     strPtr("9876543210987"),
			Unit:        strPtr("pcs"),
			Sale:        150000,
			Buy:         100000,
			Stock:       20,
			IsHaveStock: true,
		},
		{
			ShopID:      shop.ID,
			CatID:       &electronicsCatID,
			Name:        "USB Cable",
			Barcode:     strPtr("9876543210988"),
			Unit:        strPtr("pcs"),
			Sale:        25000,
			Buy:         15000,
			Stock:       100,
			IsHaveStock: true,
		},
		{
			ShopID:      shop.ID,
			CatID:       &clothingCatID,
			Name:        "T-Shirt Basic",
			Barcode:     strPtr("5556667778889"),
			Unit:        strPtr("pcs"),
			Sale:        75000,
			Buy:         45000,
			Stock:       25,
			IsHaveStock: true,
		},
		{
			ShopID:      shop.ID,
			CatID:       &householdCatID,
			Name:        "Laundry Detergent",
			Barcode:     strPtr("1112223334445"),
			Unit:        strPtr("bottle"),
			Sale:        35000,
			Buy:         25000,
			Stock:       15,
			IsHaveStock: true,
		},
	}

	for _, product := range products {
		// Check if product already exists by name
		// Get all products for the shop and check names
		existingProducts, err := s.productRepo.GetByShopID(ctx, product.ShopID)
		if err != nil {
			log.Printf("Error checking products: %v", err)
			return err
		}

		exists := false
		for _, existing := range existingProducts {
			if existing.Name == product.Name {
				exists = true
				break
			}
		}

		if !exists {
			// Create the product
			if err := s.productRepo.Create(ctx, &product); err != nil {
				log.Printf("Failed to create product %s: %v", product.Name, err)
				return err
			}
			log.Printf("Created product: %s", product.Name)
		} else {
			log.Printf("Product %s already exists", product.Name)
		}
	}

	return nil
}

// SeedUserDomains creates initial user domains for super admin and admin
func (s *InitialDataSeeder) SeedUserDomains() error {
	ctx := context.Background()

	// Get users
	superAdmin, err := s.userRepo.GetByEmail(ctx, "superadmin@example.com")
	if err != nil {
		log.Printf("Failed to get super admin user: %v", err)
		return err
	}

	admin, err := s.userRepo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Printf("Failed to get admin user: %v", err)
		return err
	}

	// Get license for domain
	license, err := s.licenseRepo.GetBySerialNumber(ctx, "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to get license: %v", err)
		return err
	}

	domain := license.SerialNumber

	// Create user domains
	userDomains := []entities.UserDomain{
		{
			UserID: superAdmin.ID,
			Domain: "*", // Super admin gets global domain access
		},
		{
			UserID: superAdmin.ID,
			Domain: domain, // Super admin also gets specific domain access
		},
		{
			UserID: admin.ID,
			Domain: "*", // Admin gets global domain access
		},
		{
			UserID: admin.ID,
			Domain: domain, // Admin also gets specific domain access
		},
	}

	for _, userDomain := range userDomains {
		// Check if user domain already exists
		existing, err := s.userDomainRepo.GetByUserAndDomain(ctx, userDomain.UserID, userDomain.Domain)
		if err == gorm.ErrRecordNotFound {
			// Create the user domain
			if err := s.userDomainRepo.Create(ctx, &userDomain); err != nil {
				log.Printf("Failed to create user domain for user %s: %v", userDomain.UserID, err)
				return err
			}
			log.Printf("Created user domain for user %s in domain %s", userDomain.UserID, userDomain.Domain)
		} else if err != nil {
			log.Printf("Error checking user domain for user %s: %v", userDomain.UserID, err)
			return err
		} else {
			log.Printf("User domain for user %s in domain %s already exists", existing.UserID, existing.Domain)
		}
	}

	return nil
}

// SeedCasbinRules creates initial Casbin rules for super admin and admin
func (s *InitialDataSeeder) SeedCasbinRules() error {
	ctx := context.Background()

	// Get users
	superAdmin, err := s.userRepo.GetByEmail(ctx, "superadmin@example.com")
	if err != nil {
		log.Printf("Failed to get super admin user: %v", err)
		return err
	}

	admin, err := s.userRepo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Printf("Failed to get admin user: %v", err)
		return err
	}

	// Get license for domain
	license, err := s.licenseRepo.GetBySerialNumber(ctx, "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to get license: %v", err)
		return err
	}

	domain := license.SerialNumber

	// Assign roles to users in domain
	// Super admin gets super_admin role in global domain "*" and specific domain
	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", "*")
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in global domain: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in global domain *")

	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", domain)
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in domain %s: %v", domain, err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in domain %s", domain)

	// Admin gets admin role in global domain and specific domain
	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", "*")
	if err != nil {
		log.Printf("Failed to assign admin role to admin in global domain: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin in global domain *")

	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", domain)
	if err != nil {
		log.Printf("Failed to assign admin role to admin in domain %s: %v", domain, err)
		return err
	}
	log.Printf("Assigned admin role to admin in domain %s", domain)

	return nil
}

// SeedAll runs all initial data seeders
func (s *InitialDataSeeder) SeedAll() error {
	log.Println("Starting initial data seeding...")

	if err := s.SeedLicenses(); err != nil {
		return err
	}

	if err := s.SeedUsers(); err != nil {
		return err
	}

	if err := s.SeedUserRoles(); err != nil {
		return err
	}

	if err := s.SeedShops(); err != nil {
		return err
	}

	if err := s.SeedCategories(); err != nil {
		return err
	}

	if err := s.SeedProducts(); err != nil {
		return err
	}

	if err := s.SeedUserDomains(); err != nil {
		return err
	}

	if err := s.SeedCasbinRules(); err != nil {
		return err
	}

	log.Println("Initial data seeding completed successfully")
	return nil
}
