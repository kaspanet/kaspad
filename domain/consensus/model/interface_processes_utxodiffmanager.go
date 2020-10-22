package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// UTXODiffManager provides methods to access
// and store UTXO diffs
type UTXODiffManager interface {
	RestoreBlockDiffFromVirtualDiffParent(blockHash *externalapi.DomainHash) (utxoDiff *UTXODiff,
		virtualDiffParentHash *externalapi.DomainHash, err error)
}
