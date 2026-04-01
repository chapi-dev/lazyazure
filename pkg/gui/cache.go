package gui

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/matsest/lazyazure/pkg/domain"
)

const (
	rgCacheTTL  = 5 * time.Minute
	resCacheTTL = 3 * time.Minute
	maxRGCache  = 100
	maxResCache = 500
)

// cachedRGs holds cached resource groups for a subscription
type cachedRGs struct {
	groups    []*domain.ResourceGroup
	timestamp time.Time
	cancel    context.CancelFunc
}

// cachedResources holds cached resources for a resource group
type cachedResources struct {
	resources []*domain.Resource
	timestamp time.Time
	cancel    context.CancelFunc
}

// cachedFullResource holds cached full resource details
type cachedFullResource struct {
	resource  *domain.Resource
	timestamp time.Time
	cancel    context.CancelFunc
}

// PreloadCache provides in-memory caching for resource groups and resources
// with TTL-based expiration and size limits
type PreloadCache struct {
	mu             sync.RWMutex
	rgs            map[string]*cachedRGs          // key: subscriptionID
	res            map[string]*cachedResources    // key: "subID/rgName"
	fullRes        map[string]*cachedFullResource // key: resourceID
	rgLimit        int
	resLimit       int
	fullResLimit   int
	rgLoading      map[string]bool // Track in-progress RG preloads
	resLoading     map[string]bool // Track in-progress resource preloads
	fullResLoading map[string]bool // Track in-progress full resource detail loads
}

// NewPreloadCache creates a new preload cache with default limits
func NewPreloadCache() *PreloadCache {
	return &PreloadCache{
		rgs:            make(map[string]*cachedRGs),
		res:            make(map[string]*cachedResources),
		fullRes:        make(map[string]*cachedFullResource),
		rgLimit:        maxRGCache,
		resLimit:       maxResCache,
		fullResLimit:   maxResCache, // Same limit as resources
		rgLoading:      make(map[string]bool),
		resLoading:     make(map[string]bool),
		fullResLoading: make(map[string]bool),
	}
}

// GetRGs retrieves cached resource groups for a subscription
// Returns nil, false if not found or expired
func (c *PreloadCache) GetRGs(subID string) ([]*domain.ResourceGroup, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.rgs[subID]
	if !ok {
		return nil, false
	}

	if c.isExpired(cached.timestamp, rgCacheTTL) {
		return nil, false
	}

	return cached.groups, true
}

// SetRGs stores resource groups for a subscription in cache
// Cancels any existing preload operation for this subscription
func (c *PreloadCache) SetRGs(subID string, groups []*domain.ResourceGroup, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel existing preload if any
	if existing, ok := c.rgs[subID]; ok && existing.cancel != nil {
		existing.cancel()
	}

	// Check if we need to evict
	if len(c.rgs) >= c.rgLimit {
		c.evictOldestRGs(c.rgLimit / 2)
	}

	c.rgs[subID] = &cachedRGs{
		groups:    groups,
		timestamp: time.Now(),
		cancel:    cancel,
	}
	// Clear loading flag
	delete(c.rgLoading, subID)
}

// GetRes retrieves cached resources for a resource group
// Returns nil, false if not found or expired
func (c *PreloadCache) GetRes(subID, rgName string) ([]*domain.Resource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := subID + "/" + rgName
	cached, ok := c.res[key]
	if !ok {
		return nil, false
	}

	if c.isExpired(cached.timestamp, resCacheTTL) {
		return nil, false
	}

	return cached.resources, true
}

// SetRes stores resources for a resource group in cache
// Cancels any existing preload operation for this resource group
func (c *PreloadCache) SetRes(subID, rgName string, resources []*domain.Resource, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := subID + "/" + rgName

	// Cancel existing preload if any
	if existing, ok := c.res[key]; ok && existing.cancel != nil {
		existing.cancel()
	}

	// Check if we need to evict
	if len(c.res) >= c.resLimit {
		c.evictOldestRes(c.resLimit / 2)
	}

	c.res[key] = &cachedResources{
		resources: resources,
		timestamp: time.Now(),
		cancel:    cancel,
	}
	// Clear loading flag
	delete(c.resLoading, key)
}

// IsRGLoading checks if resource groups are currently being loaded for a subscription
func (c *PreloadCache) IsRGLoading(subID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rgLoading[subID]
}

// SetRGLoading sets the loading state for resource groups
func (c *PreloadCache) SetRGLoading(subID string, loading bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if loading {
		c.rgLoading[subID] = true
	} else {
		delete(c.rgLoading, subID)
	}
}

// IsResLoading checks if resources are currently being loaded for a resource group
func (c *PreloadCache) IsResLoading(subID, rgName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := subID + "/" + rgName
	return c.resLoading[key]
}

// SetResLoading sets the loading state for resources
func (c *PreloadCache) SetResLoading(subID, rgName string, loading bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := subID + "/" + rgName
	if loading {
		c.resLoading[key] = true
	} else {
		delete(c.resLoading, key)
	}
}

// GetFullRes retrieves cached full resource details
// Returns nil, false if not found or expired
func (c *PreloadCache) GetFullRes(resourceID string) (*domain.Resource, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.fullRes[resourceID]
	if !ok {
		return nil, false
	}

	if c.isExpired(cached.timestamp, resCacheTTL) {
		return nil, false
	}

	return cached.resource, true
}

// SetFullRes stores full resource details in cache
func (c *PreloadCache) SetFullRes(resourceID string, resource *domain.Resource, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel existing preload if any
	if existing, ok := c.fullRes[resourceID]; ok && existing.cancel != nil {
		existing.cancel()
	}

	// Check if we need to evict
	if len(c.fullRes) >= c.fullResLimit {
		c.evictOldestFullRes(c.fullResLimit / 2)
	}

	c.fullRes[resourceID] = &cachedFullResource{
		resource:  resource,
		timestamp: time.Now(),
		cancel:    cancel,
	}
	// Clear loading flag
	delete(c.fullResLoading, resourceID)
}

// IsFullResLoading checks if full resource details are currently being loaded
func (c *PreloadCache) IsFullResLoading(resourceID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fullResLoading[resourceID]
}

// SetFullResLoading sets the loading state for full resource details
func (c *PreloadCache) SetFullResLoading(resourceID string, loading bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if loading {
		c.fullResLoading[resourceID] = true
	} else {
		delete(c.fullResLoading, resourceID)
	}
}

// InvalidateSubs clears all cached subscriptions, resource groups, and resources
func (c *PreloadCache) InvalidateSubs() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel all active preloads
	for _, cached := range c.rgs {
		if cached.cancel != nil {
			cached.cancel()
		}
	}
	for _, cached := range c.res {
		if cached.cancel != nil {
			cached.cancel()
		}
	}
	for _, cached := range c.fullRes {
		if cached.cancel != nil {
			cached.cancel()
		}
	}

	c.rgs = make(map[string]*cachedRGs)
	c.res = make(map[string]*cachedResources)
	c.fullRes = make(map[string]*cachedFullResource)
	c.rgLoading = make(map[string]bool)
	c.resLoading = make(map[string]bool)
	c.fullResLoading = make(map[string]bool)
}

// InvalidateRGs clears cached resource groups for a subscription and their resources
func (c *PreloadCache) InvalidateRGs(subID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Cancel and remove RG cache
	if cached, ok := c.rgs[subID]; ok {
		if cached.cancel != nil {
			cached.cancel()
		}
		delete(c.rgs, subID)
	}

	// Cancel and remove all resource caches for this subscription
	prefix := subID + "/"
	for key, cached := range c.res {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			if cached.cancel != nil {
				cached.cancel()
			}
			delete(c.res, key)
		}
	}

	// Clear all resource loading flags for this subscription
	for key := range c.resLoading {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.resLoading, key)
		}
	}

	// Clear loading flag for this subscription
	delete(c.rgLoading, subID)
}

// InvalidateRes clears cached resources for a specific resource group
func (c *PreloadCache) InvalidateRes(subID, rgName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := subID + "/" + rgName
	if cached, ok := c.res[key]; ok {
		if cached.cancel != nil {
			cached.cancel()
		}
		delete(c.res, key)
	}
	// Clear loading flag
	delete(c.resLoading, key)
}

// isExpired checks if a timestamp has exceeded the TTL
func (c *PreloadCache) isExpired(timestamp time.Time, ttl time.Duration) bool {
	return time.Since(timestamp) > ttl
}

// evictOldestRGs removes the oldest N resource group cache entries
func (c *PreloadCache) evictOldestRGs(count int) {
	if count <= 0 {
		return
	}

	// Get all keys with timestamps
	type keyTime struct {
		key       string
		timestamp time.Time
	}
	entries := make([]keyTime, 0, len(c.rgs))
	for key, cached := range c.rgs {
		entries = append(entries, keyTime{key, cached.timestamp})
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		key := entries[i].key
		if cached, ok := c.rgs[key]; ok && cached.cancel != nil {
			cached.cancel()
		}
		delete(c.rgs, key)
	}
}

// evictOldestRes removes the oldest N resource cache entries
func (c *PreloadCache) evictOldestRes(count int) {
	if count <= 0 {
		return
	}

	// Get all keys with timestamps
	type keyTime struct {
		key       string
		timestamp time.Time
	}
	entries := make([]keyTime, 0, len(c.res))
	for key, cached := range c.res {
		entries = append(entries, keyTime{key, cached.timestamp})
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		key := entries[i].key
		if cached, ok := c.res[key]; ok && cached.cancel != nil {
			cached.cancel()
		}
		delete(c.res, key)
	}
}

// evictOldestFullRes removes the oldest N full resource cache entries
func (c *PreloadCache) evictOldestFullRes(count int) {
	if count <= 0 {
		return
	}

	// Get all keys with timestamps
	type keyTime struct {
		key       string
		timestamp time.Time
	}
	entries := make([]keyTime, 0, len(c.fullRes))
	for key, cached := range c.fullRes {
		entries = append(entries, keyTime{key, cached.timestamp})
	}

	// Sort by timestamp (oldest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp.Before(entries[j].timestamp)
	})

	// Remove oldest entries
	for i := 0; i < count && i < len(entries); i++ {
		key := entries[i].key
		if cached, ok := c.fullRes[key]; ok && cached.cancel != nil {
			cached.cancel()
		}
		delete(c.fullRes, key)
		delete(c.fullResLoading, key)
	}
}

// GetCacheStats returns current cache statistics (for debugging)
func (c *PreloadCache) GetCacheStats() (rgCount, resCount, fullResCount int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.rgs), len(c.res), len(c.fullRes)
}
