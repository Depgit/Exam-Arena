package cache

import (
	"strings"
	"sync"
	"time"
)

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]cacheItem
	stop  chan struct{}
}

func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		items: make(map[string]cacheItem),
		stop:  make(chan struct{}),
	}
	go mc.janitor()
	return mc
}

func (mc *MemoryCache) Get(key string) (interface{}, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	item, ok := mc.items[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(item.expiresAt) {
		return nil, false
	}
	return item.value, true
}

func (mc *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (mc *MemoryCache) Delete(key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.items, key)
}

func (mc *MemoryCache) DeletePrefix(prefix string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for k := range mc.items {
		if strings.HasPrefix(k, prefix) {
			delete(mc.items, k)
		}
	}
}

// Add this method to MemoryCache. It returns true only if the key existed
// and was deleted. Used by endMatch to prevent double-ending a match.
func (mc *MemoryCache) DeleteIfPresent(key string) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	item, ok := mc.items[key]
	if !ok {
		return false
	}
	if time.Now().After(item.expiresAt) {
		delete(mc.items, key)
		return false // expired = not present
	}
	delete(mc.items, key)
	return true
}

func (mc *MemoryCache) janitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.mu.Lock()
			now := time.Now()
			for k, v := range mc.items {
				if now.After(v.expiresAt) {
					delete(mc.items, k)
				}
			}
			mc.mu.Unlock()
		case <-mc.stop:
			return
		}
	}
}

func (mc *MemoryCache) Close() {
	close(mc.stop)
}
