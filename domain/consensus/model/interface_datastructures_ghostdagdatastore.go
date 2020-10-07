package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, blockGHOSTDAGData *BlockGHOSTDAGData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *BlockGHOSTDAGData
}
