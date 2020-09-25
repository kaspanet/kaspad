package consensusstatestore

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/model"
	"github.com/kaspanet/kaspad/domain/state/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type ConsensusStateStore interface {
	UpdateWithDiff(dbTx *dbaccess.TxContext, utxoDiff *utxodiffstore.UTXODiff)
	UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *model.UTXOEntry
}
