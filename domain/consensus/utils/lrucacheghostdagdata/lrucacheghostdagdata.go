package lrucacheghostdagdata

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

type lruKey struct {
	blockHash  externalapi.DomainHash
	isMetaData bool
}

func newKey(blockHash *externalapi.DomainHash, isMetaData bool) lruKey {
	return lruKey{
		blockHash:  *blockHash,
		isMetaData: isMetaData,
	}
}

// LRUCache is a least-recently-used cache from
// lruKey to *externalapi.BlockGHOSTDAGData
type LRUCache struct {
	cache    map[lruKey]*externalapi.BlockGHOSTDAGData
	capacity int
}

// New creates a new LRUCache
func New(capacity int, preallocate bool) *LRUCache {
	var cache map[lruKey]*externalapi.BlockGHOSTDAGData
	if preallocate {
		cache = make(map[lruKey]*externalapi.BlockGHOSTDAGData, capacity+1)
	} else {
		cache = make(map[lruKey]*externalapi.BlockGHOSTDAGData)
	}
	return &LRUCache{
		cache:    cache,
		capacity: capacity,
	}
}

// Add adds an entry to the LRUCache
func (c *LRUCache) Add(blockHash *externalapi.DomainHash, isMetaData bool, value *externalapi.BlockGHOSTDAGData) {
	key := newKey(blockHash, isMetaData)
	c.cache[key] = value

	if len(c.cache) > c.capacity {
		c.evictRandom()
	}
}

// Get returns the entry for the given key, or (nil, false) otherwise
func (c *LRUCache) Get(blockHash *externalapi.DomainHash, isMetaData bool) (*externalapi.BlockGHOSTDAGData, bool) {
	key := newKey(blockHash, isMetaData)
	value, ok := c.cache[key]
	if !ok {
		return nil, false
	}
	return value, true
}

// Has returns whether the LRUCache contains the given key
func (c *LRUCache) Has(blockHash *externalapi.DomainHash, isMetaData bool) bool {
	key := newKey(blockHash, isMetaData)
	_, ok := c.cache[key]
	return ok
}

// Remove removes the entry for the the given key. Does nothing if
// the entry does not exist
func (c *LRUCache) Remove(blockHash *externalapi.DomainHash, isMetaData bool) {
	key := newKey(blockHash, isMetaData)
	delete(c.cache, key)
}

func (c *LRUCache) evictRandom() {
	var keyToEvict lruKey
	for key := range c.cache {
		keyToEvict = key
		break
	}
	c.Remove(&keyToEvict.blockHash, keyToEvict.isMetaData)
}
