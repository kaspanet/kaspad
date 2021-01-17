package utxolrucache

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// LRUCache is a least-recently-used cache for UTXO entries
// indexed by DomainOutpoint
type LRUCache struct {
	cache    map[externalapi.DomainOutpoint]externalapi.UTXOEntry
	capacity int
}

// New creates a new LRUCache
func New(capacity int) *LRUCache {
	return &LRUCache{
		cache:    make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry, capacity+1),
		capacity: capacity,
	}
}

// Add adds an entry to the LRUCache
func (c *LRUCache) Add(key *externalapi.DomainOutpoint, value externalapi.UTXOEntry) {
	c.cache[*key] = value

	if len(c.cache) > c.capacity {
		c.evictRandom()
	}
}

// Get returns the entry for the given key, or (nil, false) otherwise
func (c *LRUCache) Get(key *externalapi.DomainOutpoint) (externalapi.UTXOEntry, bool) {
	value, ok := c.cache[*key]
	if !ok {
		return nil, false
	}
	return value, true
}

// Has returns whether the LRUCache contains the given key
func (c *LRUCache) Has(key *externalapi.DomainOutpoint) bool {
	_, ok := c.cache[*key]
	return ok
}

// Remove removes the entry for the the given key. Does nothing if
// the entry does not exist
func (c *LRUCache) Remove(key *externalapi.DomainOutpoint) {
	delete(c.cache, *key)
}

// Clear clears the cache
func (c *LRUCache) Clear() {
	for key := range c.cache {
		delete(c.cache, key)
	}
}

func (c *LRUCache) evictRandom() {
	var keyToEvict externalapi.DomainOutpoint
	for key := range c.cache {
		keyToEvict = key
		break
	}
	c.Remove(&keyToEvict)
}
