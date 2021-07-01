package lrucachehashpairtoblockghostdagdatahashpair

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type lruKey struct {
	blockHash externalapi.DomainHash
	index     uint64
}

func newKey(blockHash *externalapi.DomainHash, index uint64) lruKey {
	return lruKey{
		blockHash: *blockHash,
		index:     index,
	}
}

// LRUCache is a least-recently-used cache from
// uint64 to DomainHash
type LRUCache struct {
	cache    map[lruKey]*externalapi.BlockGHOSTDAGDataHashPair
	capacity int
}

// New creates a new LRUCache
func New(capacity int, preallocate bool) *LRUCache {
	var cache map[lruKey]*externalapi.BlockGHOSTDAGDataHashPair
	if preallocate {
		cache = make(map[lruKey]*externalapi.BlockGHOSTDAGDataHashPair, capacity+1)
	} else {
		cache = make(map[lruKey]*externalapi.BlockGHOSTDAGDataHashPair)
	}
	return &LRUCache{
		cache:    cache,
		capacity: capacity,
	}
}

// Add adds an entry to the LRUCache
func (c *LRUCache) Add(blockHash *externalapi.DomainHash, index uint64, value *externalapi.BlockGHOSTDAGDataHashPair) {
	key := newKey(blockHash, index)
	c.cache[key] = value

	if len(c.cache) > c.capacity {
		c.evictRandom()
	}
}

// Get returns the entry for the given key, or (nil, false) otherwise
func (c *LRUCache) Get(blockHash *externalapi.DomainHash, index uint64) (*externalapi.BlockGHOSTDAGDataHashPair, bool) {
	key := newKey(blockHash, index)
	value, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	return value, true
}

// Has returns whether the LRUCache contains the given key
func (c *LRUCache) Has(blockHash *externalapi.DomainHash, index uint64) bool {
	key := newKey(blockHash, index)
	_, ok := c.cache[key]
	return ok
}

// Remove removes the entry for the the given key. Does nothing if
// the entry does not exist
func (c *LRUCache) Remove(blockHash *externalapi.DomainHash, index uint64) {
	key := newKey(blockHash, index)
	delete(c.cache, key)
}

func (c *LRUCache) evictRandom() {
	var keyToEvict lruKey
	for key := range c.cache {
		keyToEvict = key
		break
	}
	c.Remove(&keyToEvict.blockHash, keyToEvict.index)
}
