package consensusstatestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore struct {
}

// New instantiates a new ConsensusStateStore
func New() *ConsensusStateStore {
	return &ConsensusStateStore{}
}

// Update updates the store with the given utxoDiff
func (css *ConsensusStateStore) Update(dbTx model.TxContextProxy, utxoDiff *model.UTXODiff) {

}

// UTXOByOutpoint gets the utxoEntry associated with the given outpoint
func (css *ConsensusStateStore) UTXOByOutpoint(dbContext model.ContextProxy, outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}
