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

// Fixed UUID constants for consistent seeding
var (
	// Licenses
	License1ID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	License2ID = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	License3ID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

	// Users
	SuperAdminID = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	Admin1ID     = uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	Owner1ID     = uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	Owner2ID     = uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	Cashier1ID   = uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	Cashier2ID   = uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")

	// Shops
	Shop1ID = uuid.MustParse("11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	Shop2ID = uuid.MustParse("22222222-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	Shop3ID = uuid.MustParse("33333333-cccc-cccc-cccc-cccccccccccc")

	// Categories
	FoodCat1ID     = uuid.MustParse("aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa")
	ElectronicsCat1ID = uuid.MustParse("bbbbbbbb-1111-1111-1111-bbbbbbbbbbbb")
	ClothingCat1ID = uuid.MustParse("cccccccc-1111-1111-1111-cccccccccccc")
	HouseholdCat1ID = uuid.MustParse("dddddddd-1111-1111-1111-dddddddddddd")

	FoodCat2ID     = uuid.MustParse("aaaaaaaa-2222-2222-2222-aaaaaaaaaaaa")
	ElectronicsCat2ID = uuid.MustParse("bbbbbbbb-2222-2222-2222-bbbbbbbbbbbb")
	ClothingCat2ID = uuid.MustParse("cccccccc-2222-2222-2222-cccccccccccc")
	HouseholdCat2ID = uuid.MustParse("dddddddd-2222-2222-2222-dddddddddddd")

	FoodCat3ID     = uuid.MustParse("aaaaaaaa-3333-3333-3333-aaaaaaaaaaaa")
	ElectronicsCat3ID = uuid.MustParse("bbbbbbbb-3333-3333-3333-bbbbbbbbbbbb")
	ClothingCat3ID = uuid.MustParse("cccccccc-3333-3333-3333-cccccccccccc")
	HouseholdCat3ID = uuid.MustParse("dddddddd-3333-3333-3333-dddddddddddd")

	// Products
	Product1ID = uuid.MustParse("11111111-1111-aaaa-aaaa-aaaaaaaaaaaa")
	Product2ID = uuid.MustParse("22222222-1111-aaaa-aaaa-bbbbbbbbbbbb")
	Product3ID = uuid.MustParse("33333333-1111-aaaa-aaaa-cccccccccccc")
	Product4ID = uuid.MustParse("44444444-1111-aaaa-aaaa-dddddddddddd")

	Product5ID = uuid.MustParse("11111111-2222-bbbb-bbbb-aaaaaaaaaaaa")
	Product6ID = uuid.MustParse("22222222-2222-bbbb-bbbb-bbbbbbbbbbbb")
	Product7ID = uuid.MustParse("33333333-2222-bbbb-bbbb-cccccccccccc")
	Product8ID = uuid.MustParse("44444444-2222-bbbb-bbbb-dddddddddddd")

	Product9ID  = uuid.MustParse("11111111-3333-cccc-cccc-aaaaaaaaaaaa")
	Product10ID = uuid.MustParse("22222222-3333-cccc-cccc-bbbbbbbbbbbb")
	Product11ID = uuid.MustParse("33333333-3333-cccc-cccc-cccccccccccc")
	Product12ID = uuid.MustParse("44444444-3333-cccc-cccc-dddddddddddd")
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
	policyRepo      repositories.PolicyRepository
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
	policyRepo repositories.PolicyRepository,
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
		policyRepo:      policyRepo,
		enforcerService: enforcerService,
	}
}

// SeedLicenses creates initial licenses
func (s *InitialDataSeeder) SeedLicenses() error {
	ctx := context.Background()

	licenses := []entities.License{
		{
			ID:           License1ID,
			SerialNumber: "LIC-001-DEMO",
		},
		{
			ID:           License2ID,
			SerialNumber: "LIC-002-DEMO",
		},
		{
			ID:           License3ID,
			SerialNumber: "LIC-003-DEMO",
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

// SeedBasicUsers creates initial users without shop references for cashiers
func (s *InitialDataSeeder) SeedBasicUsers() error {
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
			ID:       SuperAdminID,
			Email:    strPtr("superadmin@example.com"),
			Username: strPtr("superadmin"),
			Name:     "Super Admin",
			Pin:      strPtr(string(hashedPin)),
			Password: string(hashedPassword),
		},
		{
			ID:       Admin1ID,
			Email:    strPtr("admin@example.com"),
			Username: strPtr("admin"),
			Name:     "Admin User",
			Pin:      strPtr(string(hashedPin)),
			Password: string(hashedPassword),
		},
		{
			ID:         Owner1ID,
			LicenseID:  &License1ID,
			Email:      strPtr("owner1@example.com"),
			Username:   strPtr("owner1"),
			Name:       "Owner Business 1",
			Pin:        strPtr(string(hashedPin)),
			Password:   string(hashedPassword),
		},
		{
			ID:         Owner2ID,
			LicenseID:  &License2ID,
			Email:      strPtr("owner2@example.com"),
			Username:   strPtr("owner2"),
			Name:       "Owner Business 2",
			Pin:        strPtr(string(hashedPin)),
			Password:   string(hashedPassword),
		},
		// Cashiers without shop references initially
		{
			ID:         Cashier1ID,
			LicenseID:  &License1ID,
			// ShopID will be set later
			Email:      strPtr("cashier1@example.com"),
			Username:   strPtr("cashier1"),
			Name:       "Cashier 1",
			Pin:        strPtr(string(hashedPin)),
			Password:   string(hashedPassword),
		},
		{
			ID:         Cashier2ID,
			LicenseID:  &License2ID,
			// ShopID will be set later
			Email:      strPtr("cashier2@example.com"),
			Username:   strPtr("cashier2"),
			Name:       "Cashier 2",
			Pin:        strPtr(string(hashedPin)),
			Password:   string(hashedPassword),
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

// UpdateCashierShopReferences updates cashiers with shop references after shops are created
func (s *InitialDataSeeder) UpdateCashierShopReferences() error {
	ctx := context.Background()

	// Get cashiers
	cashier1, err := s.userRepo.GetByEmail(ctx, "cashier1@example.com")
	if err != nil {
		log.Printf("Failed to get cashier1 user: %v", err)
		return err
	}

	cashier2, err := s.userRepo.GetByEmail(ctx, "cashier2@example.com")
	if err != nil {
		log.Printf("Failed to get cashier2 user: %v", err)
		return err
	}

	// Update cashier1 with shop reference
	cashier1.ShopID = &Shop1ID
	if err := s.userRepo.Update(ctx, cashier1); err != nil {
		log.Printf("Failed to update cashier1 with shop reference: %v", err)
		return err
	}
	log.Printf("Updated cashier1 with shop reference")

	// Update cashier2 with shop reference
	cashier2.ShopID = &Shop2ID
	if err := s.userRepo.Update(ctx, cashier2); err != nil {
		log.Printf("Failed to update cashier2 with shop reference: %v", err)
		return err
	}
	log.Printf("Updated cashier2 with shop reference")

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

	owner1, err := s.userRepo.GetByEmail(ctx, "owner1@example.com")
	if err != nil {
		log.Printf("Failed to get owner1 user: %v", err)
		return err
	}

	owner2, err := s.userRepo.GetByEmail(ctx, "owner2@example.com")
	if err != nil {
		log.Printf("Failed to get owner2 user: %v", err)
		return err
	}

	cashier1, err := s.userRepo.GetByEmail(ctx, "cashier1@example.com")
	if err != nil {
		log.Printf("Failed to get cashier1 user: %v", err)
		return err
	}

	cashier2, err := s.userRepo.GetByEmail(ctx, "cashier2@example.com")
	if err != nil {
		log.Printf("Failed to get cashier2 user: %v", err)
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

	ownerRole, err := s.roleRepo.GetByName(ctx, "owner_business")
	if err != nil {
		log.Printf("Failed to get owner_business role: %v", err)
		return err
	}

	cashierRole, err := s.roleRepo.GetByName(ctx, "cashier")
	if err != nil {
		log.Printf("Failed to get cashier role: %v", err)
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

	owner1.RoleID = &ownerRole.ID
	if err := s.userRepo.Update(ctx, owner1); err != nil {
		log.Printf("Failed to assign role to owner1: %v", err)
		return err
	}
	log.Printf("Assigned owner_business role to owner1 user")

	owner2.RoleID = &ownerRole.ID
	if err := s.userRepo.Update(ctx, owner2); err != nil {
		log.Printf("Failed to assign role to owner2: %v", err)
		return err
	}
	log.Printf("Assigned owner_business role to owner2 user")

	cashier1.RoleID = &cashierRole.ID
	if err := s.userRepo.Update(ctx, cashier1); err != nil {
		log.Printf("Failed to assign role to cashier1: %v", err)
		return err
	}
	log.Printf("Assigned cashier role to cashier1 user")

	cashier2.RoleID = &cashierRole.ID
	if err := s.userRepo.Update(ctx, cashier2); err != nil {
		log.Printf("Failed to assign role to cashier2: %v", err)
		return err
	}
	log.Printf("Assigned cashier role to cashier2 user")

	return nil
}

// SeedShops creates initial shops
func (s *InitialDataSeeder) SeedShops() error {
	ctx := context.Background()

	// Get users
	admin, err := s.userRepo.GetByEmail(ctx, "admin@example.com")
	if err != nil {
		log.Printf("Failed to get admin user: %v", err)
		return err
	}

	owner1, err := s.userRepo.GetByEmail(ctx, "owner1@example.com")
	if err != nil {
		log.Printf("Failed to get owner1 user: %v", err)
		return err
	}

	owner2, err := s.userRepo.GetByEmail(ctx, "owner2@example.com")
	if err != nil {
		log.Printf("Failed to get owner2 user: %v", err)
		return err
	}

	shops := []entities.Shop{
		{
			ID:       Shop1ID,
			LicenseID: License1ID,
			UserID:    admin.ID,
			Name:      "Demo Shop Jakarta",
			Address:   strPtr("Jl. Demo No. 123, Jakarta"),
			Slogan:    strPtr("Your Trusted Partner in Jakarta"),
		},
		{
			ID:       Shop2ID,
			LicenseID: License2ID,
			UserID:    owner1.ID,
			Name:      "Demo Shop Bandung",
			Address:   strPtr("Jl. Demo No. 456, Bandung"),
			Slogan:    strPtr("Quality Service in Bandung"),
		},
		{
			ID:       Shop3ID,
			LicenseID: License3ID,
			UserID:    owner2.ID,
			Name:      "Demo Shop Surabaya",
			Address:   strPtr("Jl. Demo No. 789, Surabaya"),
			Slogan:    strPtr("Excellence in Surabaya"),
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

	// Get all shops
	shops, err := s.shopRepo.List(ctx, 1000, 0) // Get up to 1000 shops
	if err != nil {
		log.Printf("Failed to get shops: %v", err)
		return err
	}

	if len(shops) == 0 {
		log.Printf("No shops found")
		return nil
	}

	// Create categories for each shop
	for _, shop := range shops {
		var categories []entities.Category

		switch shop.ID {
		case Shop1ID:
			categories = []entities.Category{
				{
					ID:     FoodCat1ID,
					ShopID: shop.ID,
					Name:   "Food & Beverages",
				},
				{
					ID:     ElectronicsCat1ID,
					ShopID: shop.ID,
					Name:   "Electronics",
				},
				{
					ID:     ClothingCat1ID,
					ShopID: shop.ID,
					Name:   "Clothing",
				},
				{
					ID:     HouseholdCat1ID,
					ShopID: shop.ID,
					Name:   "Household",
				},
			}
		case Shop2ID:
			categories = []entities.Category{
				{
					ID:     FoodCat2ID,
					ShopID: shop.ID,
					Name:   "Food & Beverages",
				},
				{
					ID:     ElectronicsCat2ID,
					ShopID: shop.ID,
					Name:   "Electronics",
				},
				{
					ID:     ClothingCat2ID,
					ShopID: shop.ID,
					Name:   "Clothing",
				},
				{
					ID:     HouseholdCat2ID,
					ShopID: shop.ID,
					Name:   "Household",
				},
			}
		case Shop3ID:
			categories = []entities.Category{
				{
					ID:     FoodCat3ID,
					ShopID: shop.ID,
					Name:   "Food & Beverages",
				},
				{
					ID:     ElectronicsCat3ID,
					ShopID: shop.ID,
					Name:   "Electronics",
				},
				{
					ID:     ClothingCat3ID,
					ShopID: shop.ID,
					Name:   "Clothing",
				},
				{
					ID:     HouseholdCat3ID,
					ShopID: shop.ID,
					Name:   "Household",
				},
			}
		default:
			continue
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
				log.Printf("Created category: %s for shop %s", category.Name, shop.Name)
			} else if err != nil {
				log.Printf("Error checking category %s: %v", category.Name, err)
				return err
			} else {
				log.Printf("Category %s already exists for shop %s", existing.Name, shop.Name)
			}
		}
	}

	return nil
}

// SeedProducts creates initial products
func (s *InitialDataSeeder) SeedProducts() error {
	ctx := context.Background()

	// Get all shops
	shops, err := s.shopRepo.List(ctx, 1000, 0)
	if err != nil {
		log.Printf("Failed to get shops: %v", err)
		return err
	}

	if len(shops) == 0 {
		log.Printf("No shops found")
		return nil
	}

	// Create products for each shop
	for _, shop := range shops {
		// Get categories for this shop
		categories, err := s.categoryRepo.GetByShopID(ctx, shop.ID)
		if err != nil {
			log.Printf("Failed to get categories for shop %s: %v", shop.Name, err)
			return err
		}

		if len(categories) == 0 {
			log.Printf("No categories found for shop %s", shop.Name)
			continue
		}

		// Map category names to IDs
		categoryMap := make(map[string]uuid.UUID)
		for _, cat := range categories {
			categoryMap[cat.Name] = cat.ID
		}

		var products []entities.Product

		switch shop.ID {
		case Shop1ID:
			foodCatID := categoryMap["Food & Beverages"]
			electronicsCatID := categoryMap["Electronics"]
			products = []entities.Product{
				{
					ID:        Product1ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Nasi Goreng Special Jakarta",
					Barcode:   strPtr("1234567890123"),
					Unit:      strPtr("portion"),
					Sale:      25000,
					Buy:       15000,
					Stock:     50,
					IsHaveStock: true,
				},
				{
					ID:        Product2ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Ayam Bakar Jakarta",
					Barcode:   strPtr("1234567890124"),
					Unit:      strPtr("portion"),
					Sale:      30000,
					Buy:       18000,
					Stock:     30,
					IsHaveStock: true,
				},
				{
					ID:        Product3ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "Wireless Mouse Jakarta",
					Barcode:   strPtr("9876543210987"),
					Unit:      strPtr("pcs"),
					Sale:      150000,
					Buy:       100000,
					Stock:     20,
					IsHaveStock: true,
				},
				{
					ID:        Product4ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "USB Cable Jakarta",
					Barcode:   strPtr("9876543210988"),
					Unit:      strPtr("pcs"),
					Sale:      25000,
					Buy:       15000,
					Stock:     100,
					IsHaveStock: true,
				},
			}
		case Shop2ID:
			foodCatID := categoryMap["Food & Beverages"]
			electronicsCatID := categoryMap["Electronics"]
			products = []entities.Product{
				{
					ID:        Product5ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Sate Ayam Bandung",
					Barcode:   strPtr("1111111111111"),
					Unit:      strPtr("portion"),
					Sale:      35000,
					Buy:       20000,
					Stock:     40,
					IsHaveStock: true,
				},
				{
					ID:        Product6ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Es Teh Manis Bandung",
					Barcode:   strPtr("2222222222222"),
					Unit:      strPtr("glass"),
					Sale:      8000,
					Buy:       4000,
					Stock:     80,
					IsHaveStock: true,
				},
				{
					ID:        Product7ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "Bluetooth Speaker Bandung",
					Barcode:   strPtr("3333333333333"),
					Unit:      strPtr("pcs"),
					Sale:      200000,
					Buy:       150000,
					Stock:     15,
					IsHaveStock: true,
				},
				{
					ID:        Product8ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "Power Bank Bandung",
					Barcode:   strPtr("4444444444444"),
					Unit:      strPtr("pcs"),
					Sale:      120000,
					Buy:       80000,
					Stock:     25,
					IsHaveStock: true,
				},
			}
		case Shop3ID:
			foodCatID := categoryMap["Food & Beverages"]
			electronicsCatID := categoryMap["Electronics"]
			products = []entities.Product{
				{
					ID:        Product9ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Rawon Surabaya",
					Barcode:   strPtr("5555555555555"),
					Unit:      strPtr("portion"),
					Sale:      28000,
					Buy:       16000,
					Stock:     35,
					IsHaveStock: true,
				},
				{
					ID:        Product10ID,
					ShopID:    shop.ID,
					CatID:     &foodCatID,
					Name:      "Tahu Tek Surabaya",
					Barcode:   strPtr("6666666666666"),
					Unit:      strPtr("portion"),
					Sale:      15000,
					Buy:       8000,
					Stock:     60,
					IsHaveStock: true,
				},
				{
					ID:        Product11ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "Smartphone Case Surabaya",
					Barcode:   strPtr("7777777777777"),
					Unit:      strPtr("pcs"),
					Sale:      50000,
					Buy:       30000,
					Stock:     45,
					IsHaveStock: true,
				},
				{
					ID:        Product12ID,
					ShopID:    shop.ID,
					CatID:     &electronicsCatID,
					Name:      "Screen Protector Surabaya",
					Barcode:   strPtr("8888888888888"),
					Unit:      strPtr("pcs"),
					Sale:      30000,
					Buy:       15000,
					Stock:     70,
					IsHaveStock: true,
				},
			}
		default:
			continue
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
				log.Printf("Created product: %s for shop %s", product.Name, shop.Name)
			} else {
				log.Printf("Product %s already exists for shop %s", product.Name, shop.Name)
			}
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

	owner1, err := s.userRepo.GetByEmail(ctx, "owner1@example.com")
	if err != nil {
		log.Printf("Failed to get owner1 user: %v", err)
		return err
	}

	owner2, err := s.userRepo.GetByEmail(ctx, "owner2@example.com")
	if err != nil {
		log.Printf("Failed to get owner2 user: %v", err)
		return err
	}

	cashier1, err := s.userRepo.GetByEmail(ctx, "cashier1@example.com")
	if err != nil {
		log.Printf("Failed to get cashier1 user: %v", err)
		return err
	}

	cashier2, err := s.userRepo.GetByEmail(ctx, "cashier2@example.com")
	if err != nil {
		log.Printf("Failed to get cashier2 user: %v", err)
		return err
	}

	// Get shop domains for cashiers (shop domains are "shop-{uuid}")
	shop1, err := s.shopRepo.GetByID(ctx, Shop1ID)
	if err != nil {
		log.Printf("Failed to get shop1: %v", err)
		return err
	}

	shop2, err := s.shopRepo.GetByID(ctx, Shop2ID)
	if err != nil {
		log.Printf("Failed to get shop2: %v", err)
		return err
	}

	// Create user domains
	userDomains := []entities.UserDomain{
		// Super admin gets access to all domains
		{
			UserID: superAdmin.ID,
			Domain: "*", // Global domain access
		},
		{
			UserID: superAdmin.ID,
			Domain: "LIC-001-DEMO",
		},
		{
			UserID: superAdmin.ID,
			Domain: "LIC-002-DEMO",
		},
		{
			UserID: superAdmin.ID,
			Domain: "LIC-003-DEMO",
		},
		{
			UserID: superAdmin.ID,
			Domain: shop1.Domain, // Access to shop1 domain
		},
		{
			UserID: superAdmin.ID,
			Domain: shop2.Domain, // Access to shop2 domain
		},

		// Admin gets access to all domains
		{
			UserID: admin.ID,
			Domain: "*", // Global domain access
		},
		{
			UserID: admin.ID,
			Domain: "LIC-001-DEMO",
		},
		{
			UserID: admin.ID,
			Domain: "LIC-002-DEMO",
		},
		{
			UserID: admin.ID,
			Domain: "LIC-003-DEMO",
		},
		{
			UserID: admin.ID,
			Domain: shop1.Domain, // Access to shop1 domain
		},
		{
			UserID: admin.ID,
			Domain: shop2.Domain, // Access to shop2 domain
		},

		// Owner1 gets access to their license domain
		{
			UserID: owner1.ID,
			Domain: "LIC-001-DEMO",
		},

		// Owner2 gets access to their license domain
		{
			UserID: owner2.ID,
			Domain: "LIC-002-DEMO",
		},

		// Cashier1 gets access to their shop domain only (not license domain)
		{
			UserID: cashier1.ID,
			Domain: shop1.Domain, // shop-{Shop1ID}
		},

		// Cashier2 gets access to their shop domain only (not license domain)
		{
			UserID: cashier2.ID,
			Domain: shop2.Domain, // shop-{Shop2ID}
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

	owner1, err := s.userRepo.GetByEmail(ctx, "owner1@example.com")
	if err != nil {
		log.Printf("Failed to get owner1 user: %v", err)
		return err
	}

	owner2, err := s.userRepo.GetByEmail(ctx, "owner2@example.com")
	if err != nil {
		log.Printf("Failed to get owner2 user: %v", err)
		return err
	}

	cashier1, err := s.userRepo.GetByEmail(ctx, "cashier1@example.com")
	if err != nil {
		log.Printf("Failed to get cashier1 user: %v", err)
		return err
	}

	cashier2, err := s.userRepo.GetByEmail(ctx, "cashier2@example.com")
	if err != nil {
		log.Printf("Failed to get cashier2 user: %v", err)
		return err
	}

	// Get shop domains for cashiers
	shop1, err := s.shopRepo.GetByID(ctx, Shop1ID)
	if err != nil {
		log.Printf("Failed to get shop1: %v", err)
		return err
	}

	shop2, err := s.shopRepo.GetByID(ctx, Shop2ID)
	if err != nil {
		log.Printf("Failed to get shop2: %v", err)
		return err
	}

	// Assign roles to users in domains
	// Super admin gets super_admin role in all domains
	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", "*")
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in global domain: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in global domain *")

	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in LIC-001-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in LIC-001-DEMO")

	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", "LIC-002-DEMO")
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in LIC-002-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in LIC-002-DEMO")

	_, err = s.enforcerService.AddRoleForUser(superAdmin.ID.String(), "super_admin", "LIC-003-DEMO")
	if err != nil {
		log.Printf("Failed to assign super_admin role to super admin in LIC-003-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned super_admin role to super admin in LIC-003-DEMO")

	// Admin gets admin role in all domains
	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", "*")
	if err != nil {
		log.Printf("Failed to assign admin role to admin in global domain: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin in global domain *")

	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to assign admin role to admin in LIC-001-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin in LIC-001-DEMO")

	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", "LIC-002-DEMO")
	if err != nil {
		log.Printf("Failed to assign admin role to admin in LIC-002-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin in LIC-002-DEMO")

	_, err = s.enforcerService.AddRoleForUser(admin.ID.String(), "admin", "LIC-003-DEMO")
	if err != nil {
		log.Printf("Failed to assign admin role to admin in LIC-003-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned admin role to admin in LIC-003-DEMO")

	// Owner1 gets owner_business role in their domain
	_, err = s.enforcerService.AddRoleForUser(owner1.ID.String(), "owner_business", "LIC-001-DEMO")
	if err != nil {
		log.Printf("Failed to assign owner_business role to owner1 in LIC-001-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned owner_business role to owner1 in LIC-001-DEMO")

	// Owner2 gets owner_business role in their domain
	_, err = s.enforcerService.AddRoleForUser(owner2.ID.String(), "owner_business", "LIC-002-DEMO")
	if err != nil {
		log.Printf("Failed to assign owner_business role to owner2 in LIC-002-DEMO: %v", err)
		return err
	}
	log.Printf("Assigned owner_business role to owner2 in LIC-002-DEMO")

	// Cashier1 gets cashier role in their shop domain (not license domain)
	_, err = s.enforcerService.AddRoleForUser(cashier1.ID.String(), "cashier", shop1.Domain)
	if err != nil {
		log.Printf("Failed to assign cashier role to cashier1 in %s: %v", shop1.Domain, err)
		return err
	}
	log.Printf("Assigned cashier role to cashier1 in shop domain %s", shop1.Domain)

	// Cashier2 gets cashier role in their shop domain (not license domain)
	_, err = s.enforcerService.AddRoleForUser(cashier2.ID.String(), "cashier", shop2.Domain)
	if err != nil {
		log.Printf("Failed to assign cashier role to cashier2 in %s: %v", shop2.Domain, err)
		return err
	}
	log.Printf("Assigned cashier role to cashier2 in shop domain %s", shop2.Domain)

	return nil
}

// SeedDomainSpecificPolicies creates domain-specific policies for tenant users
func (s *InitialDataSeeder) SeedDomainSpecificPolicies() error {
	log.Println("Creating domain-specific policies for tenant users...")
	
	// Create auth seeder with policy repository for domain-specific policy creation
	authSeeder := NewAuthSeeder(s.roleRepo, s.policyRepo, s.enforcerService)
	
	// Assign policies for owner_business roles
	if err := authSeeder.AssignPoliciesForRole("owner_business", "LIC-001-DEMO"); err != nil {
		log.Printf("Failed to assign policies for owner_business LIC-001-DEMO: %v", err)
		return err
	}
	log.Printf("✅ Assigned domain-specific policies for owner_business LIC-001-DEMO")
	
	if err := authSeeder.AssignPoliciesForRole("owner_business", "LIC-002-DEMO"); err != nil {
		log.Printf("Failed to assign policies for owner_business LIC-002-DEMO: %v", err)
		return err
	}
	log.Printf("✅ Assigned domain-specific policies for owner_business LIC-002-DEMO")
	
	// Get shop domains for cashier policies
	shop1, err := s.shopRepo.GetByID(context.Background(), Shop1ID)
	if err != nil {
		log.Printf("Failed to get shop1: %v", err)
		return err
	}
	
	shop2, err := s.shopRepo.GetByID(context.Background(), Shop2ID)
	if err != nil {
		log.Printf("Failed to get shop2: %v", err)
		return err
	}
	
	// Assign policies for cashier roles in their shop domains
	if err := authSeeder.AssignPoliciesForRole("cashier", shop1.Domain); err != nil {
		log.Printf("Failed to assign policies for cashier %s: %v", shop1.Domain, err)
		return err
	}
	log.Printf("✅ Assigned domain-specific policies for cashier in shop domain %s", shop1.Domain)
	
	if err := authSeeder.AssignPoliciesForRole("cashier", shop2.Domain); err != nil {
		log.Printf("Failed to assign policies for cashier %s: %v", shop2.Domain, err)
		return err
	}
	log.Printf("✅ Assigned domain-specific policies for cashier in shop domain %s", shop2.Domain)
	
	log.Println("Domain-specific policies creation completed successfully")
	return nil
}

// SeedAll runs all initial data seeders
func (s *InitialDataSeeder) SeedAll() error {
	log.Println("Starting initial data seeding...")

	if err := s.SeedLicenses(); err != nil {
		return err
	}

	// Create basic users first (without shop references for cashiers)
	if err := s.SeedBasicUsers(); err != nil {
		return err
	}

	if err := s.SeedShops(); err != nil {
		return err
	}

	// Now update cashiers with shop references
	if err := s.UpdateCashierShopReferences(); err != nil {
		return err
	}

	if err := s.SeedUserRoles(); err != nil {
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

	// Add domain-specific policy creation for tenant users
	if err := s.SeedDomainSpecificPolicies(); err != nil {
		return err
	}

	log.Println("Initial data seeding completed successfully")
	return nil
}
