package flowcontext

import (
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// SharedRequestedBlocks is a data structure that is shared between peers that
// holds the hashes of all the requested blocks to prevent redundant requests.
type SharedRequestedBlocks struct {
	blocks map[externalapi.DomainHash]struct{}
	sync.Mutex
}

// Remove removes a block from the set.
func (s *SharedRequestedBlocks) Remove(hash *externalapi.DomainHash) {
	s.Lock()
	defer s.Unlock()
	delete(s.blocks, *hash)
}

// RemoveSet removes a set of blocks from the set.
func (s *SharedRequestedBlocks) RemoveSet(blockHashes map[externalapi.DomainHash]struct{}) {
	s.Lock()
	defer s.Unlock()
	for hash := range blockHashes {
		delete(s.blocks, hash)
	}
}

// AddIfNotExists adds a block to the set if it doesn't exist yet.
func (s *SharedRequestedBlocks) AddIfNotExists(hash *externalapi.DomainHash) (exists bool) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.blocks[*hash]
	if ok {
		return true
	}
	s.blocks[*hash] = struct{}{}
	return false
}

// NewSharedRequestedBlocks returns a new instance of SharedRequestedBlocks.
func NewSharedRequestedBlocks() *SharedRequestedBlocks {
	return &SharedRequestedBlocks{
		blocks: make(map[externalapi.DomainHash]struct{}),
	}
}
