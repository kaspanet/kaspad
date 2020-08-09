package blockrelay

import "github.com/kaspanet/kaspad/util/daghash"

type hashesQueueSet struct {
	queue []*daghash.Hash
	set   map[daghash.Hash]struct{}
}

func (r *hashesQueueSet) enqueueIfNotExists(hash *daghash.Hash) {
	if _, ok := r.set[*hash]; ok {
		return
	}
	r.queue = append(r.queue, hash)
	r.set[*hash] = struct{}{}
}

func (r *hashesQueueSet) dequeue(numItems int) []*daghash.Hash {
	var hashes []*daghash.Hash
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
		set: make(map[daghash.Hash]struct{}),
	}
}
