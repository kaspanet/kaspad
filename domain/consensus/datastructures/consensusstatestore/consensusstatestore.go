package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore struct {
}

// New instantiates a new ConsensusStateStore
func New() *ConsensusStateStore {
	return &ConsensusStateStore{}
}

// Update updates the store with the given consensusStateChanges
func (css *ConsensusStateStore) Update(dbTx model.DBTxProxy, consensusStateChanges *model.ConsensusStateChanges) {

}

// UTXOByOutpoint gets the utxoEntry associated with the given outpoint
func (css *ConsensusStateStore) UTXOByOutpoint(dbContext model.DBContextProxy, outpoint *model.DomainOutpoint) *model.UTXOEntry {
	return nil
}

// Tips returns the current tips
func (css *ConsensusStateStore) Tips(dbContext model.DBContextProxy) []*model.DomainHash {
	return nil
}
