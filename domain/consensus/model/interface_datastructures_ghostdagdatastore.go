package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, blockGHOSTDAGData *BlockGHOSTDAGData)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *BlockGHOSTDAGData
}
