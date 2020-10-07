package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, acceptanceData *BlockAcceptanceData)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *BlockAcceptanceData
}
