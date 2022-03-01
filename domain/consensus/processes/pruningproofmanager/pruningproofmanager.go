package pruningproofmanager

import (
	consensusDB "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/staging"
	"github.com/pkg/errors"
	"math/big"
)

type pruningProofManager struct {
	databaseContext model.DBManager

	dagTopologyManagers  []model.DAGTopologyManager
	ghostdagManagers     []model.GHOSTDAGManager
	reachabilityManager  model.ReachabilityManager
	dagTraversalManagers []model.DAGTraversalManager
	parentsManager       model.ParentsManager

	ghostdagDataStores  []model.GHOSTDAGDataStore
	pruningStore        model.PruningStore
	blockHeaderStore    model.BlockHeaderStore
	blockStatusStore    model.BlockStatusStore
	finalityStore       model.FinalityStore
	consensusStateStore model.ConsensusStateStore
	blockRelationStore  model.BlockRelationStore

	genesisHash   *externalapi.DomainHash
	k             externalapi.KType
	pruningProofM uint64
	maxBlockLevel int

	cachedPruningPoint *externalapi.DomainHash
	cachedProof        *externalapi.PruningPointProof
}

// New instantiates a new PruningManager
func New(
	databaseContext model.DBManager,

	dagTopologyManagers []model.DAGTopologyManager,
	ghostdagManagers []model.GHOSTDAGManager,
	reachabilityManager model.ReachabilityManager,
	dagTraversalManagers []model.DAGTraversalManager,
	parentsManager model.ParentsManager,

	ghostdagDataStores []model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore,
	finalityStore model.FinalityStore,
	consensusStateStore model.ConsensusStateStore,
	blockRelationStore model.BlockRelationStore,

	genesisHash *externalapi.DomainHash,
	k externalapi.KType,
	pruningProofM uint64,
	maxBlockLevel int,
) model.PruningProofManager {

	return &pruningProofManager{
		databaseContext:      databaseContext,
		dagTopologyManagers:  dagTopologyManagers,
		ghostdagManagers:     ghostdagManagers,
		reachabilityManager:  reachabilityManager,
		dagTraversalManagers: dagTraversalManagers,
		parentsManager:       parentsManager,

		ghostdagDataStores:  ghostdagDataStores,
		pruningStore:        pruningStore,
		blockHeaderStore:    blockHeaderStore,
		blockStatusStore:    blockStatusStore,
		finalityStore:       finalityStore,
		consensusStateStore: consensusStateStore,
		blockRelationStore:  blockRelationStore,

		genesisHash:   genesisHash,
		k:             k,
		pruningProofM: pruningProofM,
		maxBlockLevel: maxBlockLevel,
	}
}

func (ppm *pruningProofManager) BuildPruningPointProof(stagingArea *model.StagingArea) (*externalapi.PruningPointProof, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "BuildPruningPointProof")
	defer onEnd()

	pruningPoint, err := ppm.pruningStore.PruningPoint(ppm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	if ppm.cachedPruningPoint != nil && ppm.cachedPruningPoint.Equal(pruningPoint) {
		return ppm.cachedProof, nil
	}

	proof, err := ppm.buildPruningPointProof(stagingArea)
	if err != nil {
		return nil, err
	}

	ppm.cachedProof = proof
	ppm.cachedPruningPoint = pruningPoint

	return proof, nil
}

func (ppm *pruningProofManager) buildPruningPointProof(stagingArea *model.StagingArea) (*externalapi.PruningPointProof, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "buildPruningPointProof")
	defer onEnd()

	pruningPoint, err := ppm.pruningStore.PruningPoint(ppm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	if pruningPoint.Equal(ppm.genesisHash) {
		return &externalapi.PruningPointProof{}, nil
	}

	pruningPointHeader, err := ppm.blockHeaderStore.BlockHeader(ppm.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}

	maxLevel := len(ppm.parentsManager.Parents(pruningPointHeader)) - 1
	headersByLevel := make(map[int][]externalapi.BlockHeader)
	selectedTipByLevel := make([]*externalapi.DomainHash, maxLevel+1)
	pruningPointLevel := pruningPointHeader.BlockLevel(ppm.maxBlockLevel)
	for blockLevel := maxLevel; blockLevel >= 0; blockLevel-- {
		var selectedTip *externalapi.DomainHash
		if blockLevel <= pruningPointLevel {
			selectedTip = pruningPoint
		} else {
			blockLevelParents := ppm.parentsManager.ParentsAtLevel(pruningPointHeader, blockLevel)
			selectedTipCandidates := make([]*externalapi.DomainHash, 0, len(blockLevelParents))

			// In a pruned node, some pruning point parents might be missing, but we're guaranteed that its
			// selected parent is not missing.
			for _, parent := range blockLevelParents {
				_, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, parent, false)
				if database.IsNotFoundError(err) {
					continue
				}
				if err != nil {
					return nil, err
				}

				selectedTipCandidates = append(selectedTipCandidates, parent)
			}

			selectedTip, err = ppm.ghostdagManagers[blockLevel].ChooseSelectedParent(stagingArea, selectedTipCandidates...)
			if err != nil {
				return nil, err
			}
		}
		selectedTipByLevel[blockLevel] = selectedTip

		blockAtDepth2M, err := ppm.blockAtDepth(stagingArea, ppm.ghostdagDataStores[blockLevel], selectedTip, 2*ppm.pruningProofM)
		if err != nil {
			return nil, err
		}

		root := blockAtDepth2M
		if blockLevel != maxLevel {
			blockAtDepthMAtNextLevel, err := ppm.blockAtDepth(stagingArea, ppm.ghostdagDataStores[blockLevel+1], selectedTipByLevel[blockLevel+1], ppm.pruningProofM)
			if err != nil {
				return nil, err
			}

			isBlockAtDepthMAtNextLevelAncestorOfBlockAtDepth2M, err := ppm.dagTopologyManagers[blockLevel].IsAncestorOf(stagingArea, blockAtDepthMAtNextLevel, blockAtDepth2M)
			if err != nil {
				return nil, err
			}

			if isBlockAtDepthMAtNextLevelAncestorOfBlockAtDepth2M {
				root = blockAtDepthMAtNextLevel
			} else {
				isBlockAtDepth2MAncestorOfBlockAtDepthMAtNextLevel, err := ppm.dagTopologyManagers[blockLevel].IsAncestorOf(stagingArea, blockAtDepth2M, blockAtDepthMAtNextLevel)
				if err != nil {
					return nil, err
				}

				if !isBlockAtDepth2MAncestorOfBlockAtDepthMAtNextLevel {
					// find common ancestor
					current := blockAtDepthMAtNextLevel
					for {
						ghostdagData, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, current, false)
						if err != nil {
							return nil, err
						}

						current = ghostdagData.SelectedParent()
						if current.Equal(model.VirtualGenesisBlockHash) {
							return nil, errors.Errorf("No common ancestor between %s and %s at level %d", blockAtDepth2M, blockAtDepthMAtNextLevel, blockLevel)
						}

						isCurrentAncestorOfBlockAtDepth2M, err := ppm.dagTopologyManagers[blockLevel].IsAncestorOf(stagingArea, current, blockAtDepth2M)
						if err != nil {
							return nil, err
						}

						if isCurrentAncestorOfBlockAtDepth2M {
							root = current
							break
						}
					}
				}
			}
		}

		headers := make([]externalapi.BlockHeader, 0, 2*ppm.pruningProofM)
		visited := hashset.New()
		queue := ppm.dagTraversalManagers[blockLevel].NewUpHeap(stagingArea)
		err = queue.Push(root)
		if err != nil {
			return nil, err
		}
		for queue.Len() > 0 {
			current := queue.Pop()

			if visited.Contains(current) {
				continue
			}

			visited.Add(current)
			isAncestorOfSelectedTip, err := ppm.dagTopologyManagers[blockLevel].IsAncestorOf(stagingArea, current, selectedTip)
			if err != nil {
				return nil, err
			}

			if !isAncestorOfSelectedTip {
				continue
			}

			currentHeader, err := ppm.blockHeaderStore.BlockHeader(ppm.databaseContext, stagingArea, current)
			if err != nil {
				return nil, err
			}

			headers = append(headers, currentHeader)
			children, err := ppm.dagTopologyManagers[blockLevel].Children(stagingArea, current)
			if err != nil {
				return nil, err
			}

			err = queue.PushSlice(children)
			if err != nil {
				return nil, err
			}
		}

		headersByLevel[blockLevel] = headers
	}

	proof := &externalapi.PruningPointProof{Headers: make([][]externalapi.BlockHeader, len(headersByLevel))}
	for i := 0; i < len(headersByLevel); i++ {
		proof.Headers[i] = headersByLevel[i]
	}

	return proof, nil
}

func (ppm *pruningProofManager) blockAtDepth(stagingArea *model.StagingArea, ghostdagDataStore model.GHOSTDAGDataStore, highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error) {
	currentBlockHash := highHash
	highBlockGHOSTDAGData, err := ghostdagDataStore.Get(ppm.databaseContext, stagingArea, highHash, false)
	if err != nil {
		return nil, err
	}

	requiredBlueScore := uint64(0)
	if highBlockGHOSTDAGData.BlueScore() > depth {
		requiredBlueScore = highBlockGHOSTDAGData.BlueScore() - depth
	}

	currentBlockGHOSTDAGData := highBlockGHOSTDAGData
	// If we used `BlockIterator` we'd need to do more calls to `ghostdagDataStore` so we can get the blueScore
	for currentBlockGHOSTDAGData.BlueScore() >= requiredBlueScore {
		if currentBlockGHOSTDAGData.SelectedParent().Equal(model.VirtualGenesisBlockHash) {
			break
		}

		currentBlockHash = currentBlockGHOSTDAGData.SelectedParent()
		currentBlockGHOSTDAGData, err = ghostdagDataStore.Get(ppm.databaseContext, stagingArea, currentBlockHash, false)
		if err != nil {
			return nil, err
		}
	}
	return currentBlockHash, nil
}

func (ppm *pruningProofManager) ValidatePruningPointProof(pruningPointProof *externalapi.PruningPointProof) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidatePruningPointProof")
	defer onEnd()

	stagingArea := model.NewStagingArea()

	if len(pruningPointProof.Headers) == 0 {
		return errors.Wrap(ruleerrors.ErrPruningProofEmpty, "pruning proof is empty")
	}

	level0Headers := pruningPointProof.Headers[0]
	pruningPointHeader := level0Headers[len(level0Headers)-1]
	pruningPoint := consensushashing.HeaderHash(pruningPointHeader)
	pruningPointBlockLevel := pruningPointHeader.BlockLevel(ppm.maxBlockLevel)
	maxLevel := len(ppm.parentsManager.Parents(pruningPointHeader)) - 1
	if maxLevel >= len(pruningPointProof.Headers) {
		return errors.Wrapf(ruleerrors.ErrPruningProofEmpty, "proof has only %d levels while pruning point "+
			"has parents from %d levels", len(pruningPointProof.Headers), maxLevel+1)
	}

	blockHeaderStore, blockRelationStores, reachabilityDataStores, ghostdagDataStores, err := ppm.dagStores(maxLevel)
	if err != nil {
		return err
	}

	reachabilityManagers, dagTopologyManagers, ghostdagManagers := ppm.dagProcesses(maxLevel, blockHeaderStore, blockRelationStores, reachabilityDataStores, ghostdagDataStores)

	for blockLevel := 0; blockLevel <= maxLevel; blockLevel++ {
		err := reachabilityManagers[blockLevel].Init(stagingArea)
		if err != nil {
			return err
		}

		err = dagTopologyManagers[blockLevel].SetParents(stagingArea, model.VirtualGenesisBlockHash, nil)
		if err != nil {
			return err
		}

		ghostdagDataStores[blockLevel].Stage(stagingArea, model.VirtualGenesisBlockHash, externalapi.NewBlockGHOSTDAGData(
			0,
			big.NewInt(0),
			nil,
			nil,
			nil,
			nil,
		), false)
	}

	selectedTipByLevel := make([]*externalapi.DomainHash, maxLevel+1)
	for blockLevel := maxLevel; blockLevel >= 0; blockLevel-- {
		log.Infof("Validating level %d from the pruning point proof", blockLevel)
		headers := make([]externalapi.BlockHeader, len(pruningPointProof.Headers[blockLevel]))
		copy(headers, pruningPointProof.Headers[blockLevel])

		var selectedTip *externalapi.DomainHash
		for i, header := range headers {
			blockHash := consensushashing.HeaderHash(header)
			if header.BlockLevel(ppm.maxBlockLevel) < blockLevel {
				return errors.Wrapf(ruleerrors.ErrPruningProofWrongBlockLevel, "block %s level is %d when it's "+
					"expected to be at least %d", blockHash, header.BlockLevel(ppm.maxBlockLevel), blockLevel)
			}

			blockHeaderStore.Stage(stagingArea, blockHash, header)

			var parents []*externalapi.DomainHash
			for _, parent := range ppm.parentsManager.ParentsAtLevel(header, blockLevel) {
				_, err := ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, parent, false)
				if database.IsNotFoundError(err) {
					continue
				}
				if err != nil {
					return err
				}

				parents = append(parents, parent)
			}

			if len(parents) == 0 {
				if i != 0 {
					return errors.Wrapf(ruleerrors.ErrPruningProofHeaderWithNoKnownParents, "the proof header "+
						"%s is missing known parents", blockHash)
				}
				parents = append(parents, model.VirtualGenesisBlockHash)
			}

			err := dagTopologyManagers[blockLevel].SetParents(stagingArea, blockHash, parents)
			if err != nil {
				return err
			}

			err = ghostdagManagers[blockLevel].GHOSTDAG(stagingArea, blockHash)
			if err != nil {
				return err
			}

			if selectedTip == nil {
				selectedTip = blockHash
			} else {
				selectedTip, err = ghostdagManagers[blockLevel].ChooseSelectedParent(stagingArea, selectedTip, blockHash)
				if err != nil {
					return err
				}
			}

			err = reachabilityManagers[blockLevel].AddBlock(stagingArea, blockHash)
			if err != nil {
				return err
			}

			if selectedTip.Equal(blockHash) {
				err := reachabilityManagers[blockLevel].UpdateReindexRoot(stagingArea, selectedTip)
				if err != nil {
					return err
				}
			}
		}

		if blockLevel < maxLevel {
			blockAtDepthMAtNextLevel, err := ppm.blockAtDepth(stagingArea, ghostdagDataStores[blockLevel+1], selectedTipByLevel[blockLevel+1], ppm.pruningProofM)
			if err != nil {
				return err
			}

			hasBlockAtDepthMAtNextLevel, err := blockRelationStores[blockLevel].Has(ppm.databaseContext, stagingArea, blockAtDepthMAtNextLevel)
			if err != nil {
				return err
			}

			if !hasBlockAtDepthMAtNextLevel {
				return errors.Wrapf(ruleerrors.ErrPruningProofMissingBlockAtDepthMFromNextLevel, "proof level %d "+
					"is missing the block at depth m in level %d", blockLevel, blockLevel+1)
			}
		}

		if !selectedTip.Equal(pruningPoint) && !ppm.parentsManager.ParentsAtLevel(pruningPointHeader, blockLevel).Contains(selectedTip) {
			return errors.Wrapf(ruleerrors.ErrPruningProofMissesBlocksBelowPruningPoint, "the selected tip %s at "+
				"level %d is not a parent of the pruning point", selectedTip, blockLevel)
		}
		selectedTipByLevel[blockLevel] = selectedTip
	}

	currentDAGPruningPoint, err := ppm.pruningStore.PruningPoint(ppm.databaseContext, model.NewStagingArea())
	if err != nil {
		return err
	}

	currentDAGPruningPointHeader, err := ppm.blockHeaderStore.BlockHeader(ppm.databaseContext, model.NewStagingArea(), currentDAGPruningPoint)
	if err != nil {
		return err
	}

	for blockLevel, selectedTip := range selectedTipByLevel {
		if blockLevel <= pruningPointBlockLevel {
			if !selectedTip.Equal(consensushashing.HeaderHash(pruningPointHeader)) {
				return errors.Wrapf(ruleerrors.ErrPruningProofSelectedTipIsNotThePruningPoint, "the pruning "+
					"proof selected tip %s at level %d is not the pruning point", selectedTip, blockLevel)
			}
		} else if !ppm.parentsManager.ParentsAtLevel(pruningPointHeader, blockLevel).Contains(selectedTip) {
			return errors.Wrapf(ruleerrors.ErrPruningProofSelectedTipNotParentOfPruningPoint, "the pruning "+
				"proof selected tip %s at level %d is not a parent of the of the pruning point on the same "+
				"level", selectedTip, blockLevel)
		}

		selectedTipGHOSTDAGData, err := ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, selectedTip, false)
		if err != nil {
			return err
		}

		if selectedTipGHOSTDAGData.BlueScore() < 2*ppm.pruningProofM {
			continue
		}

		current := selectedTip
		currentGHOSTDAGData := selectedTipGHOSTDAGData
		var commonAncestor *externalapi.DomainHash
		var commonAncestorGHOSTDAGData *externalapi.BlockGHOSTDAGData
		var currentDAGCommonAncestorGHOSTDAGData *externalapi.BlockGHOSTDAGData
		for {
			currentDAGHOSTDAGData, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, model.NewStagingArea(), current, false)
			if err == nil {
				commonAncestor = current
				commonAncestorGHOSTDAGData = currentGHOSTDAGData
				currentDAGCommonAncestorGHOSTDAGData = currentDAGHOSTDAGData
				break
			}

			if !database.IsNotFoundError(err) {
				return err
			}

			current = currentGHOSTDAGData.SelectedParent()
			if current.Equal(model.VirtualGenesisBlockHash) {
				break
			}

			currentGHOSTDAGData, err = ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, current, false)
			if err != nil {
				return err
			}
		}

		if commonAncestor != nil {
			selectedTipBlueWorkDiff := big.NewInt(0).Sub(selectedTipGHOSTDAGData.BlueWork(), commonAncestorGHOSTDAGData.BlueWork())
			currentDAGPruningPointParents := ppm.parentsManager.ParentsAtLevel(currentDAGPruningPointHeader, blockLevel)

			foundBetterParent := false
			for _, parent := range currentDAGPruningPointParents {
				parentGHOSTDAGData, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, model.NewStagingArea(), parent, false)
				if err != nil {
					return err
				}

				parentBlueWorkDiff := big.NewInt(0).Sub(parentGHOSTDAGData.BlueWork(), currentDAGCommonAncestorGHOSTDAGData.BlueWork())
				if parentBlueWorkDiff.Cmp(selectedTipBlueWorkDiff) >= 0 {
					foundBetterParent = true
					break
				}
			}

			if foundBetterParent {
				return errors.Wrapf(ruleerrors.ErrPruningProofInsufficientBlueWork, "the proof doesn't "+
					"have sufficient blue work in order to replace the current DAG")
			}
			return nil
		}
	}

	for blockLevel := maxLevel; blockLevel >= 0; blockLevel-- {
		currentDAGPruningPointParents, err := ppm.dagTopologyManagers[blockLevel].Parents(model.NewStagingArea(), currentDAGPruningPoint)
		// If the current pruning point doesn't have a parent at this level, we consider the proof state to be better.
		if database.IsNotFoundError(err) {
			return nil
		}
		if err != nil {
			return err
		}

		for _, parent := range currentDAGPruningPointParents {
			parentGHOSTDAGData, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, model.NewStagingArea(), parent, false)
			if err != nil {
				return err
			}

			if parentGHOSTDAGData.BlueScore() < 2*ppm.pruningProofM {
				return nil
			}
		}
	}

	return errors.Wrapf(ruleerrors.ErrPruningProofInsufficientBlueWork, "the pruning proof doesn't have any "+
		"shared blocks with the known DAGs, but doesn't have enough headers from levels higher than the existing block levels.")
}

func (ppm *pruningProofManager) dagStores(maxLevel int) (model.BlockHeaderStore, []model.BlockRelationStore, []model.ReachabilityDataStore, []model.GHOSTDAGDataStore, error) {
	blockRelationStores := make([]model.BlockRelationStore, maxLevel+1)
	reachabilityDataStores := make([]model.ReachabilityDataStore, maxLevel+1)
	ghostdagDataStores := make([]model.GHOSTDAGDataStore, maxLevel+1)

	prefix := consensusDB.MakeBucket([]byte("pruningProofManager"))
	blockHeaderStore, err := blockheaderstore.New(ppm.databaseContext, prefix, 0, false)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	for i := 0; i <= maxLevel; i++ {
		blockRelationStores[i] = blockrelationstore.New(prefix, 0, false)
		reachabilityDataStores[i] = reachabilitydatastore.New(prefix, 0, false)
		ghostdagDataStores[i] = ghostdagdatastore.New(prefix, 0, false)
	}

	return blockHeaderStore, blockRelationStores, reachabilityDataStores, ghostdagDataStores, nil
}

func (ppm *pruningProofManager) dagProcesses(
	maxLevel int,
	blockHeaderStore model.BlockHeaderStore,
	blockRelationStores []model.BlockRelationStore,
	reachabilityDataStores []model.ReachabilityDataStore,
	ghostdagDataStores []model.GHOSTDAGDataStore) (
	[]model.ReachabilityManager,
	[]model.DAGTopologyManager,
	[]model.GHOSTDAGManager,
) {

	reachabilityManagers := make([]model.ReachabilityManager, ppm.maxBlockLevel+1)
	dagTopologyManagers := make([]model.DAGTopologyManager, ppm.maxBlockLevel+1)
	ghostdagManagers := make([]model.GHOSTDAGManager, ppm.maxBlockLevel+1)

	for i := 0; i <= maxLevel; i++ {
		reachabilityManagers[i] = reachabilitymanager.New(
			ppm.databaseContext,
			ghostdagDataStores[i],
			reachabilityDataStores[i])

		dagTopologyManagers[i] = dagtopologymanager.New(
			ppm.databaseContext,
			reachabilityManagers[i],
			blockRelationStores[i],
			ghostdagDataStores[i])

		ghostdagManagers[i] = ghostdagmanager.New(
			ppm.databaseContext,
			dagTopologyManagers[i],
			ghostdagDataStores[i],
			blockHeaderStore,
			ppm.k,
			ppm.genesisHash)
	}

	return reachabilityManagers, dagTopologyManagers, ghostdagManagers
}

func (ppm *pruningProofManager) populateProofReachabilityAndHeaders(pruningPointProof *externalapi.PruningPointProof) error {
	// We build a DAG of all multi-level relations between blocks in the proof. We make a upHeap of all blocks, so we can iterate
	// over them in a topological way, and then build a DAG where we use all multi-level parents of a block to create edges, except
	// parents that are already in the past of another parent (This can happen between two levels). We run GHOSTDAG on each block of
	// that DAG, because GHOSTDAG is a requirement to calculate reachability. We then dismiss the GHOSTDAG data because it's not related
	// to the GHOSTDAG data of the real DAG, and was used only for reachability.

	// We need two staging areas: stagingArea which is used to commit the reachability data, and tmpStagingArea for the GHOSTDAG data
	// of allProofBlocksUpHeap. The reason we need two areas is that we use the real GHOSTDAG data in order to order the heap in a topological
	// way, and fake GHOSTDAG data for calculating reachability.
	stagingArea := model.NewStagingArea()
	tmpStagingArea := model.NewStagingArea()

	ghostdagDataStore := ghostdagdatastore.New(consensusDB.MakeBucket(nil), 0, false)
	ghostdagManager := ghostdagmanager.New(nil, nil, ghostdagDataStore, nil, 0, nil)
	dagTraversalManager := dagtraversalmanager.New(nil, nil, ghostdagDataStore, nil, ghostdagManager, nil, nil, nil, 0)
	allProofBlocksUpHeap := dagTraversalManager.NewUpHeap(tmpStagingArea)
	dag := make(map[externalapi.DomainHash]struct {
		parents hashset.HashSet
		header  externalapi.BlockHeader
	})
	for _, headers := range pruningPointProof.Headers {
		for _, header := range headers {
			blockHash := consensushashing.HeaderHash(header)
			if _, ok := dag[*blockHash]; ok {
				continue
			}

			dag[*blockHash] = struct {
				parents hashset.HashSet
				header  externalapi.BlockHeader
			}{parents: hashset.New(), header: header}

			for level := 0; level <= ppm.maxBlockLevel; level++ {
				for _, parent := range ppm.parentsManager.ParentsAtLevel(header, level) {
					parent := parent
					dag[*blockHash].parents.Add(parent)
				}
			}

			// We stage temporary GHOSTDAG data that is needed in order to sort allProofBlocksUpHeap.
			ghostdagDataStore.Stage(tmpStagingArea, blockHash, externalapi.NewBlockGHOSTDAGData(header.BlueScore(), header.BlueWork(), nil, nil, nil, nil), false)
			err := allProofBlocksUpHeap.Push(blockHash)
			if err != nil {
				return err
			}
		}
	}

	dagTopologyManager := dagtopologymanager.New(nil, ppm.reachabilityManager, nil, nil)

	var selectedTip *externalapi.DomainHash
	for allProofBlocksUpHeap.Len() > 0 {
		blockHash := allProofBlocksUpHeap.Pop()
		block := dag[*blockHash]
		ppm.blockHeaderStore.Stage(stagingArea, blockHash, block.header)
		parentsHeap := dagTraversalManager.NewDownHeap(tmpStagingArea)
		for parent := range block.parents {
			parent := parent
			if _, ok := dag[parent]; !ok {
				continue
			}

			err := parentsHeap.Push(&parent)
			if err != nil {
				return err
			}
		}

		fakeParents := []*externalapi.DomainHash{}
		for parentsHeap.Len() > 0 {
			parent := parentsHeap.Pop()
			isAncestorOfAny, err := dagTopologyManager.IsAncestorOfAny(stagingArea, parent, fakeParents)
			if err != nil {
				return err
			}

			if isAncestorOfAny {
				continue
			}

			fakeParents = append(fakeParents, parent)
		}

		if len(fakeParents) == 0 {
			fakeParents = append(fakeParents, model.VirtualGenesisBlockHash)
		}

		err := ppm.dagTopologyManagers[0].SetParents(stagingArea, blockHash, fakeParents)
		if err != nil {
			return err
		}

		err = ppm.ghostdagManagers[0].GHOSTDAG(stagingArea, blockHash)
		if err != nil {
			return err
		}

		err = ppm.reachabilityManager.AddBlock(stagingArea, blockHash)
		if err != nil {
			return err
		}

		if selectedTip == nil {
			selectedTip = blockHash
		} else {
			selectedTip, err = ppm.ghostdagManagers[0].ChooseSelectedParent(stagingArea, selectedTip, blockHash)
			if err != nil {
				return err
			}
		}

		if selectedTip.Equal(blockHash) {
			err := ppm.reachabilityManager.UpdateReindexRoot(stagingArea, selectedTip)
			if err != nil {
				return err
			}
		}
	}

	ppm.ghostdagDataStores[0].UnstageAll(stagingArea)
	ppm.blockRelationStore.UnstageAll(stagingArea)
	err := staging.CommitAllChanges(ppm.databaseContext, stagingArea)
	if err != nil {
		return err
	}
	return nil
}

// ApplyPruningPointProof applies the given pruning proof to the current consensus. Specifically,
// it's meant to be used against the StagingConsensus during headers-proof IBD. Note that for
// performance reasons this operation is NOT atomic. If the process fails for whatever reason
// (e.g. the process was killed) then the database for this consensus MUST be discarded.
func (ppm *pruningProofManager) ApplyPruningPointProof(pruningPointProof *externalapi.PruningPointProof) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ApplyPruningPointProof")
	defer onEnd()

	err := ppm.populateProofReachabilityAndHeaders(pruningPointProof)
	if err != nil {
		return err
	}

	for blockLevel, headers := range pruningPointProof.Headers {
		log.Infof("Applying level %d from the pruning point proof", blockLevel)
		for i, header := range headers {
			if i%1000 == 0 {
				log.Infof("Applying level %d from the pruning point proof - applied %d headers out of %d", blockLevel, i, len(headers))
			}
			stagingArea := model.NewStagingArea()

			blockHash := consensushashing.HeaderHash(header)
			if header.BlockLevel(ppm.maxBlockLevel) < blockLevel {
				return errors.Wrapf(ruleerrors.ErrPruningProofWrongBlockLevel, "block %s level is %d when it's "+
					"expected to be at least %d", blockHash, header.BlockLevel(ppm.maxBlockLevel), blockLevel)
			}

			ppm.blockHeaderStore.Stage(stagingArea, blockHash, header)

			var parents []*externalapi.DomainHash
			for _, parent := range ppm.parentsManager.ParentsAtLevel(header, blockLevel) {
				_, err := ppm.ghostdagDataStores[blockLevel].Get(ppm.databaseContext, stagingArea, parent, false)
				if database.IsNotFoundError(err) {
					continue
				}
				if err != nil {
					return err
				}

				parents = append(parents, parent)
			}

			if len(parents) == 0 {
				if i != 0 {
					return errors.Wrapf(ruleerrors.ErrPruningProofHeaderWithNoKnownParents, "the proof header "+
						"%s is missing known parents", blockHash)
				}
				parents = append(parents, model.VirtualGenesisBlockHash)
			}

			err := ppm.dagTopologyManagers[blockLevel].SetParents(stagingArea, blockHash, parents)
			if err != nil {
				return err
			}

			err = ppm.ghostdagManagers[blockLevel].GHOSTDAG(stagingArea, blockHash)
			if err != nil {
				return err
			}

			if blockLevel == 0 {
				// Override the ghostdag data with the real blue score and blue work
				ghostdagData, err := ppm.ghostdagDataStores[0].Get(ppm.databaseContext, stagingArea, blockHash, false)
				if err != nil {
					return err
				}

				ppm.ghostdagDataStores[0].Stage(stagingArea, blockHash, externalapi.NewBlockGHOSTDAGData(
					header.BlueScore(),
					header.BlueWork(),
					ghostdagData.SelectedParent(),
					ghostdagData.MergeSetBlues(),
					ghostdagData.MergeSetReds(),
					ghostdagData.BluesAnticoneSizes(),
				), false)

				ppm.finalityStore.StageFinalityPoint(stagingArea, blockHash, model.VirtualGenesisBlockHash)
				ppm.blockStatusStore.Stage(stagingArea, blockHash, externalapi.StatusHeaderOnly)
			}

			err = staging.CommitAllChanges(ppm.databaseContext, stagingArea)
			if err != nil {
				return err
			}
		}
	}

	pruningPointHeader := pruningPointProof.Headers[0][len(pruningPointProof.Headers[0])-1]
	pruningPoint := consensushashing.HeaderHash(pruningPointHeader)

	stagingArea := model.NewStagingArea()
	ppm.consensusStateStore.StageTips(stagingArea, []*externalapi.DomainHash{pruningPoint})
	return staging.CommitAllChanges(ppm.databaseContext, stagingArea)
}
