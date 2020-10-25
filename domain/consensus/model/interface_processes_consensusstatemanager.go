package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	CalculateConsensusStateChanges(blockHash *externalapi.DomainHash) error
	PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error
	VirtualData() (medianTime int64, blueScore uint64, err error)
	RestorePastUTXOSet(blockHash *externalapi.DomainHash) (ReadOnlyUTXOSet, error)
	RestoreDiffFromVirtual(utxoDiff *UTXODiff, virtualDiffParentHash *externalapi.DomainHash) (*UTXODiff, error)
}
