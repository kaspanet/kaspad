package model

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *DomainOutpoint) *UTXOEntry
	CalculateConsensusStateChanges(block *DomainBlock) *ConsensusStateChanges
}
