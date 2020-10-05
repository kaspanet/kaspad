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

// UpdateWithDiff ...
func (css *ConsensusStateStore) UpdateWithDiff(dbTx *dbaccess.TxContext, utxoDiff *model.UTXODiff) {

}

// UTXOByOutpoint ...
func (css *ConsensusStateStore) UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}
