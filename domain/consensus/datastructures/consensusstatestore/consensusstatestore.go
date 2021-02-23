package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxolrucache"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	tipsStaging               []*externalapi.DomainHash
	virtualDiffParentsStaging []*externalapi.DomainHash
	virtualUTXODiffStaging    externalapi.UTXODiff

	virtualUTXOSetCache *utxolrucache.LRUCache

	tipsCache               []*externalapi.DomainHash
	virtualDiffParentsCache []*externalapi.DomainHash
}

// New instantiates a new ConsensusStateStore
func New(utxoSetCacheSize int, preallocate bool) model.ConsensusStateStore {
	return &consensusStateStore{
		virtualUTXOSetCache: utxolrucache.New(utxoSetCacheSize, preallocate),
	}
}

func (css *consensusStateStore) Discard() {
	css.tipsStaging = nil
	css.virtualUTXODiffStaging = nil
	css.virtualDiffParentsStaging = nil
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

	css.Discard()

	return nil
}

func (css *consensusStateStore) IsStaged() bool {
	return css.tipsStaging != nil ||
		css.virtualDiffParentsStaging != nil ||
		css.virtualUTXODiffStaging != nil
}
