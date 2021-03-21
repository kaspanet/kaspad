package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxolrucache"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	virtualUTXOSetCache *utxolrucache.LRUCache

	tipsCache []*externalapi.DomainHash
}

// New instantiates a new ConsensusStateStore
func New(utxoSetCacheSize int, preallocate bool) model.ConsensusStateStore {
	return &consensusStateStore{
		virtualUTXOSetCache: utxolrucache.New(utxoSetCacheSize, preallocate),
	}
}

func (css *consensusStateStore) IsStaged(stagingArea *model.StagingArea) bool {
	stagingShard := css.stagingShard(stagingArea)

	return stagingShard.tipsStaging != nil || stagingShard.virtualUTXODiffStaging != nil
}
