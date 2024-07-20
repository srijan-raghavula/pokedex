package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	entries map[string]cacheEntry
	mu      *sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	cache := Cache{
		entries: make(map[string]cacheEntry),
		mu:      &sync.Mutex{},
	}
	go cache.reapLoop(interval)
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{
		val:       val,
		createdAt: time.Now().UTC(),
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[key]
	return entry.val, ok
}

func (c *Cache) reapLoop(interval time.Duration) {
	tick := time.NewTicker(interval)
	for range tick.C {
		c.mu.Lock()
		c.mu.Unlock()
		toDelBefore := time.Now().UTC().Add(-interval)
		for k, v := range c.entries {
			if v.createdAt.Before(toDelBefore) {
				delete(c.entries, k)
			}
		}
	}
}
