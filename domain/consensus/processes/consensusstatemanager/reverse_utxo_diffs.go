package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
)

func (csm *consensusStateManager) ReverseUTXODiffs(tipHash *externalapi.DomainHash,
	reversalData *model.UTXODiffReversalData) error {

	// During the process of resolving a chain of blocks, we temporarily set all blocks' (except the tip)
	// UTXODiffChild to be the selected parent.
	// Once the process is complete, we can reverse said chain, to now go directly to virtual through the relevant tip
	onEnd := logger.LogAndMeasureExecutionTime(log, "reverseUTXODiffs")
	defer onEnd()

	readStagingArea := model.NewStagingArea()

	log.Debugf("Reversing utxoDiffs")

	// Set previousUTXODiff and previousBlock to tip.SelectedParent before we start touching them,
	// since previousBlock's UTXODiff is going to be over-written in the next step
	previousBlock := reversalData.SelectedParentHash
	previousUTXODiff, err := csm.utxoDiffStore.UTXODiff(csm.databaseContext, readStagingArea, previousBlock)
	if err != nil {
		return err
	}

	// tip.selectedParent is special in the sense that we don't have it's diff available in reverse, however,
	// we were able to calculate it when the tip's and tip.selectedParent's UTXOSets were known during resolveBlockStatus.
	// Therefore - we treat it separately
	err = csm.commitUTXODiffInSeparateStagingArea(previousBlock, reversalData.SelectedParentUTXODiff, tipHash)
	if err != nil {
		return err
	}

	log.Trace("Reversed 1 utxoDiff")

	previousBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, readStagingArea, previousBlock)
	if err != nil {
		return err
	}
	// Now go over the rest of the blocks and assign for every block Bi.UTXODiff = Bi+1.UTXODiff.Reversed()
	for i := 1; ; i++ {
		currentBlock := previousBlockGHOSTDAGData.SelectedParent()

		currentBlockUTXODiffChild, err := csm.utxoDiffStore.UTXODiffChild(csm.databaseContext, readStagingArea, currentBlock)
		if err != nil {
			return err
		}
		currentBlockGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, readStagingArea, currentBlock)
		if err != nil {
			return err
		}

		// We stop reversing when current's UTXODiffChild is not current's SelectedParent
		if !currentBlockGHOSTDAGData.SelectedParent().Equal(currentBlockUTXODiffChild) {
			log.Debugf("Block %s's UTXODiffChild is not it's selected parent - finish reversing", currentBlock)
			break
		}

		currentUTXODiff := previousUTXODiff.Reversed()

		// retrieve current utxoDiff for Bi, to be used by next block
		previousUTXODiff, err = csm.utxoDiffStore.UTXODiff(csm.databaseContext, readStagingArea, currentBlock)
		if err != nil {
			return err
		}

		err = csm.commitUTXODiffInSeparateStagingArea(currentBlock, currentUTXODiff, previousBlock)
		if err != nil {
			return err
		}

		previousBlock = currentBlock
		previousBlockGHOSTDAGData = currentBlockGHOSTDAGData

		log.Tracef("Reversed %d utxoDiffs", i)
	}

	return nil
}

func (csm *consensusStateManager) commitUTXODiffInSeparateStagingArea(
	blockHash *externalapi.DomainHash, utxoDiff externalapi.UTXODiff, utxoDiffChild *externalapi.DomainHash) error {

	stagingAreaForCurrentBlock := model.NewStagingArea()

	csm.utxoDiffStore.Stage(stagingAreaForCurrentBlock, blockHash, utxoDiff, utxoDiffChild)

	return staging.CommitAllChanges(csm.databaseContext, stagingAreaForCurrentBlock)
}
