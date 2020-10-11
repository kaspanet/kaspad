package model

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	UTXOByOutpoint(outpoint *DomainOutpoint) *UTXOEntry
	ValidateTransaction(transaction *DomainTransaction, utxoEntries []*UTXOEntry) error
	CalculateConsensusStateChanges(block *DomainBlock) *ConsensusStateChanges
}
