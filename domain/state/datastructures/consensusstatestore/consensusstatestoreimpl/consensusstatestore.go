package consensusstatestoreimpl

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/domain/state/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type ConsensusStateStore struct {
}

func New() *ConsensusStateStore {
	return &ConsensusStateStore{}
}

func (css *ConsensusStateStore) UpdateWithDiff(dbTx *dbaccess.TxContext, utxoDiff *utxodiffstore.UTXODiff) {

}

func (css *ConsensusStateStore) UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry {
	return nil
}
