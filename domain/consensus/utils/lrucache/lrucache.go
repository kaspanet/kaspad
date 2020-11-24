package lrucache

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type LRUCache struct {
	cache    map[externalapi.DomainHash]interface{}
	capacity int
}

func New(capacity int) *LRUCache {
	return &LRUCache{
		cache:    make(map[externalapi.DomainHash]interface{}, capacity+1),
		capacity: capacity,
	}
}

func (c *LRUCache) Add(key *externalapi.DomainHash, data interface{}) {
	c.cache[*key] = data

	if len(c.cache) > c.capacity {
		c.evictRandom()
	}
}

func (c *LRUCache) Get(key *externalapi.DomainHash) (interface{}, bool) {
	data, ok := c.cache[*key]
	if !ok {
		return nil, false
	}
	return data, true
}

func (c *LRUCache) Has(key *externalapi.DomainHash) bool {
	_, ok := c.cache[*key]
	return ok
}

func (c *LRUCache) Remove(key *externalapi.DomainHash) {
	delete(c.cache, *key)
}

func (c *LRUCache) evictRandom() {
	var keyToEvict externalapi.DomainHash
	for key := range c.cache {
		keyToEvict = key
	}
	c.Remove(&keyToEvict)
}
