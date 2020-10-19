package model

// UTXODiffManager provides methods to access
// and store UTXO diffs
type UTXODiffManager interface {
	StoreUTXODiff(blockHash *DomainHash, utxoDiff *UTXODiff) error
	RestoreBlockDiffFromVirtualDiffParent(blockHash *DomainHash) (utxoDiff *UTXODiff,
		virtualDiffParentHash *DomainHash, err error)
}
