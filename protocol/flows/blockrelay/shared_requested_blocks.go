package blockrelay

import (
	"sync"

	"github.com/kaspanet/kaspad/util/daghash"
)

// SharedRequestedBlocks is a data structure that is shared between peers that
// holds the hashes of all the requested blocks to prevent redundant requests.
type SharedRequestedBlocks struct {
	blocks map[daghash.Hash]struct{}
	sync.Mutex
}

func (s *SharedRequestedBlocks) remove(hash *daghash.Hash) {
	s.Lock()
	defer s.Unlock()
	delete(s.blocks, *hash)
}

func (s *SharedRequestedBlocks) removeSet(blockHashes map[daghash.Hash]struct{}) {
	s.Lock()
	defer s.Unlock()
	for hash := range blockHashes {
		delete(s.blocks, hash)
	}
}

func (s *SharedRequestedBlocks) addIfNotExists(hash *daghash.Hash) (exists bool) {
	s.Lock()
	defer s.Unlock()
	_, ok := s.blocks[*hash]
	if ok {
		return true
	}
	s.blocks[*hash] = struct{}{}
	return false
}

// NewSharedRequestedBlocks returns a new instance of *SharedRequestedBlocks.
func NewSharedRequestedBlocks() *SharedRequestedBlocks {
	return &SharedRequestedBlocks{
		blocks: make(map[daghash.Hash]struct{}),
	}
}
