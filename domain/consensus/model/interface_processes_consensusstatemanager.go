package model

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *DomainOutpoint) (*UTXOEntry, error)
	CalculateConsensusStateChanges(block *DomainBlock, isDisqualified bool) (stateChanges *ConsensusStateChanges,
		utxoDiffChanges *UTXODiffChanges, virtualGHOSTDAGData *BlockGHOSTDAGData, err error)
	CalculateAcceptanceDataAndUTXOMultiset(blockGHOSTDAGData *BlockGHOSTDAGData) (*BlockAcceptanceData, Multiset, error)
	VirtualData() (medianTime int64, blueScore uint64, err error)
	RestorePastUTXOSet(blockHash *DomainHash) (ReadOnlyUTXOSet, error)
}
