package mcache

import (
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiryTime time.Time
}

type Cache struct {
	mu    sync.Mutex
	items map[string]item
	ttl   time.Duration
	done  chan struct{}
}

func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]item),
		ttl:   ttl,
		done:  make(chan struct{}),
	}

	go c.cleanup()

	return c
}

func (c *Cache) Close() {
	close(c.done)
}

func (c *Cache) GetAll() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	result := make(map[string]interface{}, len(c.items))

	for k, it := range c.items {
		if now.Before(it.expiryTime) {
			result[k] = it.value
		}
	}

	return result
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:      value,
		expiryTime: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok || time.Now().After(it.expiryTime) {
		delete(c.items, key)
		return nil, false
	}

	return it.value, true
}

func (c *Cache) Release(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok || time.Now().After(it.expiryTime) {
		delete(c.items, key)
		return nil, false
	}

	delete(c.items, key)
	return it.value, true
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()

			for k, it := range c.items {
				if now.After(it.expiryTime) {
					delete(c.items, k)
				}
			}

			c.mu.Unlock()
		case <-c.done:
			return
		}
	}
}
