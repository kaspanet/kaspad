package consensusstatestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// ConsensusStateStore ...
type ConsensusStateStore struct {
}

// New ...
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
