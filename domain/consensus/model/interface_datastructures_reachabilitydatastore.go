package model

import (
	"github.com/kaspanet/kaspad/util/daghash"
)

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *daghash.Hash, reachabilityData *ReachabilityData)
	Get(dbContext ContextProxy, blockHash *daghash.Hash) *ReachabilityData
}
