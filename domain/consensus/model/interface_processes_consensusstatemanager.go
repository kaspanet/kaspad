package model

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *DomainOutpoint) *UTXOEntry
	CalculateConsensusStateChanges(block *DomainBlock, isDisqualified bool) (
		stateChanges *ConsensusStateChanges, utxoDiffChanges *UTXODiffChanges, virtualGHOSTDAGData *BlockGHOSTDAGData)
	CalculateAcceptanceDataAndMultiset(blockHash *DomainHash) (*BlockAcceptanceData, Multiset)
	Tips() []*DomainHash
	VirtualData() (medianTime int64, blueScore uint64)
	RestoreUTXOSet(blockHash *DomainHash) ReadOnlyUTXOSet
}
