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

func (c consensusStateStore) Discard() {
	panic("implement me")
}

func (c consensusStateStore) Commit(dbTx model.DBTransaction) error {
	panic("implement me")
}

func (c consensusStateStore) IsStaged() bool {
	panic("implement me")
}

func (c consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff *model.UTXODiff) {
	panic("implement me")
}

func (c consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (*externalapi.UTXOEntry, error) {
	panic("implement me")
}

func (c consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	panic("implement me")
}

func (c consensusStateStore) StageVirtualDiffParents(virtualDiffParents []*externalapi.DomainHash) error {
	panic("implement me")
}

func (c consensusStateStore) VirtualDiffParents(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (c consensusStateStore) Tips(dbContext model.DBReader) ([]*externalapi.DomainHash, error) {
	panic("implement me")
}

func (c consensusStateStore) StageTips(tipHashes []*externalapi.DomainHash) error {
	panic("implement me")
}

func (c consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
	panic("implement me")
}
