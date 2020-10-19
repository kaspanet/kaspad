package blockdag

import (
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/blocknode"
	"github.com/kaspanet/kaspad/util/daghash"
)

// ResolveFinalityConflict resolves all finality conflicts by setting an arbitrary finality block, and
// re-selecting virtual parents in such a way that given finalityBlock will be in virtual's selectedParentChain
func (dag *BlockDAG) ResolveFinalityConflict(finalityBlockHash *daghash.Hash) error {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()

	finalityBlock, ok := dag.Index.LookupNode(finalityBlockHash)
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
	})

	return nil
}

// prepareForFinalityConflictResolution makes sure that the designated selectedTip once a finality conflict is resolved
// is not UTXOPendingVerification.
func (dag *BlockDAG) prepareForFinalityConflictResolution(finalityBlock *blocknode.Node) error {
	queue := blocknode.NewDownHeap()
	queue.PushSet(dag.tips)

	disqualifiedCandidates := blocknode.NewSet()
	for {
		if queue.Len() == 0 {
			return errors.New("No valid selectedTip candidates")
		}
		candidate := queue.Pop()

		isFinalityBlockInSelectedParentChain, err := dag.isInSelectedParentChainOf(finalityBlock, candidate)
		if err != nil {
			return err
		}
		if !isFinalityBlockInSelectedParentChain {
			continue
		}
		if dag.Index.BlockNodeStatus(candidate) == blocknode.StatusUTXOPendingVerification {
			err := dag.resolveNodeStatusInNewTransaction(candidate)
			if err != nil {
				return err
			}
		}
		if dag.Index.BlockNodeStatus(candidate) == blocknode.StatusValid {
			return nil
		}

		disqualifiedCandidates.Add(candidate)

		for parent := range candidate.Parents {
			if parent.Children.AreAllIn(disqualifiedCandidates) {
				queue.Push(parent)
			}
		}
	}
}
