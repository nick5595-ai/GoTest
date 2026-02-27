package main

import (
	"sync"
	"sync/atomic"
	"time"
)

// CacheEntry stores a cached value with an expiration time.
type CacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// Cache provides a thread-safe in-memory cache with TTL-based expiration.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
	hits    atomic.Int64
	misses  atomic.Int64
}

// CacheStatsResponse represents the JSON response for cache statistics.
type CacheStatsResponse struct {
	Entries int     `json:"entries"`
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	HitRate float64 `json:"hitRate"`
	TTL     string  `json:"ttl"`
}

// NewCache creates a new Cache with the given TTL for entries.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves a cached value by key. Returns the value and true if found
// and not expired, otherwise nil and false.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		c.misses.Add(1)
		if ok {
			// Expired entry — clean it up
			c.mu.Lock()
			delete(c.entries, key)
			c.mu.Unlock()
		}
		return nil, false
	}

	c.hits.Add(1)
	return entry.value, true
}

// Set stores a value in the cache with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = CacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes all entries from the cache. Called on data mutations.
func (c *Cache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]CacheEntry)
}

// Stats returns current cache statistics.
func (c *Cache) Stats() CacheStatsResponse {
	c.mu.RLock()
	count := len(c.entries)
	c.mu.RUnlock()

	hits := c.hits.Load()
	misses := c.misses.Load()
	total := hits + misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return CacheStatsResponse{
		Entries: count,
		Hits:    hits,
		Misses:  misses,
		HitRate: hitRate,
		TTL:     c.ttl.String(),
	}
}
