package utxodiffmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// UTXODiffManager provides methods to access
// and store UTXO diffs
type utxoDiffManager struct {
}

// New instantiates a new UTXODiffManager
func New() model.UTXODiffManager {
	return &utxoDiffManager{}
}

// StoreUTXODiff stores the given utxoDiff for the given blockHash
func (u utxoDiffManager) StoreUTXODiff(blockHash *model.DomainHash, utxoDiff *model.UTXODiff) error {
	panic("implement me")
}

// RestoreBlockDiffFromVirtualDiffParent restores the UTXO diff of
// the block for the given blockHash.
func (u utxoDiffManager) RestoreBlockDiffFromVirtualDiffParent(blockHash *model.DomainHash) (utxoDiff *model.UTXODiff, virtualDiffParentHash *model.DomainHash, err error) {
	panic("implement me")
}
