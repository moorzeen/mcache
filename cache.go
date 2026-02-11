package mcache

import (
	"sync"
	"time"
)

type item[V any] struct {
	value      V
	expiryTime time.Time
}

type Cache[K comparable, V any] struct {
	mu    sync.Mutex
	items map[K]item[V]
	ttl   time.Duration
	done  chan struct{}
}

func NewCache[K comparable, V any](ttl time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]item[V]),
		ttl:   ttl,
		done:  make(chan struct{}),
	}

	go c.cleanup()

	return c
}

func (c *Cache[K, V]) Close() {
	close(c.done)
}

func (c *Cache[K, V]) GetAll() map[K]V {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	result := make(map[K]V, len(c.items))

	for k, it := range c.items {
		if now.Before(it.expiryTime) {
			result[k] = it.value
		} else {
			delete(c.items, k)
		}
	}

	return result
}

func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item[V]{
		value:      value,
		expiryTime: time.Now().Add(c.ttl),
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok || time.Now().After(it.expiryTime) {
		delete(c.items, key)
		var zero V
		return zero, false
	}

	return it.value, true
}

func (c *Cache[K, V]) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	count := 0

	for k, it := range c.items {
		if now.Before(it.expiryTime) {
			count++
		} else {
			delete(c.items, k)
		}
	}

	return count
}

func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *Cache[K, V]) Release(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	it, ok := c.items[key]
	if !ok || time.Now().After(it.expiryTime) {
		delete(c.items, key)
		var zero V
		return zero, false
	}

	delete(c.items, key)
	return it.value, true
}

func (c *Cache[K, V]) cleanup() {
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
