package pruningstore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type pruningStagingShard struct {
	store *pruningStore

	currentPruningPoint              *externalapi.DomainHash
	previousPruningPoint             *externalapi.DomainHash
	newPruningPointCandidate         *externalapi.DomainHash
	startUpdatingPruningPointUTXOSet bool
}

func (ps *pruningStore) stagingShard(stagingArea *model.StagingArea) *pruningStagingShard {
	return stagingArea.GetOrCreateShard(model.StagingShardIDPruning, func() model.StagingShard {
		return &pruningStagingShard{
			store:                            ps,
			currentPruningPoint:              nil,
			previousPruningPoint:             nil,
			newPruningPointCandidate:         nil,
			startUpdatingPruningPointUTXOSet: false,
		}
	}).(*pruningStagingShard)
}

func (mss *pruningStagingShard) Commit(dbTx model.DBTransaction) error {
	if mss.currentPruningPoint != nil {
		pruningPointBytes, err := mss.store.serializeHash(mss.currentPruningPoint)
		if err != nil {
			return err
		}
		err = dbTx.Put(pruningBlockHashKey, pruningPointBytes)
		if err != nil {
			return err
		}
		mss.store.pruningPointCache = mss.currentPruningPoint
	}

	if mss.previousPruningPoint != nil {
		oldPruningPointBytes, err := mss.store.serializeHash(mss.previousPruningPoint)
		if err != nil {
			return err
		}
		err = dbTx.Put(previousPruningBlockHashKey, oldPruningPointBytes)
		if err != nil {
			return err
		}
		mss.store.oldPruningPointCache = mss.previousPruningPoint
	}

	if mss.newPruningPointCandidate != nil {
		candidateBytes, err := mss.store.serializeHash(mss.newPruningPointCandidate)
		if err != nil {
			return err
		}
		err = dbTx.Put(candidatePruningPointHashKey, candidateBytes)
		if err != nil {
			return err
		}
		mss.store.pruningPointCandidateCache = mss.newPruningPointCandidate
	}

	if mss.startUpdatingPruningPointUTXOSet {
		err := dbTx.Put(updatingPruningPointUTXOSetKey, []byte{0})
		if err != nil {
			return err
		}
	}

	return nil
}

func (mss *pruningStagingShard) isStaged() bool {
	return mss.currentPruningPoint != nil || mss.newPruningPointCandidate != nil || mss.previousPruningPoint != nil || mss.startUpdatingPruningPointUTXOSet
}
