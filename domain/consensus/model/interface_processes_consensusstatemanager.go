package model

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *DomainOutpoint) *UTXOEntry
	CalculateConsensusStateChanges(block *DomainBlock, isDisqualified bool) (
		stateChanges *ConsensusStateChanges, utxoDiffChanges *UTXODiffChanges, virtualGHOSTDAGData *BlockGHOSTDAGData)
	CalculateAcceptanceDataAndUTXOMultiset(blockGHOSTDAGData *BlockGHOSTDAGData) (*BlockAcceptanceData, Multiset)
	Tips() []*DomainHash
	VirtualData() (medianTime int64, blueScore uint64)
	RestorePastUTXOSet(blockHash *DomainHash) ReadOnlyUTXOSet
}
