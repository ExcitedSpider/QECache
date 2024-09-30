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

// by default, get has no lock to protect it to improve efficiency
// could have dirty data
func (c *cache) get(key string) (value ByteView, ok bool) {
	if c.lru == nil {
		return
	}

	v, o := c.lru.Get(key)

	return v.(ByteView), o
}

// the synchronized version of get
func (c *cache) getSync(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.get(key)
}
