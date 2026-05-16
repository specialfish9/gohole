package dns

import (
	"sync"
	"time"

	"codeberg.org/miekg/dns"
)

type CacheKey struct {
	Name  string
	Type  uint16
	Class uint16
}

func NewCacheKey(question dns.RR) CacheKey {
	return CacheKey{
		Name:  question.Header().Name,
		Type:  dns.RRToType(question),
		Class: question.Header().Class,
	}
}

type CacheEntry struct {
	Answer     dns.RR
	Expiration time.Time
	allowed    bool
}

//go:generate go tool go.uber.org/mock/mockgen -destination=../../mock/dns/cache.go -typed -source=cache.go
type Cache interface {
	// Get retrieves a cached DNS response for the given key.
	// It returns a boolean indicating whether the entry
	// should be allowed, the cached message, and a boolean
	// indicating if the entry was found.
	Get(key CacheKey) (bool, dns.RR, bool)
	SetBlocked(key CacheKey)
	Set(key CacheKey, answer dns.RR, ttl uint32)
}

type cacheImpl struct {
	mu    sync.RWMutex
	items map[CacheKey]*CacheEntry
}

func NewCache() Cache {
	return &cacheImpl{
		items: make(map[CacheKey]*CacheEntry),
	}
}

func (c *cacheImpl) Get(key CacheKey) (bool, dns.RR, bool) {
	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()

	// If the entry is not found return false
	if !ok {
		return false, nil, false
	}

	// Blocked entries do not expire
	if entry.allowed && time.Now().After(entry.Expiration) {
		c.mu.Lock()
		defer c.mu.Unlock()

		// Re-check after acquiring write lock
		entry, ok := c.items[key]
		if !ok || time.Now().After(entry.Expiration) {
			// Entry is expired, remove it from cache and return false
			delete(c.items, key)
			return false, nil, false
		}
	}

	// Entry is valid, return the cached message
	return entry.allowed, entry.Answer, true
}

func (c *cacheImpl) SetBlocked(key CacheKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		allowed: false,
	}
}

func (c *cacheImpl) Set(key CacheKey, answer dns.RR, ttl uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		Answer:     answer,
		Expiration: time.Now().Add(time.Duration(ttl) * time.Second),
		allowed:    true,
	}
}
