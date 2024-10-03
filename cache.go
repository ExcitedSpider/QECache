// The cache structure that support concurrency
// Encapsulated QECache/lru
package qecache

import (
	"QECache/lru"
	"sync"
)

type cache struct {
	mu       sync.Mutex
	lru      *lru.LRUDict
	maxBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock() // defer will execute even during panic

	if c.lru == nil {
		c.lru = lru.New(c.maxBytes, nil)
	}

	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
