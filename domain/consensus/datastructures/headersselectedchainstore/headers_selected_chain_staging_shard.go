package headersselectedchainstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database/binaryserialization"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type headersSelectedChainStagingShard struct {
	store          *headersSelectedChainStore
	addedByHash    map[externalapi.DomainHash]uint64
	removedByHash  map[externalapi.DomainHash]struct{}
	addedByIndex   map[uint64]*externalapi.DomainHash
	removedByIndex map[uint64]struct{}
}

func (hscs *headersSelectedChainStore) stagingShard(stagingArea *model.StagingArea) *headersSelectedChainStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDHeadersSelectedChain, func() model.StagingShard {
		return &headersSelectedChainStagingShard{
			store:          hscs,
			addedByHash:    make(map[externalapi.DomainHash]uint64),
			removedByHash:  make(map[externalapi.DomainHash]struct{}),
			addedByIndex:   make(map[uint64]*externalapi.DomainHash),
			removedByIndex: make(map[uint64]struct{}),
		}
	}).(*headersSelectedChainStagingShard)
}

func (hscss headersSelectedChainStagingShard) Commit(dbTx model.DBTransaction) error {
	if !hscss.isStaged() {
		return nil
	}

	for hash := range hscss.removedByHash {
		hashCopy := hash
		err := dbTx.Delete(hscss.store.hashAsKey(&hashCopy))
		if err != nil {
			return err
		}
		hscss.store.cacheByHash.Remove(&hashCopy)
	}

	for index := range hscss.removedByIndex {
		err := dbTx.Delete(hscss.store.indexAsKey(index))
		if err != nil {
			return err
		}
		hscss.store.cacheByIndex.Remove(index)
	}

	highestIndex := uint64(0)
	for hash, index := range hscss.addedByHash {
		hashCopy := hash
		err := dbTx.Put(hscss.store.hashAsKey(&hashCopy), hscss.store.serializeIndex(index))
		if err != nil {
			return err
		}

		err = dbTx.Put(hscss.store.indexAsKey(index), binaryserialization.SerializeHash(&hashCopy))
		if err != nil {
			return err
		}

		hscss.store.cacheByHash.Add(&hashCopy, index)
		hscss.store.cacheByIndex.Add(index, &hashCopy)

		if index > highestIndex {
			highestIndex = index
		}
	}

	err := dbTx.Put(highestChainBlockIndexKey, hscss.store.serializeIndex(highestIndex))
	if err != nil {
		return err
	}

	hscss.store.cacheHighestChainBlockIndex = highestIndex

	return nil
}

func (hscss headersSelectedChainStagingShard) isStaged() bool {
	return len(hscss.addedByHash) != 0 ||
		len(hscss.removedByHash) != 0 ||
		len(hscss.addedByIndex) != 0 ||
		len(hscss.addedByIndex) != 0
}
