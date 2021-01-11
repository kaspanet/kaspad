package lrucacheuint64tohash

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// LRUCache is a least-recently-used cache from
// uint64 to DomainHash
type LRUCache struct {
	cache    map[uint64]*externalapi.DomainHash
	capacity int
}

// New creates a new LRUCache
func New(capacity int) *LRUCache {
	return &LRUCache{
		cache:    make(map[uint64]*externalapi.DomainHash, capacity+1),
		capacity: capacity,
	}
}

// Add adds an entry to the LRUCache
func (c *LRUCache) Add(key uint64, value *externalapi.DomainHash) {
	c.cache[key] = value

	if len(c.cache) > c.capacity {
		c.evictRandom()
	}
}

// Get returns the entry for the given key, or (nil, false) otherwise
func (c *LRUCache) Get(key uint64) (*externalapi.DomainHash, bool) {
	value, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	return value, true
}

// Has returns whether the LRUCache contains the given key
func (c *LRUCache) Has(key uint64) bool {
	_, ok := c.cache[key]
	return ok
}

// Remove removes the entry for the the given key. Does nothing if
// the entry does not exist
func (c *LRUCache) Remove(key uint64) {
	delete(c.cache, key)
}

func (c *LRUCache) evictRandom() {
	var keyToEvict uint64
	for key := range c.cache {
		keyToEvict = key
		break
	}
	c.Remove(keyToEvict)
}
