package blockrelay

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

func (s *SharedRequestedBlocks) remove(hash *externalapi.DomainHash) {
	s.Lock()
	defer s.Unlock()
	delete(s.blocks, *hash)
}

func (s *SharedRequestedBlocks) removeSet(blockHashes map[externalapi.DomainHash]struct{}) {
	s.Lock()
	defer s.Unlock()
	for hash := range blockHashes {
		delete(s.blocks, hash)
	}
}

func (s *SharedRequestedBlocks) addIfNotExists(hash *externalapi.DomainHash) (exists bool) {
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
