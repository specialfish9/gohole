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

func NewCacheKey(msg *dns.Msg) CacheKey {
	question := msg.Question[0]
	return CacheKey{
		Name:  question.Header().Name,
		Type:  dns.RRToType(question),
		Class: question.Header().Class,
	}
}

type CacheEntry struct {
	Answer     []dns.RR
	Expiration time.Time
	allowed    bool
}

type Cache struct {
	mu    sync.RWMutex
	items map[CacheKey]*CacheEntry
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[CacheKey]*CacheEntry),
	}
}

// Get retrieves a cached DNS response for the given key.
// It returns a boolean indicating whether the entry
// should be allowed, the cached message, and a boolean
// indicating if the entry was found.
func (c *Cache) Get(key CacheKey) (bool, []dns.RR, bool) {
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

func (c *Cache) SetBlocked(key CacheKey, msg *dns.Msg) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		allowed: false,
		Answer:  msg.Answer,
	}
}

func (c *Cache) Set(key CacheKey, msg *dns.Msg, ttl uint32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheEntry{
		Answer:     msg.Answer,
		Expiration: time.Now().Add(time.Duration(ttl) * time.Second),
	}
}
