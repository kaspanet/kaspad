package model

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx DBTxProxy, consensusStateChanges *ConsensusStateChanges)
	UTXOByOutpoint(dbContext DBContextProxy, outpoint *DomainOutpoint) *UTXOEntry
}
