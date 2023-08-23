package virtual

import (
	"github.com/c4ei/kaspad/domain/consensus/model"
	"github.com/c4ei/kaspad/domain/consensus/model/externalapi"
)

// ContainsOnlyVirtualGenesis returns whether the given block hashes contain only the virtual
// genesis hash.
func ContainsOnlyVirtualGenesis(blockHashes []*externalapi.DomainHash) bool {
	return len(blockHashes) == 1 && blockHashes[0].Equal(model.VirtualGenesisBlockHash)
}
