package cache

import (
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type Cache struct {
	cache    map[externalapi.DomainHash]*cacheEntry
	capacity int
}

type cacheEntry struct {
	timestamp time.Time
	data      interface{}
}

func New(capacity int) *Cache {
	return &Cache{
		cache:    make(map[externalapi.DomainHash]*cacheEntry, capacity),
		capacity: capacity,
	}
}

func (c *Cache) Add(key *externalapi.DomainHash, data interface{}) {
	c.cache[*key] = &cacheEntry{
		timestamp: time.Now(),
		data:      data,
	}

	if len(c.cache) > c.capacity {
		c.evict()
	}
}

func (c *Cache) Get(key *externalapi.DomainHash) (interface{}, bool) {
	entry, ok := c.cache[*key]
	if !ok {
		return nil, false
	}

	entry.timestamp = time.Now()

	return entry.data, true
}

func (c *Cache) Remove(key *externalapi.DomainHash) {
	delete(c.cache, *key)
}

func (c *Cache) evict() {
	for len(c.cache) > c.capacity {
		var oldestKey *externalapi.DomainHash
		oldestTimestamp := time.Now()

		for key, entry := range c.cache {
			if entry.timestamp.Before(oldestTimestamp) {
				oldestTimestamp = entry.timestamp
				oldestKey = &key
			}
		}

		c.Remove(oldestKey)
	}
}
