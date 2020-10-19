package utxodiffstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// utxoDiffStore represents a store of UTXODiffs
type utxoDiffStore struct {
}

// New instantiates a new UTXODiffStore
func New() model.UTXODiffStore {
	return &utxoDiffStore{}
}

// Insert inserts the given utxoDiff for the given blockHash
func (uds *utxoDiffStore) Insert(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash, utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {
	return nil
}

// UTXODiff gets the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) UTXODiff(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	return nil, nil
}

// UTXODiffChild gets the utxoDiff child associated with the given blockHash
func (uds *utxoDiffStore) UTXODiffChild(dbContext model.DBContextProxy, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	return nil, nil
}

// Delete deletes the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) Delete(dbTx model.DBTxProxy, blockHash *externalapi.DomainHash) error {
	return nil
}
