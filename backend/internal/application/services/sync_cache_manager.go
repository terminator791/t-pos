package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/terminator791/t-pos/internal/domain/entities"
	"gorm.io/gorm"
)

// SyncCacheManager handles intelligent caching for sync operations
type SyncCacheManager struct {
	db         *gorm.DB
	cache      *sync.Map
	config     SyncCacheConfig
	stats      SyncCacheStats
	statsMutex sync.RWMutex
}

// SyncCacheConfig contains configuration for sync caching
type SyncCacheConfig struct {
	EnableCaching         bool          `json:"enable_caching"`
	ShopLicenseCacheTTL   time.Duration `json:"shop_license_cache_ttl"`
	ProductShopCacheTTL   time.Duration `json:"product_shop_cache_ttl"`
	UserShopsCacheTTL     time.Duration `json:"user_shops_cache_ttl"`
	MaxCacheEntries       int           `json:"max_cache_entries"`
	CacheCleanupInterval  time.Duration `json:"cache_cleanup_interval"`
	EnableCacheStatistics bool          `json:"enable_cache_statistics"`
}

// SyncCacheStats tracks cache performance metrics
type SyncCacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Evictions   int64     `json:"evictions"`
	Entries     int64     `json:"entries"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
}

// ShopLicenseCacheEntry caches shop-to-license mappings
type ShopLicenseCacheEntry struct {
	ShopID    uuid.UUID
	LicenseID uuid.UUID
	Valid     bool
}

// ProductShopCacheEntry caches product-to-shop mappings
type ProductShopCacheEntry struct {
	ProductID uuid.UUID
	ShopID    uuid.UUID
	Valid     bool
}

// UserShopsCacheEntry caches user accessible shops
type UserShopsCacheEntry struct {
	UserID  uuid.UUID
	ShopIDs []uuid.UUID
}

// NewSyncCacheManager creates a new sync cache manager
func NewSyncCacheManager(db *gorm.DB, config SyncCacheConfig) *SyncCacheManager {
	manager := &SyncCacheManager{
		db:     db,
		cache:  &sync.Map{},
		config: config,
		stats: SyncCacheStats{
			LastCleanup: time.Now(),
		},
	}

	// Start background cleanup goroutine if caching is enabled
	if config.EnableCaching && config.CacheCleanupInterval > 0 {
		go manager.startCleanupWorker()
	}

	return manager
}

// GetShopLicenseMapping retrieves or caches shop-license mappings for bulk validation
func (c *SyncCacheManager) GetShopLicenseMapping(ctx context.Context, licenseID uuid.UUID) (map[uuid.UUID]bool, error) {
	if !c.config.EnableCaching {
		return c.fetchShopLicenseMappingFromDB(ctx, licenseID)
	}

	cacheKey := fmt.Sprintf("shop_license:%s", licenseID.String())

	// Try to get from cache first
	if cached, found := c.cache.Load(cacheKey); found {
		if entry, ok := cached.(*CacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
			c.recordCacheHit()
			if mapping, ok := entry.Data.(map[uuid.UUID]bool); ok {
				return mapping, nil
			}
		}
		// Entry expired or invalid, remove it
		c.cache.Delete(cacheKey)
	}

	// Cache miss, fetch from database
	c.recordCacheMiss()
	mapping, err := c.fetchShopLicenseMappingFromDB(ctx, licenseID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	entry := &CacheEntry{
		Data:      mapping,
		ExpiresAt: time.Now().Add(c.config.ShopLicenseCacheTTL),
		CreatedAt: time.Now(),
	}
	c.cache.Store(cacheKey, entry)
	c.recordCacheEntry()

	return mapping, nil
}

// GetProductShopMapping retrieves or caches product-shop mappings for bulk validation
func (c *SyncCacheManager) GetProductShopMapping(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if !c.config.EnableCaching || len(productIDs) == 0 {
		return c.fetchProductShopMappingFromDB(ctx, productIDs)
	}

	productToShop := make(map[uuid.UUID]uuid.UUID)
	uncachedProductIDs := make([]uuid.UUID, 0)

	// Check cache for each product
	for _, productID := range productIDs {
		cacheKey := fmt.Sprintf("product_shop:%s", productID.String())

		if cached, found := c.cache.Load(cacheKey); found {
			if entry, ok := cached.(*CacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
				if productShopEntry, ok := entry.Data.(*ProductShopCacheEntry); ok && productShopEntry.Valid {
					productToShop[productID] = productShopEntry.ShopID
					c.recordCacheHit()
					continue
				}
			}
			// Entry expired or invalid, remove it
			c.cache.Delete(cacheKey)
		}

		// Add to list of uncached products
		uncachedProductIDs = append(uncachedProductIDs, productID)
		c.recordCacheMiss()
	}

	// Fetch uncached products from database
	if len(uncachedProductIDs) > 0 {
		dbMapping, err := c.fetchProductShopMappingFromDB(ctx, uncachedProductIDs)
		if err != nil {
			return nil, err
		}

		// Cache the results and merge with cached data
		for productID, shopID := range dbMapping {
			productToShop[productID] = shopID

			// Cache this mapping
			cacheKey := fmt.Sprintf("product_shop:%s", productID.String())
			entry := &CacheEntry{
				Data: &ProductShopCacheEntry{
					ProductID: productID,
					ShopID:    shopID,
					Valid:     true,
				},
				ExpiresAt: time.Now().Add(c.config.ProductShopCacheTTL),
				CreatedAt: time.Now(),
			}
			c.cache.Store(cacheKey, entry)
			c.recordCacheEntry()
		}

		// Cache negative results (products that don't exist)
		for _, productID := range uncachedProductIDs {
			if _, exists := dbMapping[productID]; !exists {
				cacheKey := fmt.Sprintf("product_shop:%s", productID.String())
				entry := &CacheEntry{
					Data: &ProductShopCacheEntry{
						ProductID: productID,
						Valid:     false,
					},
					ExpiresAt: time.Now().Add(c.config.ProductShopCacheTTL),
					CreatedAt: time.Now(),
				}
				c.cache.Store(cacheKey, entry)
				c.recordCacheEntry()
			}
		}
	}

	return productToShop, nil
}

// GetUserAccessibleShops retrieves or caches user accessible shops
func (c *SyncCacheManager) GetUserAccessibleShops(ctx context.Context, userID uuid.UUID, licenseID uuid.UUID) ([]uuid.UUID, error) {
	if !c.config.EnableCaching {
		return c.fetchUserAccessibleShopsFromDB(ctx, userID, licenseID)
	}

	cacheKey := fmt.Sprintf("user_shops:%s:%s", userID.String(), licenseID.String())

	// Try to get from cache first
	if cached, found := c.cache.Load(cacheKey); found {
		if entry, ok := cached.(*CacheEntry); ok && time.Now().Before(entry.ExpiresAt) {
			c.recordCacheHit()
			if userShopsEntry, ok := entry.Data.(*UserShopsCacheEntry); ok {
				return userShopsEntry.ShopIDs, nil
			}
		}
		// Entry expired or invalid, remove it
		c.cache.Delete(cacheKey)
	}

	// Cache miss, fetch from database
	c.recordCacheMiss()
	shopIDs, err := c.fetchUserAccessibleShopsFromDB(ctx, userID, licenseID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	entry := &CacheEntry{
		Data: &UserShopsCacheEntry{
			UserID:  userID,
			ShopIDs: shopIDs,
		},
		ExpiresAt: time.Now().Add(c.config.UserShopsCacheTTL),
		CreatedAt: time.Now(),
	}
	c.cache.Store(cacheKey, entry)
	c.recordCacheEntry()

	return shopIDs, nil
}

// InvalidateShopLicenseCache invalidates shop-license cache entries
func (c *SyncCacheManager) InvalidateShopLicenseCache(licenseID uuid.UUID) {
	if !c.config.EnableCaching {
		return
	}

	cacheKey := fmt.Sprintf("shop_license:%s", licenseID.String())
	c.cache.Delete(cacheKey)
}

// InvalidateProductShopCache invalidates product-shop cache entries
func (c *SyncCacheManager) InvalidateProductShopCache(productIDs []uuid.UUID) {
	if !c.config.EnableCaching {
		return
	}

	for _, productID := range productIDs {
		cacheKey := fmt.Sprintf("product_shop:%s", productID.String())
		c.cache.Delete(cacheKey)
	}
}

// InvalidateUserShopsCache invalidates user shops cache entries
func (c *SyncCacheManager) InvalidateUserShopsCache(userID uuid.UUID) {
	if !c.config.EnableCaching {
		return
	}

	// Pattern-based deletion for all license combinations
	c.cache.Range(func(key, value interface{}) bool {
		if strKey, ok := key.(string); ok {
			userPrefix := fmt.Sprintf("user_shops:%s:", userID.String())
			if len(strKey) > len(userPrefix) && strKey[:len(userPrefix)] == userPrefix {
				c.cache.Delete(key)
			}
		}
		return true
	})
}

// GetCacheStats returns current cache statistics
func (c *SyncCacheManager) GetCacheStats() SyncCacheStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	return c.stats
}

// ClearCache clears all cache entries
func (c *SyncCacheManager) ClearCache() {
	c.cache = &sync.Map{}
	c.statsMutex.Lock()
	c.stats.Entries = 0
	c.statsMutex.Unlock()
}

// Database fetch methods

func (c *SyncCacheManager) fetchShopLicenseMappingFromDB(ctx context.Context, licenseID uuid.UUID) (map[uuid.UUID]bool, error) {
	var shops []entities.Shop
	err := c.db.WithContext(ctx).
		Select("id").
		Where("license_id = ?", licenseID).
		Find(&shops).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch shop license mapping: %w", err)
	}

	mapping := make(map[uuid.UUID]bool)
	for _, shop := range shops {
		mapping[shop.ID] = true
	}

	return mapping, nil
}

func (c *SyncCacheManager) fetchProductShopMappingFromDB(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	if len(productIDs) == 0 {
		return make(map[uuid.UUID]uuid.UUID), nil
	}

	var products []entities.Product
	err := c.db.WithContext(ctx).
		Select("id, shop_id").
		Where("id IN (?)", productIDs).
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch product shop mapping: %w", err)
	}

	mapping := make(map[uuid.UUID]uuid.UUID)
	for _, product := range products {
		mapping[product.ID] = product.ShopID
	}

	return mapping, nil
}

func (c *SyncCacheManager) fetchUserAccessibleShopsFromDB(ctx context.Context, userID uuid.UUID, licenseID uuid.UUID) ([]uuid.UUID, error) {
	// This is a simplified implementation - in practice, this would involve
	// more complex role-based access control logic
	var shops []entities.Shop
	err := c.db.WithContext(ctx).
		Select("id").
		Where("license_id = ?", licenseID).
		Find(&shops).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch user accessible shops: %w", err)
	}

	shopIDs := make([]uuid.UUID, len(shops))
	for i, shop := range shops {
		shopIDs[i] = shop.ID
	}

	return shopIDs, nil
}

// Cache statistics and cleanup methods

func (c *SyncCacheManager) recordCacheHit() {
	if c.config.EnableCacheStatistics {
		c.statsMutex.Lock()
		c.stats.Hits++
		c.statsMutex.Unlock()
	}
}

func (c *SyncCacheManager) recordCacheMiss() {
	if c.config.EnableCacheStatistics {
		c.statsMutex.Lock()
		c.stats.Misses++
		c.statsMutex.Unlock()
	}
}

func (c *SyncCacheManager) recordCacheEntry() {
	if c.config.EnableCacheStatistics {
		c.statsMutex.Lock()
		c.stats.Entries++
		c.statsMutex.Unlock()
	}
}

func (c *SyncCacheManager) recordCacheEviction() {
	if c.config.EnableCacheStatistics {
		c.statsMutex.Lock()
		c.stats.Evictions++
		c.stats.Entries--
		c.statsMutex.Unlock()
	}
}

// startCleanupWorker runs periodic cache cleanup in background
func (c *SyncCacheManager) startCleanupWorker() {
	ticker := time.NewTicker(c.config.CacheCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpiredEntries()
	}
}

// cleanupExpiredEntries removes expired cache entries
func (c *SyncCacheManager) cleanupExpiredEntries() {
	now := time.Now()
	keysToDelete := make([]interface{}, 0)

	c.cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			if now.After(entry.ExpiresAt) {
				keysToDelete = append(keysToDelete, key)
			}
		}
		return true
	})

	// Delete expired entries
	for _, key := range keysToDelete {
		c.cache.Delete(key)
		c.recordCacheEviction()
	}

	// Update cleanup timestamp
	c.statsMutex.Lock()
	c.stats.LastCleanup = now
	c.statsMutex.Unlock()

	// Enforce max cache entries limit
	if c.config.MaxCacheEntries > 0 {
		c.enforceMaxCacheEntries()
	}
}

// enforceMaxCacheEntries removes oldest entries if cache size exceeds limit
func (c *SyncCacheManager) enforceMaxCacheEntries() {
	type cacheItem struct {
		key       interface{}
		entry     *CacheEntry
		createdAt time.Time
	}

	items := make([]cacheItem, 0)
	c.cache.Range(func(key, value interface{}) bool {
		if entry, ok := value.(*CacheEntry); ok {
			items = append(items, cacheItem{
				key:       key,
				entry:     entry,
				createdAt: entry.CreatedAt,
			})
		}
		return true
	})

	if len(items) <= c.config.MaxCacheEntries {
		return
	}

	// Sort by creation time (oldest first)
	// Note: In a production system, you might want to use a more sophisticated
	// eviction policy like LRU (Least Recently Used)
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].createdAt.After(items[j].createdAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Remove oldest entries
	entriesToRemove := len(items) - c.config.MaxCacheEntries
	for i := 0; i < entriesToRemove; i++ {
		c.cache.Delete(items[i].key)
		c.recordCacheEviction()
	}
}
