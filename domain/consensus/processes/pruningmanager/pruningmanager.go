package pruningmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/multiset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/consensus/utils/virtual"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
	"sort"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	databaseContext model.DBManager

	dagTraversalManager   model.DAGTraversalManager
	dagTopologyManager    model.DAGTopologyManager
	consensusStateManager model.ConsensusStateManager
	finalityManager       model.FinalityManager

	consensusStateStore                 model.ConsensusStateStore
	ghostdagDataStore                   model.GHOSTDAGDataStore
	pruningStore                        model.PruningStore
	blockStatusStore                    model.BlockStatusStore
	headerSelectedTipStore              model.HeaderSelectedTipStore
	blocksWithTrustedDataDAAWindowStore model.BlocksWithTrustedDataDAAWindowStore
	multiSetStore                       model.MultisetStore
	acceptanceDataStore                 model.AcceptanceDataStore
	blocksStore                         model.BlockStore
	blockHeaderStore                    model.BlockHeaderStore
	utxoDiffStore                       model.UTXODiffStore
	daaBlocksStore                      model.DAABlocksStore
	reachabilityDataStore               model.ReachabilityDataStore

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
	finalityManager model.FinalityManager,

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
	blocksWithTrustedDataDAAWindowStore model.BlocksWithTrustedDataDAAWindowStore,

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
		finalityManager:       finalityManager,

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
		blocksWithTrustedDataDAAWindowStore: blocksWithTrustedDataDAAWindowStore,

		isArchivalNode:                  isArchivalNode,
		genesisHash:                     genesisHash,
		pruningDepth:                    pruningDepth,
		finalityInterval:                finalityInterval,
		shouldSanityCheckPruningUTXOSet: shouldSanityCheckPruningUTXOSet,
		k:                               k,
		difficultyAdjustmentWindowSize:  difficultyAdjustmentWindowSize,
	}
}

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

	virtualGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return err
	}

	newPruningPoint, newCandidate, err := pm.nextPruningPointAndCandidateByBlockHash(stagingArea, virtualGHOSTDAGData.SelectedParent(), nil)
	if err != nil {
		return err
	}

	currentCandidate, err := pm.pruningPointCandidate(stagingArea)
	if err != nil {
		return err
	}

	if !newCandidate.Equal(currentCandidate) {
		log.Debugf("Staged a new pruning candidate, old: %s, new: %s", currentCandidate, newCandidate)
		pm.pruningStore.StagePruningPointCandidate(stagingArea, newCandidate)
	}

	currentPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return err
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

func (pm *pruningManager) nextPruningPointAndCandidateByBlockHash(stagingArea *model.StagingArea,
	blockHash, suggestedLowHash *externalapi.DomainHash) (*externalapi.DomainHash, *externalapi.DomainHash, error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "pruningManager.nextPruningPointAndCandidateByBlockHash")
	defer onEnd()

	currentCandidate, err := pm.pruningPointCandidate(stagingArea)
	if err != nil {
		return nil, nil, err
	}

	lowHash := currentCandidate
	if suggestedLowHash != nil {
		isSuggestedLowHashInSelectedParentChainOfCurrentCandidate, err := pm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, suggestedLowHash, currentCandidate)
		if err != nil {
			return nil, nil, err
		}

		if !isSuggestedLowHashInSelectedParentChainOfCurrentCandidate {
			isCurrentCandidateInSelectedParentChainOfSuggestedLowHash, err := pm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, currentCandidate, suggestedLowHash)
			if err != nil {
				return nil, nil, err
			}

			if !isCurrentCandidateInSelectedParentChainOfSuggestedLowHash {
				panic(errors.Errorf("suggested low hash %s is not on the same selected chain as the pruning candidate %s", suggestedLowHash, currentCandidate))
			}
			lowHash = suggestedLowHash
		}
	}

	ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, nil, err
	}

	currentPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, nil, err
	}

	currentPruningPointGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningPoint, false)
	if err != nil {
		return nil, nil, err
	}

	// We iterate until the selected parent of the given block, in order to allow a situation where the given block hash
	// belongs to the virtual. This shouldn't change anything since the max blue score difference between a block and its
	// selected parent is K, and K << pm.pruningDepth.
	iterator, err := pm.dagTraversalManager.SelectedChildIterator(stagingArea, ghostdagData.SelectedParent(), lowHash, true)
	if err != nil {
		return nil, nil, err
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

	newPruningPoint := currentPruningPoint
	newPruningPointGHOSTDAGData := currentPruningPointGHOSTDAGData
	for ok := iterator.First(); ok; ok = iterator.Next() {
		selectedChild, err := iterator.Get()
		if err != nil {
			return nil, nil, err
		}
		selectedChildGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, selectedChild, false)
		if err != nil {
			return nil, nil, err
		}

		if ghostdagData.BlueScore()-selectedChildGHOSTDAGData.BlueScore() < pm.pruningDepth {
			break
		}

		newCandidate = selectedChild
		newCandidateGHOSTDAGData := selectedChildGHOSTDAGData

		// We move the pruning point every time the candidate's finality score is
		// bigger than the current pruning point finality score.
		if pm.finalityScore(newCandidateGHOSTDAGData.BlueScore()) > pm.finalityScore(newPruningPointGHOSTDAGData.BlueScore()) {
			newPruningPoint = newCandidate
			newPruningPointGHOSTDAGData = newCandidateGHOSTDAGData
		}
	}

	return newPruningPoint, newCandidate, nil
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

	if !virtual.ContainsOnlyVirtualGenesis(parents) {
		err = queue.PushSlice(parents)
		if err != nil {
			return err
		}
	}

	blocksToKeep, err := pm.calculateBlocksToKeep(stagingArea, pruningPoint)
	if err != nil {
		return err
	}
	err = pm.deleteBlocksDownward(stagingArea, queue, blocksToKeep)
	if err != nil {
		return err
	}

	return nil
}

func (pm *pruningManager) calculateBlocksToKeep(stagingArea *model.StagingArea,
	pruningPoint *externalapi.DomainHash) (map[externalapi.DomainHash]struct{}, error) {

	pruningPointAnticone, err := pm.dagTraversalManager.AnticoneFromVirtualPOV(stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}
	pruningPointAndItsAnticone := append(pruningPointAnticone, pruningPoint)
	blocksToKeep := make(map[externalapi.DomainHash]struct{})
	for _, blockHash := range pruningPointAndItsAnticone {
		blocksToKeep[*blockHash] = struct{}{}
		blockWindow, err := pm.dagTraversalManager.BlockWindow(stagingArea, blockHash, pm.difficultyAdjustmentWindowSize)
		if err != nil {
			return nil, err
		}
		for _, windowBlockHash := range blockWindow {
			blocksToKeep[*windowBlockHash] = struct{}{}
		}
	}
	return blocksToKeep, nil
}

func (pm *pruningManager) deleteBlocksDownward(stagingArea *model.StagingArea,
	queue model.BlockHeap, blocksToKeep map[externalapi.DomainHash]struct{}) error {

	visited := map[externalapi.DomainHash]struct{}{}
	// Prune everything in the queue including its past, unless it's in `blocksToKeep`
	for queue.Len() > 0 {
		current := queue.Pop()
		if _, ok := visited[*current]; ok {
			continue
		}
		visited[*current] = struct{}{}

		shouldAddParents := true
		if _, ok := blocksToKeep[*current]; !ok {
			alreadyPruned, err := pm.deleteBlock(stagingArea, current)
			if err != nil {
				return err
			}
			shouldAddParents = !alreadyPruned
		}

		if shouldAddParents {
			parents, err := pm.dagTopologyManager.Parents(stagingArea, current)
			if err != nil {
				return err
			}

			if !virtual.ContainsOnlyVirtualGenesis(parents) {
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
	err := pm.pruningStore.StagePruningPoint(pm.databaseContext, stagingArea, pruningPointHash)
	if err != nil {
		return err
	}
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
	headersSelectedTipGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, headersSelectedTip, false)
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

	ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return false, err
	}

	// A pruning point has to be at depth of at least pm.pruningDepth
	if headersSelectedTipGHOSTDAGData.BlueScore()-ghostdagData.BlueScore() < pm.pruningDepth {
		return false, nil
	}

	return true, nil
}

func (pm *pruningManager) ArePruningPointsViolatingFinality(stagingArea *model.StagingArea,
	pruningPoints []externalapi.BlockHeader) (bool, error) {

	virtualFinalityPoint, err := pm.finalityManager.VirtualFinalityPoint(stagingArea)
	if err != nil {
		return false, err
	}

	virtualFinalityPointFinalityPoint, err := pm.finalityManager.FinalityPoint(stagingArea, virtualFinalityPoint, false)
	if err != nil {
		return false, err
	}

	for i := len(pruningPoints) - 1; i >= 0; i-- {
		blockHash := consensushashing.HeaderHash(pruningPoints[i])
		exists, err := pm.blockStatusStore.Exists(pm.databaseContext, stagingArea, blockHash)
		if err != nil {
			return false, err
		}

		if !exists {
			continue
		}

		isInSelectedParentChainOfVirtualFinalityPointFinalityPoint, err := pm.dagTopologyManager.
			IsInSelectedParentChainOf(stagingArea, virtualFinalityPointFinalityPoint, blockHash)
		if err != nil {
			return false, err
		}

		return !isInSelectedParentChainOfVirtualFinalityPointFinalityPoint, nil
	}

	// If no pruning point is known, there's definitely a finality violation
	return true, nil
}

func (pm *pruningManager) ArePruningPointsInValidChain(stagingArea *model.StagingArea) (bool, error) {
	lastPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}

	expectedPruningPoints := make([]*externalapi.DomainHash, 0)
	headersSelectedTip, err := pm.headerSelectedTipStore.HeadersSelectedTip(pm.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}

	current := headersSelectedTip
	for !current.Equal(lastPruningPoint) {
		header, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, current)
		if err != nil {
			return false, err
		}

		if len(expectedPruningPoints) == 0 ||
			!expectedPruningPoints[len(expectedPruningPoints)-1].Equal(header.PruningPoint()) {

			expectedPruningPoints = append(expectedPruningPoints, header.PruningPoint())
		}

		currentGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, current, false)
		if err != nil {
			return false, err
		}

		current = currentGHOSTDAGData.SelectedParent()
	}

	lastPruningPointIndex, err := pm.pruningStore.CurrentPruningPointIndex(pm.databaseContext, stagingArea)
	if err != nil {
		return false, err
	}

	for i := lastPruningPointIndex; ; i-- {
		pruningPoint, err := pm.pruningStore.PruningPointByIndex(pm.databaseContext, stagingArea, i)
		if err != nil {
			return false, err
		}

		header, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, pruningPoint)
		if err != nil {
			return false, err
		}

		var expectedPruningPoint *externalapi.DomainHash
		expectedPruningPoint, expectedPruningPoints = expectedPruningPoints[0], expectedPruningPoints[1:]
		if !pruningPoint.Equal(expectedPruningPoint) {
			return false, nil
		}

		if i == 0 {
			if len(expectedPruningPoints) != 0 {
				return false, nil
			}
			if !pruningPoint.Equal(pm.genesisHash) {
				return false, nil
			}
			break
		}

		if !expectedPruningPoints[len(expectedPruningPoints)-1].Equal(header.PruningPoint()) {
			expectedPruningPoints = append(expectedPruningPoints, header.PruningPoint())
		}
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
		iter, err := pm.consensusStateManager.RestorePastUTXOSetIterator(stagingArea, currentPruningHash)
		if err != nil {
			return nil, err
		}
		set := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry)
		for ok := iter.First(); ok; ok = iter.Next() {
			outpoint, entry, err := iter.Get()
			if err != nil {
				return nil, err
			}
			set[*outpoint] = entry
		}
		return utxo.NewUTXODiffFromCollections(utxo.NewUTXOCollection(set), utxo.NewUTXOCollection(make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry)))
	}

	pruningPointIndex, err := pm.pruningStore.CurrentPruningPointIndex(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	if pruningPointIndex == 0 {
		return nil, errors.Errorf("previous pruning point doesn't exist")
	}

	previousPruningHash, err := pm.pruningStore.PruningPointByIndex(pm.databaseContext, stagingArea, pruningPointIndex-1)
	if err != nil {
		return nil, err
	}
	currentPruningGhostDAG, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningHash, false)
	if err != nil {
		return nil, err
	}
	previousPruningGhostDAG, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, previousPruningHash, false)
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

	var diffHashesFromPrevious []*externalapi.DomainHash
	var diffHashesFromCurrent []*externalapi.DomainHash
	for {
		// if currentPruningCurrentDiffChildBlueWork > previousPruningCurrentDiffChildBlueWork
		if currentPruningCurrentDiffChildBlueWork.Cmp(previousPruningCurrentDiffChildBlueWork) == 1 {
			diffHashesFromPrevious = append(diffHashesFromPrevious, previousPruningCurrentDiffChild)
			previousPruningCurrentDiffChild, err = pm.utxoDiffStore.UTXODiffChild(pm.databaseContext, stagingArea, previousPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffChildGhostDag, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, previousPruningCurrentDiffChild, false)
			if err != nil {
				return nil, err
			}
			previousPruningCurrentDiffChildBlueWork = diffChildGhostDag.BlueWork()
		} else if currentPruningCurrentDiffChild.Equal(previousPruningCurrentDiffChild) {
			break
		} else {
			diffHashesFromCurrent = append(diffHashesFromCurrent, currentPruningCurrentDiffChild)
			currentPruningCurrentDiffChild, err = pm.utxoDiffStore.UTXODiffChild(pm.databaseContext, stagingArea, currentPruningCurrentDiffChild)
			if err != nil {
				return nil, err
			}
			diffChildGhostDag, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, currentPruningCurrentDiffChild, false)
			if err != nil {
				return nil, err
			}
			currentPruningCurrentDiffChildBlueWork = diffChildGhostDag.BlueWork()
		}
	}
	// The order in which we apply the diffs should be from top to bottom, but we traversed from bottom to top
	// so we apply the diffs in reverse order.
	oldDiff := utxo.NewMutableUTXODiff()
	for i := len(diffHashesFromPrevious) - 1; i >= 0; i-- {
		utxoDiff, err := pm.utxoDiffStore.UTXODiff(pm.databaseContext, stagingArea, diffHashesFromPrevious[i])
		if err != nil {
			return nil, err
		}
		err = oldDiff.WithDiffInPlace(utxoDiff)
		if err != nil {
			return nil, err
		}
	}
	newDiff := utxo.NewMutableUTXODiff()
	for i := len(diffHashesFromCurrent) - 1; i >= 0; i-- {
		utxoDiff, err := pm.utxoDiffStore.UTXODiff(pm.databaseContext, stagingArea, diffHashesFromCurrent[i])
		if err != nil {
			return nil, err
		}
		err = newDiff.WithDiffInPlace(utxoDiff)
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

func (pm *pruningManager) PruningPointAndItsAnticone() ([]*externalapi.DomainHash, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "PruningPointAndItsAnticone")
	defer onEnd()

	stagingArea := model.NewStagingArea()
	pruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	pruningPointAnticone, err := pm.dagTraversalManager.AnticoneFromVirtualPOV(stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}

	// Sorting the blocks in topological order
	var sortErr error
	sort.Slice(pruningPointAnticone, func(i, j int) bool {
		headerI, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, pruningPointAnticone[i])
		if err != nil {
			sortErr = err
			return false
		}

		headerJ, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, pruningPointAnticone[j])
		if err != nil {
			sortErr = err
			return false
		}

		return headerI.BlueWork().Cmp(headerJ.BlueWork()) < 0
	})
	if sortErr != nil {
		return nil, sortErr
	}

	// The pruning point should always come first
	return append([]*externalapi.DomainHash{pruningPoint}, pruningPointAnticone...), nil
}

func (pm *pruningManager) BlockWithTrustedData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.BlockWithTrustedData, error) {
	block, err := pm.blocksStore.Block(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	daaScore, err := pm.daaBlocksStore.DAAScore(pm.databaseContext, stagingArea, blockHash)
	if err != nil {
		return nil, err
	}

	windowSize := pm.difficultyAdjustmentWindowSize
	window, err := pm.dagTraversalManager.BlockWindowWithGHOSTDAGData(stagingArea, blockHash, windowSize)
	if err != nil {
		return nil, err
	}

	windowPairs := make([]*externalapi.TrustedDataDataDAABlock, len(window))
	for i, daaBlock := range window {
		daaDomainBlock, err := pm.blocksStore.Block(pm.databaseContext, stagingArea, daaBlock.Hash)
		if err != nil {
			return nil, err
		}

		windowPairs[i] = &externalapi.TrustedDataDataDAABlock{
			Block:        daaDomainBlock,
			GHOSTDAGData: daaBlock.GHOSTDAGData,
		}
	}

	ghostdagDataHashPairs := make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, pm.k)
	current := blockHash
	isTrustedData := false
	for i := externalapi.KType(0); i < pm.k+1; i++ {
		ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, current, isTrustedData)
		isNotFoundError := database.IsNotFoundError(err)
		if !isNotFoundError && err != nil {
			return nil, err
		}
		if isNotFoundError || ghostdagData.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
			isTrustedData = true
			ghostdagData, err = pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, current, true)
			if err != nil {
				return nil, err
			}
		}

		ghostdagDataHashPairs = append(ghostdagDataHashPairs, &externalapi.BlockGHOSTDAGDataHashPair{
			Hash:         current,
			GHOSTDAGData: ghostdagData,
		})

		if ghostdagData.SelectedParent().Equal(pm.genesisHash) {
			break
		}

		if current.Equal(pm.genesisHash) {
			break
		}

		current = ghostdagData.SelectedParent()
	}

	return &externalapi.BlockWithTrustedData{
		Block:        block,
		DAAScore:     daaScore,
		DAAWindow:    windowPairs,
		GHOSTDAGData: ghostdagDataHashPairs,
	}, nil
}

func (pm *pruningManager) ExpectedHeaderPruningPoint(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return nil, err
	}

	if ghostdagData.SelectedParent().Equal(pm.genesisHash) {
		return pm.genesisHash, nil
	}

	selectedParentHeader, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}

	selectedParentPruningPointHeader, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, selectedParentHeader.PruningPoint())
	if err != nil {
		return nil, err
	}

	nextOrCurrentPruningPoint := selectedParentHeader.PruningPoint()
	pruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	// If the block doesn't have the pruning in its selected chain we know for sure that it can't trigger a pruning point
	// change (we check the selected parent to take care of the case where the block is the virtual which doesn't have reachability data).
	hasPruningPointInItsSelectedChain, err := pm.dagTopologyManager.IsInSelectedParentChainOf(stagingArea, pruningPoint, ghostdagData.SelectedParent())
	if err != nil {
		return nil, err
	}

	if hasPruningPointInItsSelectedChain && pm.finalityScore(ghostdagData.BlueScore()) > pm.finalityScore(selectedParentPruningPointHeader.BlueScore()+pm.pruningDepth) {
		var suggestedLowHash *externalapi.DomainHash
		hasReachabilityData, err := pm.reachabilityDataStore.HasReachabilityData(pm.databaseContext, stagingArea, selectedParentHeader.PruningPoint())
		if err != nil {
			return nil, err
		}

		if hasReachabilityData {
			suggestedLowHash = selectedParentHeader.PruningPoint()
		}

		nextOrCurrentPruningPoint, _, err = pm.nextPruningPointAndCandidateByBlockHash(stagingArea, blockHash, suggestedLowHash)
		if err != nil {
			return nil, err
		}
	}

	isHeaderPruningPoint, err := pm.isPruningPointInPruningDepth(stagingArea, blockHash, nextOrCurrentPruningPoint)
	if err != nil {
		return nil, err
	}

	if isHeaderPruningPoint {
		return nextOrCurrentPruningPoint, nil
	}

	pruningPointIndex, err := pm.pruningStore.CurrentPruningPointIndex(pm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	for i := pruningPointIndex; ; i-- {
		currentPruningPoint, err := pm.pruningStore.PruningPointByIndex(pm.databaseContext, stagingArea, i)
		if err != nil {
			return nil, err
		}

		isHeaderPruningPoint, err := pm.isPruningPointInPruningDepth(stagingArea, blockHash, currentPruningPoint)
		if err != nil {
			return nil, err
		}

		if isHeaderPruningPoint {
			return currentPruningPoint, nil
		}

		if i == 0 {
			break
		}
	}

	return pm.genesisHash, nil
}

func (pm *pruningManager) isPruningPointInPruningDepth(stagingArea *model.StagingArea, blockHash, pruningPoint *externalapi.DomainHash) (bool, error) {
	pruningPointHeader, err := pm.blockHeaderStore.BlockHeader(pm.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return false, err
	}

	blockGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, stagingArea, blockHash, false)
	if err != nil {
		return false, err
	}

	return blockGHOSTDAGData.BlueScore() >= pruningPointHeader.BlueScore()+pm.pruningDepth, nil
}
