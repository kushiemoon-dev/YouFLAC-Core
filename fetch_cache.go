package core

import (
	"container/list"
	"sync"
	"time"
)

// FetchCache is a thread-safe LRU cache with per-entry TTL.
// It stores *VideoInfo keyed by URL or video id.
type FetchCache struct {
	mu       sync.Mutex
	capacity int
	ttl      time.Duration
	ll       *list.List
	items    map[string]*list.Element
}

type fetchCacheEntry struct {
	key       string
	value     *VideoInfo
	expiresAt time.Time
}

// NewFetchCache returns an empty LRU cache.
func NewFetchCache(capacity int, ttl time.Duration) *FetchCache {
	if capacity <= 0 {
		capacity = 128
	}
	return &FetchCache{
		capacity: capacity,
		ttl:      ttl,
		ll:       list.New(),
		items:    make(map[string]*list.Element, capacity),
	}
}

// Get returns the cached value if present and not expired.
func (c *FetchCache) Get(key string) (*VideoInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil, false
	}
	entry := el.Value.(*fetchCacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.ll.Remove(el)
		delete(c.items, key)
		return nil, false
	}
	c.ll.MoveToFront(el)
	return entry.value, true
}

// Put inserts or refreshes an entry.
func (c *FetchCache) Put(key string, value *VideoInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		entry := el.Value.(*fetchCacheEntry)
		entry.value = value
		entry.expiresAt = time.Now().Add(c.ttl)
		c.ll.MoveToFront(el)
		return
	}
	entry := &fetchCacheEntry{key: key, value: value, expiresAt: time.Now().Add(c.ttl)}
	el := c.ll.PushFront(entry)
	c.items[key] = el
	if c.ll.Len() > c.capacity {
		oldest := c.ll.Back()
		if oldest != nil {
			c.ll.Remove(oldest)
			delete(c.items, oldest.Value.(*fetchCacheEntry).key)
		}
	}
}

// Invalidate drops the entry for key.
func (c *FetchCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.ll.Remove(el)
		delete(c.items, key)
	}
}

// Size returns the current number of entries.
func (c *FetchCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

// defaultFetchCache is a package-level cache used by GetVideoMetadata
// when FetchCacheEnabled is true.
var defaultFetchCache = NewFetchCache(256, time.Hour)

// ConfigureFetchCache reconfigures the package-level cache from config values.
func ConfigureFetchCache(enabled bool, ttlSeconds int) {
	if !enabled {
		defaultFetchCache = NewFetchCache(1, time.Nanosecond)
		return
	}
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	defaultFetchCache = NewFetchCache(256, ttl)
}
