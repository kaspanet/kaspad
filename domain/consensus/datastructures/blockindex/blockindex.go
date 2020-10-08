package blockindex

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockIndex represents a store of known block hashes
type BlockIndex struct {
}

// New instantiates a new BlockIndex
func New() *BlockIndex {
	return &BlockIndex{}
}

// Insert inserts the given blockHash
func (bi *BlockIndex) Insert(dbTx model.TxContextProxy, blockHash *model.DomainHash) {

}

// Exists returns whether the given blockHash exists in the store
func (bi *BlockIndex) Exists(dbContext model.ContextProxy, blockHash *model.DomainHash) bool {
	return false
}
