package reachabilitytree

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type ReachabilityTree interface {
	AddNode(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	IsInPastOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}
