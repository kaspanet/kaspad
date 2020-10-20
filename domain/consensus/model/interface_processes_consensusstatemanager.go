package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error)
	CalculateConsensusStateChanges(block *externalapi.DomainBlock, isDisqualified bool) (stateChanges *ConsensusStateChanges,
		utxoDiffChanges *UTXODiffChanges, virtualGHOSTDAGData *BlockGHOSTDAGData, err error)
	VirtualData() (medianTime int64, blueScore uint64, err error)
	RestorePastUTXOSet(blockHash *externalapi.DomainHash) (ReadOnlyUTXOSet, error)
	RestoreDiffFromVirtual(utxoDiff *UTXODiff, virtualDiffParentHash *DomainHash) (*UTXODiff, error)
}
