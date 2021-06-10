package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxolrucache"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	virtualUTXOSetCache *utxolrucache.LRUCache
	tipsCache           []*externalapi.DomainHash
	tipsKey             model.DBKey
	utxoSetBucket       model.DBBucket
}

// New instantiates a new ConsensusStateStore
func New(prefix byte, utxoSetCacheSize int, preallocate bool) model.ConsensusStateStore {
	return &consensusStateStore{
		virtualUTXOSetCache: utxolrucache.New(utxoSetCacheSize, preallocate),
		tipsKey:             database.MakeBucket([]byte{prefix}).Key(tipsKeyName),
		utxoSetBucket:       database.MakeBucket([]byte{prefix}).Bucket(utxoSetBucketName),
	}
}

func (css *consensusStateStore) IsStaged(stagingArea *model.StagingArea) bool {
	return css.stagingShard(stagingArea).isStaged()
}
