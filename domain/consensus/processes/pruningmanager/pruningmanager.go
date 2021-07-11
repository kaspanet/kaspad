package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	databaseContext model.DBManager

	dagTraversalManager                 model.DAGTraversalManager
	dagTopologyManager                  model.DAGTopologyManager
	consensusStateManager               model.ConsensusStateManager
	consensusStateStore                 model.ConsensusStateStore
	ghostdagDataStore                   model.GHOSTDAGDataStore
	blocksWithMetaDataGHOSTDAGDataStore model.GHOSTDAGDataStore
	pruningStore                        model.PruningStore
	blockStatusStore                    model.BlockStatusStore
	headerSelectedTipStore              model.HeaderSelectedTipStore
	daaWindowStore                      model.BlocksWithMetaDataDAAWindowStore

	multiSetStore         model.MultisetStore
	acceptanceDataStore   model.AcceptanceDataStore
	blocksStore           model.BlockStore
	blockHeaderStore      model.BlockHeaderStore
	utxoDiffStore         model.UTXODiffStore
	daaBlocksStore        model.DAABlocksStore
	reachabilityDataStore model.ReachabilityDataStore

	isArchivalNode                  bool
	genesisHash                     *externalapi.DomainHash
	finalityInterval                uint64
	pruningDepth                    uint64
	shouldSanityCheckPruningUTXOSet bool
	k                               externalapi.KType
	difficultyAdjustmentWindowSize  int
}

// New instantiates a new PruningManager
func New(
	databaseContext model.DBManager,

	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	consensusStateManager model.ConsensusStateManager,

	consensusStateStore model.ConsensusStateStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockStatusStore model.BlockStatusStore,
	headerSelectedTipStore model.HeaderSelectedTipStore,
	multiSetStore model.MultisetStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blocksStore model.BlockStore,
	blockHeaderStore model.BlockHeaderStore,
	utxoDiffStore model.UTXODiffStore,
	daaBlocksStore model.DAABlocksStore,
	reachabilityDataStore model.ReachabilityDataStore,
	daaWindowStore model.BlocksWithMetaDataDAAWindowStore,
	blocksWithMetaDataGHOSTDAGDataStore model.GHOSTDAGDataStore,

	isArchivalNode bool,
	genesisHash *externalapi.DomainHash,
	finalityInterval uint64,
	pruningDepth uint64,
	shouldSanityCheckPruningUTXOSet bool,
	k externalapi.KType,
	difficultyAdjustmentWindowSize int,
) model.PruningManager {

	return &pruningManager{
		databaseContext:       databaseContext,
		dagTraversalManager:   dagTraversalManager,
		dagTopologyManager:    dagTopologyManager,
		consensusStateManager: consensusStateManager,

		consensusStateStore:                 consensusStateStore,
		ghostdagDataStore:                   ghostdagDataStore,
		pruningStore:                        pruningStore,
		blockStatusStore:                    blockStatusStore,
		multiSetStore:                       multiSetStore,
		acceptanceDataStore:                 acceptanceDataStore,
		blocksStore:                         blocksStore,
		blockHeaderStore:                    blockHeaderStore,
		utxoDiffStore:                       utxoDiffStore,
		headerSelectedTipStore:              headerSelectedTipStore,
		daaBlocksStore:                      daaBlocksStore,
		reachabilityDataStore:               reachabilityDataStore,
		daaWindowStore:                      daaWindowStore,
		blocksWithMetaDataGHOSTDAGDataStore: blocksWithMetaDataGHOSTDAGDataStore,

		isArchivalNode:                  isArchivalNode,
		genesisHash:                     genesisHash,
		pruningDepth:                    pruningDepth,
		finalityInterval:                finalityInterval,
		shouldSanityCheckPruningUTXOSet: shouldSanityCheckPruningUTXOSet,
		k:                               k,
		difficultyAdjustmentWindowSize:  difficultyAdjustmentWindowSize,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash
func (pm *pruningManager) UpdatePruningPointByVirtual(stagingArea *model.StagingArea) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.UpdatePruningPointByVirtual")
	defer onEnd()
	hasPruningPoint, err := pm.pruningStore.HasPruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	if !hasPruningPoint {
		hasGenesis, err := pm.blocksStore.HasBlock(pm.databaseContext, stagingArea, pm.genesisHash)
		if err != nil {
			return err
		}

		if hasGenesis {
			err = pm.savePruningPoint(stagingArea, pm.genesisHash)
			if err != nil {
				return err
			}
		}

		// Pruning point should initially set manually on a pruned-headers node.
		return nil
	}

	currentCandidate, err := pm.pruningPointCandidate(stagingArea)
	if err != nil {
		return err
	}

	currentCandidateGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentCandidate)
	if err != nil {
		return err
	}

	virtual, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	virtualSelectedParent, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, virtual.SelectedParent())
	if err != nil {
		return err
	}

	currentPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	currentPruningPointGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningPoint)
	if err != nil {
		return err
	}

	iterator, err := pm.dagTraversalManager.SelectedChildIterator(stagingArea, virtual.SelectedParent(), currentCandidate)
	if err != nil {
		return err
	}
	defer iterator.Close()

	// Finding the next pruning point candidate: look for the latest
	// selected child of the current candidate that is in depth of at
	// least pm.pruningDepth blocks from the virtual selected parent.
	//
	// Note: Sometimes the current candidate is less than pm.pruningDepth
	// from the virtual. This can happen only if the virtual blue score
	// got smaller, because virtual blue score is not guaranteed to always
	// increase (because sometimes a block with higher blue work can have
	// lower blue score).
	// In such cases we still keep the same candidate because it's guaranteed
	// that a block that was once in depth of pm.pruningDepth cannot be
	// reorged without causing a finality conflict first.
	newCandidate := currentCandidate
	newCandidateGHOSTDAGData := currentCandidateGHOSTDAGData

	newPruningPoint := currentPruningPoint
	newPruningPointGHOSTDAGData := currentPruningPointGHOSTDAGData
	for ok := iterator.First(); ok; ok = iterator.Next() {
		selectedChild, err := iterator.Get()
		if err != nil {
			return err
		}
		selectedChildGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, selectedChild)
		if err != nil {
			return err
		}

		if virtualSelectedParent.BlueScore()-selectedChildGHOSTDAGData.BlueScore() < pm.pruningDepth {
			break
		}

		newCandidate = selectedChild
		newCandidateGHOSTDAGData = selectedChildGHOSTDAGData

		// We move the pruning point every time the candidate's finality score is
		// bigger than the current pruning point finality score.
		if pm.finalityScore(newCandidateGHOSTDAGData.BlueScore()) > pm.finalityScore(newPruningPointGHOSTDAGData.BlueScore()) {
			newPruningPoint = newCandidate
			newPruningPointGHOSTDAGData = newCandidateGHOSTDAGData
		}
	}

	if !newCandidate.Equal(currentCandidate) {
		log.Debugf("Staged a new pruning candidate, old: %s, new: %s", currentCandidate, newCandidate)
		pm.pruningStore.StagePruningPointCandidate(stagingArea, newCandidate)
	}

	// We move the pruning point every time the candidate's finality score is
	// bigger than the current pruning point finality score.
	if pm.finalityScore(newCandidateGHOSTDAGData.BlueScore()) <= pm.finalityScore(currentPruningPointGHOSTDAGData.BlueScore()) {
		return nil
	}

	if !newPruningPoint.Equal(currentPruningPoint) {
		log.Debugf("Moving pruning point from %s to %s", currentPruningPoint, newPruningPoint)
		err = pm.savePruningPoint(stagingArea, newPruningPoint)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *pruningManager) isInPruningFutureOrInVirtualPast(stagingArea *model.StagingArea, block *externalapi.DomainHash,
	pruningPoint *externalapi.DomainHash, virtualParents []*externalapi.DomainHash) (bool, error) {

	hasPruningPointInPast, err := pm.dagTopologyManager.IsAncestorOf(stagingArea, pruningPoint, block)
	if err != nil {
		return false, err
	}
	if hasPruningPointInPast {
		return true, nil
	}
	// Because virtual doesn't have reachability data, we need to check reachability
	// using it parents.
	isInVirtualPast, err := pm.dagTopologyManager.IsAncestorOfAny(stagingArea, block, virtualParents)
	if err != nil {
		return false, err
	}
	if isInVirtualPast {
		return true, nil
	}

	return false, nil
}

func (pm *pruningManager) deletePastBlocks(stagingArea *model.StagingArea, pruningPoint *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.deletePastBlocks")
	defer onEnd()

	// Go over all pruningPoint.Past and pruningPoint.Anticone that's not in virtual.Past
	queue := pm.dagTraversalManager.NewDownHeap(stagingArea)
	virtualParents, err := pm.dagTopologyManager.Parents(stagingArea, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	// Start queue with all tips that are below the pruning point (and on the way remove them from list of tips)
	prunedTips, err := pm.pruneTips(stagingArea, pruningPoint, virtualParents)
	if err != nil {
		return err
	}
	err = queue.PushSlice(prunedTips)
	if err != nil {
		return err
	}

	// Add pruningPoint.Parents to queue
	parents, err := pm.dagTopologyManager.Parents(stagingArea, pruningPoint)
	if err != nil {
		return err
	}

	isChildOfVirtualGenesis := len(parents) == 1 && parents[0].Equal(model.VirtualGenesisBlockHash)
	if !isChildOfVirtualGenesis {
		err = queue.PushSlice(parents)
		if err != nil {
			return err
		}
	}

	err = pm.deleteBlocksDownward(stagingArea, queue)
	if err != nil {
		return err
	}

	return nil
}

func (pm *pruningManager) deleteBlocksDownward(stagingArea *model.StagingArea, queue model.BlockHeap) error {
	visited := map[externalapi.DomainHash]struct{}{}
	// Prune everything in the queue including its past
	for queue.Len() > 0 {
		current := queue.Pop()
		if _, ok := visited[*current]; ok {
			continue
		}
		visited[*current] = struct{}{}

		alreadyPruned, err := pm.deleteBlock(stagingArea, current)
		if err != nil {
			return err
		}
		if !alreadyPruned {
			parents, err := pm.dagTopologyManager.Parents(stagingArea, current)
			if err != nil {
				return err
			}

			isChildOfVirtualGenesis := len(parents) == 1 && parents[0].Equal(model.VirtualGenesisBlockHash)
			if !isChildOfVirtualGenesis {
				err = queue.PushSlice(parents)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (pm *pruningManager) pruneTips(stagingArea *model.StagingArea, pruningPoint *externalapi.DomainHash,
	virtualParents []*externalapi.DomainHash) (prunedTips []*externalapi.DomainHash, err error) {

	// Find P.AC that's not in V.Past
	dagTips, err := pm.consensusStateStore.Tips(stagingArea, pm.databaseContext)
	if err != nil {
		return nil, err
	}
	newTips := make([]*externalapi.DomainHash, 0, len(dagTips))
	for _, tip := range dagTips {
		isInPruningFutureOrInVirtualPast, err :=
			pm.isInPruningFutureOrInVirtualPast(stagingArea, tip, pruningPoint, virtualParents)
		if err != nil {
			return nil, err
		}
		if !isInPruningFutureOrInVirtualPast {
			prunedTips = append(prunedTips, tip)
		} else {
			newTips = append(newTips, tip)
		}
	}
	pm.consensusStateStore.StageTips(stagingArea, newTips)

	return prunedTips, nil
}

func (pm *pruningManager) savePruningPoint(stagingArea *model.StagingArea, pruningPointHash *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.savePruningPoint")
	defer onEnd()

	// If pruningPointHash is the genesis then there's no pruning point set right now.
	if !pruningPointHash.Equal(pm.genesisHash) {
		previousPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
		if err != nil {
			return err
		}
		pm.pruningStore.StagePreviousPruningPoint(stagingArea, previousPruningPoint)
	}
	pm.pruningStore.StagePruningPoint(stagingArea, pruningPointHash)
	pm.pruningStore.StageStartUpdatingPruningPointUTXOSet(stagingArea)

	return nil
}

func (pm *pruningManager) deleteBlock(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (
	alreadyPruned bool, err error) {

	status, err := pm.blockStatusStore.Get(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}
	if status == externalapi.StatusHeaderOnly {
		return true, nil
	}

	pm.blockStatusStore.Stage(stagingArea, blockHash, externalapi.StatusHeaderOnly)
	if pm.isArchivalNode {
		return false, nil
	}

	pm.multiSetStore.Delete(stagingArea, blockHash)
	pm.acceptanceDataStore.Delete(stagingArea, blockHash)
	pm.blocksStore.Delete(stagingArea, blockHash)
	pm.utxoDiffStore.Delete(stagingArea, blockHash)
	pm.daaBlocksStore.Delete(stagingArea, blockHash)

	return false, nil
}

func (pm *pruningManager) IsValidPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	if *pm.genesisHash == *blockHash {
		return true, nil
	}

	headersSelectedTip, err := pm.headerSelectedTipStore.HeadersSelectedTip(pm.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}

	// A pruning point has to be in the selected chain of the headers selected tip.
	headersSelectedTipGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, headersSelectedTip)
	if err != nil {
		return false, err
	}

	isInSelectedParentChainOfHeadersSelectedTip, err :=
		pm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, blockHash, headersSelectedTip)
	if err != nil {
		return false, err
	}

	if !isInSelectedParentChainOfHeadersSelectedTip {
		return false, nil
	}

	ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return false, err
	}

	// A pruning point has to be at depth of at least pm.pruningDepth
	if headersSelectedTipGHOSTDAGData.BlueScore()-ghostdagData.BlueScore() < pm.pruningDepth {
		return false, nil
	}

	return true, nil
}

func (pm *pruningManager) pruningPointCandidate(stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	hasPruningPointCandidate, err := pm.pruningStore.HasPruningPointCandidate(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	if !hasPruningPointCandidate {
		return pm.genesisHash, nil
	}

	return pm.pruningStore.PruningPointCandidate(pm.databaseContext, stagingArea)
}

// validateUTXOSetFitsCommitment makes sure that the calculated UTXOSet of the new pruning point fits the commitment.
// This is a sanity test, to make sure that kaspad doesn't store, and subsequently sends syncing peers the wrong UTXOSet.
func (pm *pruningManager) validateUTXOSetFitsCommitment(stagingArea *model.StagingArea, pruningPointHash *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.validateUTXOSetFitsCommitment")
	defer onEnd()

	utxoSetIterator, err := pm.pruningStore.PruningPointUTXOIterator(pm.databaseContext)
	if err != nil {
		return err
	}
	defer utxoSetIterator.Close()

	utxoSetMultiset := multiset.New()
	for ok := utxoSetIterator.First(); ok; ok = utxoSetIterator.Next() {
		outpoint, entry, err := utxoSetIterator.Get()
		if err != nil {
			return err
		}
		serializedUTXO, err := utxo.SerializeUTXO(entry, outpoint)
		if err != nil {
			return err
		}
		utxoSetMultiset.Add(serializedUTXO)
	}
	utxoSetHash := utxoSetMultiset.Hash()

	header, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, pruningPointHash)
	if err != nil {
		return err
	}
	expectedUTXOCommitment := header.UTXOCommitment()

	if !expectedUTXOCommitment.Equal(utxoSetHash) {
		return errors.Errorf("Calculated UTXOSet for next pruning point %s doesn't match it's UTXO commitment\n"+
			"Calculated UTXOSet hash: %s. Commitment: %s",
			pruningPointHash, utxoSetHash, expectedUTXOCommitment)
	}

	log.Debugf("Validated the pruning point %s UTXO commitment: %s", pruningPointHash, utxoSetHash)

	return nil
}

// This function takes 2 points (currentPruningHash, previousPruningHash) and traverses the UTXO diff children DAG
// until it finds a common descendant, at the worse case this descendant will be the current SelectedTip.
// it then creates 2 diffs, one from that descendant to previousPruningHash and another from that descendant to currentPruningHash
// then using `DiffFrom` it converts these 2 diffs to a single diff from previousPruningHash to currentPruningHash.
// this way should be the fastest way to get the difference between the 2 points, and should perform much better than restoring the full UTXO set.
func (pm *pruningManager) calculateDiffBetweenPreviousAndCurrentPruningPoints(stagingArea *model.StagingArea, currentPruningHash *externalapi.DomainHash) (externalapi.UTXODiff, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.calculateDiffBetweenPreviousAndCurrentPruningPoints")
	defer onEnd()
	if currentPruningHash.Equal(pm.genesisHash) {
		return utxo.NewUTXODiff(), nil
	}

	previousPruningHash, err := pm.pruningStore.PreviousPruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}
	currentPruningGhostDAG, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningHash)
	if err != nil {
		return nil, err
	}
	previousPruningGhostDAG, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, previousPruningHash)
	if err != nil {
		return nil, err
	}

	currentPruningCurrentDiffChild := currentPruningHash
	previousPruningCurrentDiffChild := previousPruningHash
	// We need to use BlueWork because it's the only thing that's monotonic in the whole DAG
	// We use the BlueWork to know which point is currently lower on the DAG so we can keep climbing its children,
	// that way we keep climbing on the lowest point until they both reach the exact same descendant
	currentPruningCurrentDiffChildBlueWork := currentPruningGhostDAG.BlueWork()
	previousPruningCurrentDiffChildBlueWork := previousPruningGhostDAG.BlueWork()

	var diffsFromPrevious []externalapi.UTXODiff
	var diffsFromCurrent []externalapi.UTXODiff
	for {
		// if currentPruningCurrentDiffChildBlueWork > previousPruningCurrentDiffChildBlueWork
		if currentPruningCurrentDiffChildBlueWork.Cmp(previousPruningCurrentDiffChildBlueWork) == 1 {
			utxoDiff, err := pm.utxoDiffStore.UTXODiff(pm.databaseContext, stagingArea, previousPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffsFromPrevious = append(diffsFromPrevious, utxoDiff)
			previousPruningCurrentDiffChild, err = pm.utxoDiffStore.UTXODiffChild(pm.databaseContext, stagingArea, previousPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffChildGhostDag, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, previousPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			previousPruningCurrentDiffChildBlueWork = diffChildGhostDag.BlueWork()
		} else if currentPruningCurrentDiffChild.Equal(previousPruningCurrentDiffChild) {
			break
		} else {
			utxoDiff, err := pm.utxoDiffStore.UTXODiff(pm.databaseContext, stagingArea, currentPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffsFromCurrent = append(diffsFromCurrent, utxoDiff)
			currentPruningCurrentDiffChild, err = pm.utxoDiffStore.UTXODiffChild(pm.databaseContext, stagingArea, currentPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffChildGhostDag, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			currentPruningCurrentDiffChildBlueWork = diffChildGhostDag.BlueWork()
		}
	}
	// The order in which we apply the diffs should be from top to bottom, but we traversed from bottom to top
	// so we apply the diffs in reverse order.
	oldDiff := utxo.NewMutableUTXODiff()
	for i := len(diffsFromPrevious) - 1; i >= 0; i-- {
		err = oldDiff.WithDiffInPlace(diffsFromPrevious[i])
		if err != nil {
			return nil, err
		}
	}
	newDiff := utxo.NewMutableUTXODiff()
	for i := len(diffsFromCurrent) - 1; i >= 0; i-- {
		err = newDiff.WithDiffInPlace(diffsFromCurrent[i])
		if err != nil {
			return nil, err
		}
	}
	return oldDiff.DiffFrom(newDiff.ToImmutable())
}

// finalityScore is the number of finality intervals passed since
// the given block.
func (pm *pruningManager) finalityScore(blueScore uint64) uint64 {
	return blueScore / pm.finalityInterval
}

func (pm *pruningManager) ClearImportedPruningPointData() error {
	err := pm.pruningStore.ClearImportedPruningPointMultiset(pm.databaseContext)
	if err != nil {
		return err
	}
	return pm.pruningStore.ClearImportedPruningPointUTXOs(pm.databaseContext)
}

func (pm *pruningManager) AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error {
	dbTx, err := pm.databaseContext.Begin()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessClosed()

	importedMultiset, err := pm.pruningStore.ImportedPruningPointMultiset(dbTx)
	if err != nil {
		if !database.IsNotFoundError(err) {
			return err
		}
		importedMultiset = multiset.New()
	}
	for _, outpointAndUTXOEntryPair := range outpointAndUTXOEntryPairs {
		serializedUTXO, err := utxo.SerializeUTXO(outpointAndUTXOEntryPair.UTXOEntry, outpointAndUTXOEntryPair.Outpoint)
		if err != nil {
			return err
		}
		importedMultiset.Add(serializedUTXO)
	}
	err = pm.pruningStore.UpdateImportedPruningPointMultiset(dbTx, importedMultiset)
	if err != nil {
		return err
	}

	err = pm.pruningStore.AppendImportedPruningPointUTXOs(dbTx, outpointAndUTXOEntryPairs)
	if err != nil {
		return err
	}

	return dbTx.Commit()
}

func (pm *pruningManager) UpdatePruningPointIfRequired() error {
	hadStartedUpdatingPruningPointUTXOSet, err := pm.pruningStore.HadStartedUpdatingPruningPointUTXOSet(pm.databaseContext)
	if err != nil {
		return err
	}
	if !hadStartedUpdatingPruningPointUTXOSet {
		return nil
	}

	log.Debugf("Pruning point UTXO set update is required")
	err = pm.updatePruningPoint()
	if err != nil {
		return err
	}
	log.Debugf("Pruning point UTXO set updated")

	return nil
}

func (pm *pruningManager) updatePruningPoint() error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "updatePruningPoint")
	defer onEnd()

	logger.LogMemoryStats(log, "updatePruningPoint start")
	defer logger.LogMemoryStats(log, "updatePruningPoint end")

	stagingArea := model.NewStagingArea()
	log.Debugf("Getting the pruning point")
	pruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	log.Debugf("Restoring the pruning point UTXO set")
	utxoSetDiff, err := pm.calculateDiffBetweenPreviousAndCurrentPruningPoints(stagingArea, pruningPoint)
	if err != nil {
		return err
	}
	log.Debugf("Updating the pruning point UTXO set")
	err = pm.pruningStore.UpdatePruningPointUTXOSet(pm.databaseContext, utxoSetDiff)
	if err != nil {
		return err
	}
	if pm.shouldSanityCheckPruningUTXOSet {
		err = pm.validateUTXOSetFitsCommitment(stagingArea, pruningPoint)
		if err != nil {
			return err
		}
	}
	err = pm.deletePastBlocks(stagingArea, pruningPoint)
	if err != nil {
		return err
	}

	err = staging.CommitAllChanges(pm.databaseContext, stagingArea)
	if err != nil {
		return err
	}

	log.Debugf("Finishing updating the pruning point UTXO set")
	return pm.pruningStore.FinishUpdatingPruningPointUTXOSet(pm.databaseContext)
}

func (pm *pruningManager) PruneAllBlocksBelow(stagingArea *model.StagingArea, pruningPointHash *externalapi.DomainHash) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "PruneAllBlocksBelow")
	defer onEnd()

	iterator, err := pm.blocksStore.AllBlockHashesIterator(pm.databaseContext)
	if err != nil {
		return err
	}
	defer iterator.Close()

	for ok := iterator.First(); ok; ok = iterator.Next() {
		blockHash, err := iterator.Get()
		if err != nil {
			return err
		}
		isInPastOfPruningPoint, err := pm.dagTopologyManager.IsAncestorOf(stagingArea, pruningPointHash, blockHash)
		if err != nil {
			return err
		}
		if !isInPastOfPruningPoint {
			continue
		}
		_, err = pm.deleteBlock(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *pruningManager) PruningPointAndItsAnticoneWithMetaData() ([]*externalapi.BlockWithMetaData, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "PruningPointAndItsAnticoneWithMetaData")
	defer onEnd()

	stagingArea := model.NewStagingArea()
	pruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	pruningPointAnticone, err := pm.dagTraversalManager.AnticoneFromVirtual(stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}

	blocks := make([]*externalapi.BlockWithMetaData, 0, len(pruningPointAnticone)+1)

	pruningPointWithMetaData, err := pm.blockWithMetaData(stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}

	blocks = append(blocks, pruningPointWithMetaData)
	for _, blockHash := range pruningPointAnticone {
		blockWithMetaData, err := pm.blockWithMetaData(stagingArea, blockHash)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, blockWithMetaData)
	}

	return blocks, nil
}

func (pm *pruningManager) blockWithMetaData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.BlockWithMetaData, error) {
	block, err := pm.blocksStore.Block(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	daaScore, err := pm.daaBlocksStore.DAAScore(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	windowSize := pm.difficultyAdjustmentWindowSize + 1
	window, err := pm.dagTraversalManager.BlockWindow(stagingArea, blockHash, windowSize)
	if err != nil {
		return nil, err
	}

	windowPairs := make([]*externalapi.BlockWithMetaDataDAABlock, len(window))
	for i, blockWindowHash := range window {
		header, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, blockWindowHash)
		if err != nil {
			return nil, err
		}

		ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockWindowHash)
		if database.IsNotFoundError(err) {
			daaBlock, err := pm.daaWindowStore.DAAWindowBlock(pm.databaseContext, stagingArea, blockHash, uint64(i))
			if err != nil {
				return nil, err
			}

			ghostdagData = daaBlock.GHOSTDAGData
		} else if err != nil {
			return nil, err
		}

		windowPairs[i] = &externalapi.BlockWithMetaDataDAABlock{
			Header:       header,
			GHOSTDAGData: ghostdagData,
		}
	}

	ghostdagDataHashPairs := make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, pm.k)
	current := blockHash
	for i := externalapi.KType(0); i < pm.k+1; i++ {
		ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, current)
		if database.IsNotFoundError(err) {
			ghostdagData, err = pm.blocksWithMetaDataGHOSTDAGDataStore.Get(pm.databaseContext, stagingArea, current)
			if err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		ghostdagDataHashPairs = append(ghostdagDataHashPairs, &externalapi.BlockGHOSTDAGDataHashPair{
			Hash:         current,
			GHOSTDAGData: ghostdagData,
		})

		if ghostdagData.SelectedParent().Equal(pm.genesisHash) {
			break
		}

		current = ghostdagData.SelectedParent()
	}

	return &externalapi.BlockWithMetaData{
		Block:        block,
		DAAScore:     daaScore,
		DAAWindow:    windowPairs,
		GHOSTDAGData: ghostdagDataHashPairs,
	}, nil
}
