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

// Stage stages the given utxoDiff for the given blockHash
func (uds *utxoDiffStore) Stage(blockHash *externalapi.DomainHash, utxoDiff *model.UTXODiff, utxoDiffChild *externalapi.DomainHash) {
	panic("implement me")
}

func (uds *utxoDiffStore) IsStaged() bool {
	panic("implement me")
}

func (uds *utxoDiffStore) Discard() {
	panic("implement me")
}

func (uds *utxoDiffStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// UTXODiff gets the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) UTXODiff(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*model.UTXODiff, error) {
	return nil, nil
}

// UTXODiffChild gets the utxoDiff child associated with the given blockHash
func (uds *utxoDiffStore) UTXODiffChild(dbContext model.DBReader, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	return nil, nil
}

// Delete deletes the utxoDiff associated with the given blockHash
func (uds *utxoDiffStore) Delete(dbTx model.DBTransaction, blockHash *externalapi.DomainHash) error {
	return nil
}
