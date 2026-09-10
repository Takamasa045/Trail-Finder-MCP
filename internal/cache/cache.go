package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type item struct {
	payload []byte
	exp     time.Time
}

type Cache struct {
	mu    sync.Mutex
	items map[string]item
	ttl   time.Duration
	max   int
}

func New(ttl time.Duration, max int) *Cache {
	if max <= 0 {
		max = 128
	}
	return &Cache{
		items: make(map[string]item),
		ttl:   ttl,
		max:   max,
	}
}

func (c *Cache) GetJSON(key string, dest any) bool {
	if c == nil || c.ttl <= 0 || key == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	it, ok := c.items[key]
	if !ok || time.Now().After(it.exp) {
		if ok {
			delete(c.items, key)
		}
		return false
	}
	return json.Unmarshal(it.payload, dest) == nil
}

func (c *Cache) SetJSON(key string, val any) {
	if c == nil || c.ttl <= 0 || key == "" || val == nil {
		return
	}
	b, err := json.Marshal(val)
	if err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.items) >= c.max {
		c.evictOldest()
	}
	c.items[key] = item{payload: b, exp: time.Now().Add(c.ttl)}
}

func (c *Cache) Size() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

func (c *Cache) evictOldest() {
	var oldestKey string
	var oldest time.Time
	for k, it := range c.items {
		if oldestKey == "" || it.exp.Before(oldest) {
			oldestKey = k
			oldest = it.exp
		}
	}
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}
