package blockrelay

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type hashesQueueSet struct {
	queue []*externalapi.DomainHash
	set   map[externalapi.DomainHash]struct{}
}

func (r *hashesQueueSet) enqueueIfNotExists(hash *externalapi.DomainHash) {
	if _, ok := r.set[*hash]; ok {
		return
	}
	r.queue = append(r.queue, hash)
	r.set[*hash] = struct{}{}
}

func (r *hashesQueueSet) dequeue(numItems int) []*externalapi.DomainHash {
	var hashes []*externalapi.DomainHash
	hashes, r.queue = r.queue[:numItems], r.queue[numItems:]
	for _, hash := range hashes {
		delete(r.set, *hash)
	}
	return hashes
}

func (r *hashesQueueSet) len() int {
	return len(r.queue)
}

func newHashesQueueSet() *hashesQueueSet {
	return &hashesQueueSet{
		set: make(map[externalapi.DomainHash]struct{}),
	}
}
