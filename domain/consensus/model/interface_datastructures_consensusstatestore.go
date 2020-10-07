package model

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Update(dbTx *dbaccess.TxContext, utxoDiff *UTXODiff)
	UTXOByOutpoint(dbContext dbaccess.Context, outpoint *appmessage.Outpoint) *UTXOEntry
}
