package model

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx *dbaccess.TxContext, blockHash *daghash.Hash, acceptanceData *BlockAcceptanceData)
	Get(dbContext dbaccess.Context, blockHash *daghash.Hash) *BlockAcceptanceData
}
