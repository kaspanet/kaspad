package consensusstatestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore struct {
}

// New instantiates a new ConsensusStateStore
func New() *ConsensusStateStore {
	return &ConsensusStateStore{}
}

// Update updates the store with the given utxoDiff
func (css *ConsensusStateStore) Update(dbTx *dbaccess.TxContext, utxoDiff *model.UTXODiff) {

}

// UTXOByOutpoint gets the utxoEntry associated with the given outpoint
func (css *ConsensusStateStore) UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}
