package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with metadata.
type CacheEntry[T any] struct {
	// Record is the cached data of type T.
	Record *T
	// CachedAt is the timestamp when the item was cached.
	CachedAt time.Time
	// TTL is the time-to-live duration for this cache entry.
	// After this duration, the entry is considered expired.
	TTL time.Duration
}

// IsExpired checks if the cache entry has expired.
func (e *CacheEntry[T]) IsExpired() bool {
	return time.Since(e.CachedAt) > e.TTL
}

// Cache manages caching of items with TTL support using Go generics.
type Cache[T any] struct {
	entries    map[string]*CacheEntry[T]
	mutex      sync.RWMutex
	defaultTTL time.Duration
}

// NewCache creates a new generic cache with the specified default TTL.
func NewCache[T any](defaultTTL time.Duration) *Cache[T] {
	return &Cache[T]{
		entries:    make(map[string]*CacheEntry[T]),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves an item from cache if it exists and is not expired.
func (c *Cache[T]) Get(id string) (*T, bool, time.Time) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return nil, false, time.Time{}
	}

	if entry.IsExpired() {
		return nil, false, entry.CachedAt
	}

	return entry.Record, true, entry.CachedAt
}

// Set stores an item in cache with the default TTL.
func (c *Cache[T]) Set(id string, record *T) {
	c.setWithTTL(id, record, c.defaultTTL)
}

func (c *Cache[T]) setWithTTL(id string, record *T, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[id] = &CacheEntry[T]{
		Record:   record,
		CachedAt: time.Now(),
		TTL:      ttl,
	}
}

// Remove removes an item from cache by its ID.
func (c *Cache[T]) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, id)
}



// IsExpired checks if a cached item is expired.
func (c *Cache[T]) IsExpired(id string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[id]
	if !exists {
		return true
	}

	return entry.IsExpired()
}