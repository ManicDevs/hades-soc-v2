package auxiliary

import (
	"context"
	"fmt"
	"sync"
	"time"

	"hades-v2/pkg/sdk"
)

// CacheType represents supported cache algorithms
type CacheType string

const (
	CacheLRU CacheType = "lru"
	CacheLFU CacheType = "lfu"
	CacheTTL CacheType = "ttl"
)

// CacheManager provides caching functionality
type CacheManager struct {
	*sdk.BaseModule
	cacheType CacheType
	maxSize   int
	cache     map[string]cacheEntry
	mu        sync.RWMutex
}

type cacheEntry struct {
	value       interface{}
	expiresAt   time.Time
	accessCount int
	lastAccess  time.Time
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager() *CacheManager {
	return &CacheManager{
		BaseModule: sdk.NewBaseModule(
			"cache_manager",
			"Manage caching for performance optimization",
			sdk.CategoryReporting,
		),
		cacheType: CacheLRU,
		maxSize:   1000,
		cache:     make(map[string]cacheEntry),
	}
}

// Execute initializes the cache manager
func (cm *CacheManager) Execute(ctx context.Context) error {
	cm.SetStatus(sdk.StatusRunning)
	defer cm.SetStatus(sdk.StatusIdle)

	if err := cm.validateConfig(); err != nil {
		return fmt.Errorf("hades.auxiliary.cache_manager: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		cm.SetStatus(sdk.StatusCompleted)
		return nil
	}
}

// SetCacheType configures the cache algorithm
func (cm *CacheManager) SetCacheType(cacheType CacheType) error {
	switch cacheType {
	case CacheLRU, CacheLFU, CacheTTL:
		cm.cacheType = cacheType
		return nil
	default:
		return fmt.Errorf("hades.auxiliary.cache_manager: invalid cache type: %s", cacheType)
	}
}

// SetMaxSize configures the maximum cache size
func (cm *CacheManager) SetMaxSize(size int) error {
	if size <= 0 {
		return fmt.Errorf("hades.auxiliary.cache_manager: size must be positive")
	}
	cm.maxSize = size
	return nil
}

// GetResult returns cache status
func (cm *CacheManager) GetResult() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return fmt.Sprintf("Cache manager initialized: type=%s, size=%d/%d",
		cm.cacheType, len(cm.cache), cm.maxSize)
}

// validateConfig ensures cache configuration is valid
func (cm *CacheManager) validateConfig() error {
	if cm.cacheType == "" {
		return fmt.Errorf("hades.auxiliary.cache_manager: cache type not configured")
	}
	if cm.maxSize <= 0 {
		return fmt.Errorf("hades.auxiliary.cache_manager: max size not configured")
	}
	return nil
}
