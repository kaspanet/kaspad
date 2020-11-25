package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	tipsStaging               []*externalapi.DomainHash
	virtualDiffParentsStaging []*externalapi.DomainHash
	virtualUTXODiffStaging    *model.UTXODiff
	virtualUTXOSetStaging     model.UTXOCollection

	tipsCache               []*externalapi.DomainHash
	virtualDiffParentsCache []*externalapi.DomainHash
}

// New instantiates a new ConsensusStateStore
func New() model.ConsensusStateStore {
	return &consensusStateStore{}
}

func (css *consensusStateStore) Discard() {
	css.tipsStaging = nil
	css.virtualUTXODiffStaging = nil
	css.virtualDiffParentsStaging = nil
	css.virtualUTXOSetStaging = nil
}

func (css *consensusStateStore) Commit(dbTx model.DBTransaction) error {
	err := css.commitTips(dbTx)
	if err != nil {
		return err
	}
	err = css.commitVirtualDiffParents(dbTx)
	if err != nil {
		return err
	}

	err = css.commitVirtualUTXODiff(dbTx)
	if err != nil {
		return err
	}

	err = css.commitVirtualUTXOSet(dbTx)
	if err != nil {
		return err
	}

	css.Discard()

	return nil
}

func (css *consensusStateStore) IsStaged() bool {
	return css.tipsStaging != nil ||
		css.virtualDiffParentsStaging != nil ||
		css.virtualUTXODiffStaging != nil
}
