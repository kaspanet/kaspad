package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type pruningStagingShard struct {
	store *pruningStore

	pruningPointByIndex              map[uint64]*externalapi.DomainHash
	pruningPointIndex                *uint64
	newPruningPointCandidate         *externalapi.DomainHash
	startUpdatingPruningPointUTXOSet bool
}

func (ps *pruningStore) stagingShard(stagingArea *model.StagingArea) *pruningStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDPruning, func() model.StagingShard {
		return &pruningStagingShard{
			store:                            ps,
			pruningPointByIndex:              map[uint64]*externalapi.DomainHash{},
			newPruningPointCandidate:         nil,
			startUpdatingPruningPointUTXOSet: false,
		}
	}).(*pruningStagingShard)
}

func (mss *pruningStagingShard) Commit(dbTx model.DBTransaction) error {
	for index, hash := range mss.pruningPointByIndex {
		hashCopy := hash
		hashBytes, err := mss.store.serializeHash(hash)
		if err != nil {
			return err
		}
		err = dbTx.Put(mss.store.indexAsKey(index), hashBytes)
		if err != nil {
			return err
		}
		mss.store.pruningPointByIndexCache.Add(index, hashCopy)
	}

	if mss.pruningPointIndex != nil {
		indexBytes := mss.store.serializeIndex(*mss.pruningPointIndex)
		err := dbTx.Put(mss.store.pruningBlockIndexKey, indexBytes)
		if err != nil {
			return err
		}

		if mss.store.pruningPointIndexCache == nil {
			var zero uint64
			mss.store.pruningPointIndexCache = &zero
		}

		*mss.store.pruningPointIndexCache = *mss.pruningPointIndex
	}

	if mss.newPruningPointCandidate != nil {
		candidateBytes, err := mss.store.serializeHash(mss.newPruningPointCandidate)
		if err != nil {
			return err
		}
		err = dbTx.Put(mss.store.candidatePruningPointHashKey, candidateBytes)
		if err != nil {
			return err
		}
		mss.store.pruningPointCandidateCache = mss.newPruningPointCandidate
	}

	if mss.startUpdatingPruningPointUTXOSet {
		err := dbTx.Put(mss.store.updatingPruningPointUTXOSetKey, []byte{0})
		if err != nil {
			return err
		}
	}

	return nil
}

func (mss *pruningStagingShard) isStaged() bool {
	return len(mss.pruningPointByIndex) > 0 || mss.newPruningPointCandidate != nil || mss.startUpdatingPruningPointUTXOSet
}
