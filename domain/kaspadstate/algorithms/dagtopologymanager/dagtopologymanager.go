package dagtopologymanager

import (
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/daghash"
)

type DAGTopologyManager interface {
	AddBlock(dbTx *dbaccess.TxContext, blockHash *daghash.Hash)
	Parents(blockHash *daghash.Hash) []*daghash.Hash
	Children(blockHash *daghash.Hash) []*daghash.Hash
	IsParentOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsChildOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsAncestorOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
	IsDescendantOf(blockHashA *daghash.Hash, blockHashB *daghash.Hash) bool
}
