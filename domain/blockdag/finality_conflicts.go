package blockdag

import (
	"github.com/kaspanet/kaspad/util/mstime"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/daghash"
)

// ResolveFinalityConflict resolve all finality conflicts by setting an arbitrary finality block, and
// re-selecting virtual parents in such a way that given finalityBlock will be in virtual's selectedParentChain
func (dag *BlockDAG) ResolveFinalityConflict(finalityBlockHash *daghash.Hash) error {
	dag.dagLock.Lock()
	defer dag.dagLock.RUnlock()

	finalityBlock, ok := dag.index.LookupNode(finalityBlockHash)
	if !ok {
		return errors.Errorf("Couldn't find finality block with hash %s", finalityBlockHash)
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
		ResolutionTime:    mstime.Now(),
	})

	return nil
}
