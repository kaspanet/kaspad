package blockdag

import (
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/daghash"
)

// ResolveFinalityConflict resolves all finality conflicts by setting an arbitrary finality block, and
// re-selecting virtual parents in such a way that given finalityBlock will be in virtual's selectedParentChain
func (dag *BlockDAG) ResolveFinalityConflict(finalityBlockHash *daghash.Hash) error {
	dag.dagLock.Lock()
	defer dag.dagLock.RUnlock()

	finalityBlock, ok := dag.index.LookupNode(finalityBlockHash)
	if !ok {
		return errors.Errorf("Couldn't find finality block with hash %s", finalityBlockHash)
	}

	err := dag.prepareForFinalityConflictResolution(finalityBlock)
	if err != nil {
		return err
	}

	_, chainUpdates, err := dag.updateVirtualParents(dag.tips, finalityBlock)
	if err != nil {
		return err
	}

	dag.sendNotification(NTChainChanged, ChainChangedNotificationData{
		RemovedChainBlockHashes: chainUpdates.removedChainBlockHashes,
		AddedChainBlockHashes:   chainUpdates.addedChainBlockHashes,
	})
	dag.sendNotification(NTFinalityConflictResolved, FinalityConflictResolvedNotificationData{
		FinalityBlockHash: finalityBlockHash,
		ResolutionTime:    dag.Now(),
	})

	return nil
}

// prepareForFinalityConflictResolution makes sure that the designated selectedTip once a finality conflict is resolved
// is not UTXONotVerified.
func (dag *BlockDAG) prepareForFinalityConflictResolution(finalityBlock *blockNode) error {
	queue := newDownHeap()
	queue.pushSet(dag.tips)

	dbTx, err := dag.databaseContext.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	disqualifiedCandidates := newBlockSet()
	for {
		if queue.Len() == 0 {
			return errors.New("No valid selectedTip candidates")
		}
		candidate := queue.pop()

		isFinalityBlockInSelectedParentChain, err := dag.isInSelectedParentChainOf(finalityBlock, candidate)
		if err != nil {
			return err
		}
		if !isFinalityBlockInSelectedParentChain {
			continue
		}
		if dag.index.BlockNodeStatus(candidate) == statusUTXONotVerified {
			err = dag.resolveNodeStatus(candidate, dbTx)
			if err != nil {
				return err
			}
		}
		if dag.index.BlockNodeStatus(candidate) == statusValid {
			err = dbTx.Commit()
			if err != nil {
				return err
			}
			return nil
		}

		disqualifiedCandidates.add(candidate)

		for parent := range candidate.parents {
			if parent.children.areAllIn(disqualifiedCandidates) {
				queue.Push(parent)
			}
		}
	}
}
