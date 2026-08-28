// Package cache ...
package cache

import (
	"sync"
)

// Cache ...
type Cache[V any] interface {
	Add(key string, value V)
	Get(key string) (V, bool)
	Delete(key string)
	GetMap() map[string]V
	GetDataArray() []V
	ClearCache()
}

// SafeCache ...
type SafeCache[V any] struct {
	mu    sync.Mutex
	order []string
	store map[string]V
}

// NewSafeCache instance of Cache map
func NewSafeCache[V any]() Cache[V] {
	return &SafeCache[V]{
		store: map[string]V{},
	}
}

// Add ...
func (c *SafeCache[V]) Add(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.store[key]
	if !ok {
		c.order = append(c.order, key)
	}
	c.store[key] = value
}

// Get ...
func (c *SafeCache[V]) Get(key string) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.store[key]
	return val, ok
}

// GetMap ...
func (c *SafeCache[V]) GetMap() map[string]V {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.store
}

// GetDataArray ...
func (c *SafeCache[V]) GetDataArray() []V {
	c.mu.Lock()
	defer c.mu.Unlock()
	var data []V
	for _, key := range c.order {
		value, ok := c.store[key]
		if !ok {
			continue
		}
		data = append(data, value)
	}
	return data
}

// Delete ...
func (c *SafeCache[V]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)

	var newOrder = []string{}
	for _, oldKey := range c.order {
		if oldKey == key {
			continue
		}
		newOrder = append(newOrder, oldKey)
	}
	c.order = newOrder
}

// ClearCache ...
func (c *SafeCache[V]) ClearCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.store {
		delete(c.store, k)
	}
}
