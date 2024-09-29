package lru

import "container/list"

// Simple cache. Not safe for concurrent access
type Cache struct {
	maxBytes  int64
	usedBytes int64
	// Use doubly linked list to implement LRU
	ll *list.List
	// Cache data as a map
	cache map[string]*list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

// the entry to save in cache
//
// TODO: extend this to support complex data structure
type entry struct {
	key   string
	value Value
}

// The value saved in the entry. Allow arbitrary type in principle.
//
// TODO: maybe use generics to introduce type
type Value interface {
	Len() int // how many bytes it takes
}

func New(maxBites int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBites,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get an entry as a method on Cache
func (c *Cache) Get(key string) (value Value, ok bool) {
	cacheNode, ok := c.cache[key]
	if ok {
		// Update the the accessed element to the front
		// So that we keep track of recently usage
		c.ll.MoveToFront(cacheNode)
		entry := cacheNode.Value.(*entry)
		return entry.value, true
	}

	// naked return as (nil, false)
	return
}

func (c *Cache) RemoveRLU() {
	if rluEle := c.ll.Back(); rluEle != nil {
		// need to remove the element from both dict and list

		c.ll.Remove(rluEle)
		entry := rluEle.Value.(*entry)

		if _, ok := c.cache[entry.key]; ok {
			// remove from dict
			delete(c.cache, entry.key)
		}

		c.usedBytes -= int64(len(entry.key)) + int64(entry.value.Len())

		if c.OnEvicted != nil {
			c.OnEvicted(entry.key, entry.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// if there exists such a key
		// update the value and used bytes
		entry := ele.Value.(*entry)
		entry.value = value
		c.usedBytes += int64(value.Len()) - int64(entry.value.Len())

		// update its frequency
		c.ll.MoveToFront(ele)
	} else {
		// insert a new element
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.usedBytes += int64(len(key)) + int64(value.Len())
	}

	// Perform delete if overload
	//
	// Caveat: no strong guarantee that it always holds that { c.maxBytes >= c.usedBytes }
	// TODO: implement strong guarantee
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveRLU()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
