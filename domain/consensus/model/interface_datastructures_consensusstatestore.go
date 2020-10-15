package model

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx DBTxProxy, consensusStateChanges *ConsensusStateChanges) error
	UTXOByOutpoint(dbContext DBContextProxy, outpoint *DomainOutpoint) (*UTXOEntry, error)
	Tips(dbContext DBContextProxy) ([]*DomainHash, error)
}
