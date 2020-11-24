package consensusstatestore

import (
	"github.com/kaspanet/golang-lru/simplelru"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// consensusStateStore represents a store for the current consensus state
type consensusStateStore struct {
	stagedTips               []*externalapi.DomainHash
	stagedVirtualDiffParents []*externalapi.DomainHash
	stagedVirtualUTXODiff    *model.UTXODiff
	stagedVirtualUTXOSet     model.UTXOCollection
	cache                    simplelru.LRUCache
}

// New instantiates a new ConsensusStateStore
func New(cacheSize int) (model.ConsensusStateStore, error) {
	consensusStateStore := &consensusStateStore{}

	cache, err := simplelru.NewLRU(cacheSize, nil)
	if err != nil {
		return nil, err
	}
	consensusStateStore.cache = cache

	return consensusStateStore, nil
}

func (c *consensusStateStore) Discard() {
	c.stagedTips = nil
	c.stagedVirtualUTXODiff = nil
	c.stagedVirtualDiffParents = nil
	c.stagedVirtualUTXOSet = nil
}

func (c *consensusStateStore) Commit(dbTx model.DBTransaction) error {
	err := c.commitTips(dbTx)
	if err != nil {
		return err
	}
	err = c.commitVirtualDiffParents(dbTx)
	if err != nil {
		return err
	}

	err = c.commitVirtualUTXODiff(dbTx)
	if err != nil {
		return err
	}

	err = c.commitVirtualUTXOSet(dbTx)
	if err != nil {
		return err
	}

	c.Discard()

	return nil
}

func (c *consensusStateStore) IsStaged() bool {
	return c.stagedTips != nil ||
		c.stagedVirtualDiffParents != nil ||
		c.stagedVirtualUTXODiff != nil
}
