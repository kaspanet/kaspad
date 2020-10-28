package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
}

// New instantiates a new ConsensusStateStore
func New() model.ConsensusStateStore {
	return &consensusStateStore{}
}

// Stage stages the store with the given consensusStateChanges
func (css *consensusStateStore) Stage(consensusStateChanges *model.ConsensusStateChanges) {
	panic("implement me")
}

func (css *consensusStateStore) IsStaged() bool {
	panic("implement me")
}

func (css *consensusStateStore) Discard() {
	panic("implement me")
}

func (css *consensusStateStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

// UTXOByOutpoint gets the utxoEntry associated with the given outpoint
func (css *consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error) {
	return nil, nil
}
