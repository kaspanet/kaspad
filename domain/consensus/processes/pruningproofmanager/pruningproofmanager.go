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
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashset"
	"github.com/kaspanet/kaspad/domain/consensus/utils/pow"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
	"math/big"
)

type pruningProofManager struct {
	databaseContext model.DBManager

	dagTopologyManagers  []model.DAGTopologyManager
	ghostdagManagers     []model.GHOSTDAGManager
	reachabilityManagers []model.ReachabilityManager

	ghostdagDataStores  []model.GHOSTDAGDataStore
	pruningStore        model.PruningStore
	blockHeaderStore    model.BlockHeaderStore
	blockStatusStore    model.BlockStatusStore
	finalityStore       model.FinalityStore
	consensusStateStore model.ConsensusStateStore

	genesisHash   *externalapi.DomainHash
	k             externalapi.KType
	pruningProofM uint64
}

// New instantiates a new PruningManager
func New(
	databaseContext model.DBManager,

	dagTopologyManagers []model.DAGTopologyManager,
	ghostdagManagers []model.GHOSTDAGManager,
	reachabilityManagers []model.ReachabilityManager,

	ghostdagDataStores []model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore,
	finalityStore model.FinalityStore,
	consensusStateStore model.ConsensusStateStore,

	genesisHash *externalapi.DomainHash,
	k externalapi.KType,
	pruningProofM uint64,
) model.PruningProofManager {

	return &pruningProofManager{
		databaseContext:      databaseContext,
		dagTopologyManagers:  dagTopologyManagers,
		ghostdagManagers:     ghostdagManagers,
		reachabilityManagers: reachabilityManagers,

		ghostdagDataStores:  ghostdagDataStores,
		pruningStore:        pruningStore,
		blockHeaderStore:    blockHeaderStore,
		blockStatusStore:    blockStatusStore,
		finalityStore:       finalityStore,
		consensusStateStore: consensusStateStore,

		genesisHash:   genesisHash,
		k:             k,
		pruningProofM: pruningProofM,
	}
}

func (ppm *pruningProofManager) BuildPruningPointProof(stagingArea *model.StagingArea) (*externalapi.PruningPointProof, error) {
	pruningPoint, err := ppm.pruningStore.PruningPoint(ppm.databaseContext, stagingArea)
	if err != nil {
		return nil, err
	}

	pruningPointHeader, err := ppm.blockHeaderStore.BlockHeader(ppm.databaseContext, stagingArea, pruningPoint)
	if err != nil {
		return nil, err
	}

	maxLevel := len(pruningPointHeader.Parents()) - 1
	headersByLevel := make(map[int][]externalapi.BlockHeader)
	selectedTipByLevel := make([]*externalapi.DomainHash, maxLevel+1)
	pruningPointLevel := pow.BlockLevel(pruningPointHeader)
	for blockLevel := maxLevel; blockLevel >= 0; blockLevel-- {
		var selectedTip *externalapi.DomainHash
		if blockLevel <= pruningPointLevel {
			selectedTip = pruningPoint
		} else {
			blockLevelParents := pruningPointHeader.ParentsAtLevel(blockLevel)
			selectedTip, err = ppm.ghostdagManagers[blockLevel].ChooseSelectedParent(stagingArea, []*externalapi.DomainHash(blockLevelParents)...)
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
					// TODO: Is it actually secure?
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
		queue := []*externalapi.DomainHash{root}
		for len(queue) > 0 {
			var current *externalapi.DomainHash
			current, queue = queue[0], queue[1:]

			if visited.Contains(current) {
				continue
			}

			visited.Add(current)
			isAncestorOfPruningPoint, err := ppm.dagTopologyManagers[blockLevel].IsAncestorOf(stagingArea, current, selectedTip)
			if err != nil {
				return nil, err
			}

			if !isAncestorOfPruningPoint {
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

			queue = append(queue, children...)
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
	stagingArea := model.NewStagingArea()
	level0Headers := pruningPointProof.Headers[0]
	pruningPointHeader := level0Headers[len(level0Headers)-1]
	pruningPointBlockLevel := pow.BlockLevel(pruningPointHeader)
	maxLevel := len(pruningPointHeader.Parents()) - 1
	if maxLevel >= len(pruningPointProof.Headers) {
		return errors.Wrapf(ruleerrors.ErrPruningProofMissingBlockLevels, "proof has only %d levels while pruning point "+
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
		headers := make([]externalapi.BlockHeader, len(pruningPointProof.Headers[blockLevel]))
		copy(headers, pruningPointProof.Headers[blockLevel])

		var selectedTip *externalapi.DomainHash
		for i, header := range headers {
			blockHash := consensushashing.HeaderHash(header)
			if pow.BlockLevel(header) < blockLevel {
				return errors.Wrapf(ruleerrors.ErrPruningProofWrongBlockLevel, "block %s level is %d when it's "+
					"expected to be at least %d", blockHash, pow.BlockLevel(header), blockLevel)
			}

			blockHeaderStore.Stage(stagingArea, blockHash, header)

			var parents []*externalapi.DomainHash
			for _, parent := range header.ParentsAtLevel(blockLevel) {
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

		selectedTipByLevel[blockLevel] = selectedTip
	}

	currentDAGPruningPoint, err := ppm.pruningStore.PruningPoint(ppm.databaseContext, model.NewStagingArea())
	if err != nil {
		return err
	}

	for blockLevel, selectedTip := range selectedTipByLevel {
		if blockLevel <= pruningPointBlockLevel {
			if !selectedTip.Equal(consensushashing.HeaderHash(pruningPointHeader)) {
				return errors.Wrapf(ruleerrors.ErrPruningProofSelectedTipIsNotThePruningPoint, "the pruning "+
					"proof selected tip %s at level %d is not the pruning point", selectedTip, blockLevel)
			}
		} else if !pruningPointHeader.ParentsAtLevel(blockLevel).Contains(selectedTip) {
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
			currentDAGPruningPointParents, err := ppm.dagTopologyManagers[blockLevel].Parents(model.NewStagingArea(), currentDAGPruningPoint)
			if err != nil {
				return err
			}

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
			} else {
				return nil
			}
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

	reachabilityManagers := make([]model.ReachabilityManager, constants.MaxBlockLevel+1)
	dagTopologyManagers := make([]model.DAGTopologyManager, constants.MaxBlockLevel+1)
	ghostdagManagers := make([]model.GHOSTDAGManager, constants.MaxBlockLevel+1)

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

func (ppm *pruningProofManager) ApplyPruningPointProof(stagingArea *model.StagingArea, pruningPointProof *externalapi.PruningPointProof) error {
	for blockLevel, headers := range pruningPointProof.Headers {
		var selectedTip *externalapi.DomainHash
		for i, header := range headers {
			blockHash := consensushashing.HeaderHash(header)
			if pow.BlockLevel(header) < blockLevel {
				return errors.Wrapf(ruleerrors.ErrPruningProofWrongBlockLevel, "block %s level is %d when it's "+
					"expected to be at least %d", blockHash, pow.BlockLevel(header), blockLevel)
			}

			ppm.blockHeaderStore.Stage(stagingArea, blockHash, header)

			var parents []*externalapi.DomainHash
			for _, parent := range header.ParentsAtLevel(blockLevel) {
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

			if blockLevel == 0 {
				selectedParent, err := ppm.ghostdagManagers[blockLevel].ChooseSelectedParent(stagingArea, parents...)
				if err != nil {
					return err
				}

				ppm.ghostdagDataStores[0].Stage(stagingArea, blockHash, externalapi.NewBlockGHOSTDAGData(
					header.BlueScore(),
					header.BlueWork(),
					selectedParent,
					nil,
					nil,
					nil,
				), false)

				ppm.finalityStore.StageFinalityPoint(stagingArea, blockHash, model.VirtualGenesisBlockHash)
				ppm.blockStatusStore.Stage(stagingArea, blockHash, externalapi.StatusHeaderOnly)
			} else {
				err = ppm.ghostdagManagers[blockLevel].GHOSTDAG(stagingArea, blockHash)
				if err != nil {
					return err
				}
			}

			if selectedTip == nil {
				selectedTip = blockHash
			} else {
				selectedTip, err = ppm.ghostdagManagers[blockLevel].ChooseSelectedParent(stagingArea, selectedTip, blockHash)
				if err != nil {
					return err
				}
			}

			err = ppm.reachabilityManagers[blockLevel].AddBlock(stagingArea, blockHash)
			if err != nil {
				return err
			}

			if selectedTip.Equal(blockHash) {
				err := ppm.reachabilityManagers[blockLevel].UpdateReindexRoot(stagingArea, selectedTip)
				if err != nil {
					return err
				}
			}
		}
	}

	pruningPointHeader := pruningPointProof.Headers[0][len(pruningPointProof.Headers[0])-1]
	pruningPoint := consensushashing.HeaderHash(pruningPointHeader)
	ppm.consensusStateStore.StageTips(stagingArea, []*externalapi.DomainHash{pruningPoint})
	return nil
}
