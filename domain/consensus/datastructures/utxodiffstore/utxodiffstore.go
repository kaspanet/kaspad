package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// UTXODiffStore represents a store of UTXODiffs
type UTXODiffStore struct {
}

// New instantiates a new UTXODiffStore
func New() *UTXODiffStore {
	return &UTXODiffStore{}
}

// Insert inserts the given utxoDiff for the given blockHash
func (uds *UTXODiffStore) Insert(dbTx model.DBTxProxy, blockHash *model.DomainHash, utxoDiff *model.UTXODiff, utxoDiffChild *model.DomainHash) {

}

// Get gets the utxoDiff associated with the given blockHash
func (uds *UTXODiffStore) Get(dbContext model.DBContextProxy, blockHash *model.DomainHash) *model.UTXODiff {
	return nil
}
func (uds *UTXODiffStore) Delete(dbTx model.DBTxProxy, blockHash *model.DomainHash) {

}
