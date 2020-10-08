package model

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx TxContextProxy, utxoDiff *UTXODiff)
	UTXOByOutpoint(dbContext ContextProxy, outpoint *DomainOutpoint) *UTXOEntry
}
