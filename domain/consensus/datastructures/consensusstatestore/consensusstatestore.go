package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	stagedTips               []*externalapi.DomainHash
	stagedVirtualDiffParents []*externalapi.DomainHash
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
