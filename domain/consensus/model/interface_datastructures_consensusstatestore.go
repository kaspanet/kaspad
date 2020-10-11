package model

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx DBTxProxy, utxoDiff *UTXODiff)
	UTXOByOutpoint(dbContext DBContextProxy, outpoint *DomainOutpoint) *UTXOEntry
}
