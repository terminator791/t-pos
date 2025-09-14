package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SyncLockManager manages distributed locks for sync operations to prevent race conditions
type SyncLockManager struct {
	// In-memory lock storage (in production, this would be Redis or similar)
	locks   map[string]*SyncLock
	locksMu sync.RWMutex
	
	// Configuration
	defaultLockTimeout time.Duration
	cleanupInterval    time.Duration
	
	// Cleanup goroutine control
	stopCleanup chan struct{}
	cleanupWg   sync.WaitGroup
}

// SyncLock represents a distributed lock for sync operations
type SyncLock struct {
	Key        string
	OwnerID    string
	AcquiredAt time.Time
	ExpiresAt  time.Time
	mu         sync.Mutex
	released   bool
}

// SyncLockConfig configures the sync lock manager
type SyncLockConfig struct {
	DefaultLockTimeout time.Duration `json:"default_lock_timeout"` // Default: 30s
	CleanupInterval    time.Duration `json:"cleanup_interval"`     // Default: 10s
	MaxLockHoldTime    time.Duration `json:"max_lock_hold_time"`   // Default: 5m
}

// NewSyncLockManager creates a new sync lock manager
func NewSyncLockManager(config SyncLockConfig) *SyncLockManager {
	if config.DefaultLockTimeout == 0 {
		config.DefaultLockTimeout = 30 * time.Second
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 10 * time.Second
	}
	if config.MaxLockHoldTime == 0 {
		config.MaxLockHoldTime = 5 * time.Minute
	}

	manager := &SyncLockManager{
		locks:              make(map[string]*SyncLock),
		defaultLockTimeout: config.DefaultLockTimeout,
		cleanupInterval:    config.CleanupInterval,
		stopCleanup:        make(chan struct{}),
	}

	// Start cleanup goroutine
	manager.cleanupWg.Add(1)
	go manager.cleanupExpiredLocks()

	return manager
}

// AcquireSyncLock attempts to acquire a distributed lock for sync operations
func (m *SyncLockManager) AcquireSyncLock(ctx context.Context, userID, licenseID uuid.UUID) (*SyncLock, error) {
	lockKey := fmt.Sprintf("sync:%s:%s", userID.String(), licenseID.String())
	ownerID := uuid.New().String()
	
	return m.AcquireLockWithKey(ctx, lockKey, ownerID, m.defaultLockTimeout)
}

// AcquireEntityLock acquires a lock for specific entity operations
func (m *SyncLockManager) AcquireEntityLock(ctx context.Context, entityType string, entityID uuid.UUID) (*SyncLock, error) {
	lockKey := fmt.Sprintf("entity:%s:%s", entityType, entityID.String())
	ownerID := uuid.New().String()
	
	// Shorter timeout for entity-level locks
	timeout := 10 * time.Second
	return m.AcquireLockWithKey(ctx, lockKey, ownerID, timeout)
}

// AcquireLockWithKey attempts to acquire a lock with a specific key
func (m *SyncLockManager) AcquireLockWithKey(ctx context.Context, key, ownerID string, timeout time.Duration) (*SyncLock, error) {
	deadline := time.Now().Add(timeout)
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Try to acquire the lock
		if lock, acquired := m.tryAcquireLock(key, ownerID, timeout); acquired {
			return lock, nil
		}

		// Check if we've exceeded the deadline
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("failed to acquire lock '%s' within timeout %v", key, timeout)
		}

		// Wait a bit before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(100 * time.Millisecond):
			// Continue to next attempt
		}
	}
}

// tryAcquireLock attempts to acquire a lock atomically
func (m *SyncLockManager) tryAcquireLock(key, ownerID string, timeout time.Duration) (*SyncLock, bool) {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	now := time.Now()
	expiresAt := now.Add(timeout)

	// Check if lock already exists and is not expired
	if existingLock, exists := m.locks[key]; exists {
		if now.Before(existingLock.ExpiresAt) {
			// Lock is still valid and held by someone else
			return nil, false
		}
		// Lock has expired, remove it
		delete(m.locks, key)
	}

	// Create new lock
	lock := &SyncLock{
		Key:        key,
		OwnerID:    ownerID,
		AcquiredAt: now,
		ExpiresAt:  expiresAt,
		released:   false,
	}

	m.locks[key] = lock
	return lock, true
}

// ReleaseLock releases a previously acquired lock
func (l *SyncLock) Release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.released {
		return fmt.Errorf("lock %s already released", l.Key)
	}

	l.released = true
	return nil
}

// IsExpired checks if the lock has expired
func (l *SyncLock) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

// ExtendLock extends the lock expiration time
func (m *SyncLockManager) ExtendLock(lock *SyncLock, additionalTime time.Duration) error {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	// Verify the lock still exists and belongs to the same owner
	if existingLock, exists := m.locks[lock.Key]; exists {
		if existingLock.OwnerID == lock.OwnerID && !existingLock.released {
			existingLock.ExpiresAt = existingLock.ExpiresAt.Add(additionalTime)
			lock.ExpiresAt = existingLock.ExpiresAt
			return nil
		}
	}

	return fmt.Errorf("cannot extend lock %s: lock not found or ownership mismatch", lock.Key)
}

// ReleaseLockByKey releases a lock by its key and owner ID
func (m *SyncLockManager) ReleaseLockByKey(key, ownerID string) error {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	if lock, exists := m.locks[key]; exists {
		if lock.OwnerID == ownerID {
			lock.released = true
			delete(m.locks, key)
			return nil
		}
		return fmt.Errorf("cannot release lock %s: ownership mismatch", key)
	}

	return fmt.Errorf("lock %s not found", key)
}

// GetLockInfo returns information about active locks (for monitoring)
func (m *SyncLockManager) GetLockInfo() map[string]SyncLockInfo {
	m.locksMu.RLock()
	defer m.locksMu.RUnlock()

	info := make(map[string]SyncLockInfo)
	for key, lock := range m.locks {
		info[key] = SyncLockInfo{
			Key:        lock.Key,
			OwnerID:    lock.OwnerID,
			AcquiredAt: lock.AcquiredAt,
			ExpiresAt:  lock.ExpiresAt,
			IsExpired:  lock.IsExpired(),
			Released:   lock.released,
		}
	}

	return info
}

// SyncLockInfo provides information about a lock for monitoring
type SyncLockInfo struct {
	Key        string    `json:"key"`
	OwnerID    string    `json:"owner_id"`
	AcquiredAt time.Time `json:"acquired_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsExpired  bool      `json:"is_expired"`
	Released   bool      `json:"released"`
}

// cleanupExpiredLocks runs in a goroutine to clean up expired locks
func (m *SyncLockManager) cleanupExpiredLocks() {
	defer m.cleanupWg.Done()

	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCleanup:
			return
		case <-ticker.C:
			m.performCleanup()
		}
	}
}

// performCleanup removes expired locks from memory
func (m *SyncLockManager) performCleanup() {
	m.locksMu.Lock()
	defer m.locksMu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, lock := range m.locks {
		if now.After(lock.ExpiresAt) || lock.released {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired locks
	for _, key := range expiredKeys {
		delete(m.locks, key)
	}

	if len(expiredKeys) > 0 {
		fmt.Printf("Cleaned up %d expired sync locks\n", len(expiredKeys))
	}
}

// Shutdown gracefully shuts down the lock manager
func (m *SyncLockManager) Shutdown() {
	close(m.stopCleanup)
	m.cleanupWg.Wait()

	// Release all remaining locks
	m.locksMu.Lock()
	defer m.locksMu.Unlock()
	
	for key := range m.locks {
		delete(m.locks, key)
	}
}

// SyncLockContext provides a context-aware wrapper for sync operations with automatic lock management
type SyncLockContext struct {
	ctx     context.Context
	lock    *SyncLock
	manager *SyncLockManager
}

// NewSyncLockContext creates a new sync lock context
func (m *SyncLockManager) NewSyncLockContext(ctx context.Context, userID, licenseID uuid.UUID) (*SyncLockContext, error) {
	lock, err := m.AcquireSyncLock(ctx, userID, licenseID)
	if err != nil {
		return nil, err
	}

	return &SyncLockContext{
		ctx:     ctx,
		lock:    lock,
		manager: m,
	}, nil
}

// Execute runs a function within the locked context
func (lc *SyncLockContext) Execute(fn func(context.Context) error) error {
	defer func() {
		if err := lc.lock.Release(); err != nil {
			fmt.Printf("Warning: failed to release sync lock: %v\n", err)
		}
		// Also remove from manager
		lc.manager.ReleaseLockByKey(lc.lock.Key, lc.lock.OwnerID)
	}()

	return fn(lc.ctx)
}

// GetLock returns the underlying lock (for extending, etc.)
func (lc *SyncLockContext) GetLock() *SyncLock {
	return lc.lock
}

// Extend extends the lock expiration time
func (lc *SyncLockContext) Extend(additionalTime time.Duration) error {
	return lc.manager.ExtendLock(lc.lock, additionalTime)
}