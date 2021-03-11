package difficultymanager

import (
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/math"
	"math/big"
	"time"

	"github.com/kaspanet/kaspad/util/difficulty"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DifficultyManager provides a method to resolve the
// difficulty value of a block
type difficultyManager struct {
	databaseContext                model.DBReader
	ghostdagManager                model.GHOSTDAGManager
	ghostdagStore                  model.GHOSTDAGDataStore
	headerStore                    model.BlockHeaderStore
	daaBlocksStore                 model.DAABlocksStore
	dagTopologyManager             model.DAGTopologyManager
	dagTraversalManager            model.DAGTraversalManager
	genesisHash                    *externalapi.DomainHash
	powMax                         *big.Int
	difficultyAdjustmentWindowSize int
	disableDifficultyAdjustment    bool
	targetTimePerBlock             time.Duration
}

// New instantiates a new DifficultyManager
func New(databaseContext model.DBReader,
	ghostdagManager model.GHOSTDAGManager,
	ghostdagStore model.GHOSTDAGDataStore,
	headerStore model.BlockHeaderStore,
	daaBlocksStore model.DAABlocksStore,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	powMax *big.Int,
	difficultyAdjustmentWindowSize int,
	disableDifficultyAdjustment bool,
	targetTimePerBlock time.Duration,
	genesisHash *externalapi.DomainHash) model.DifficultyManager {
	return &difficultyManager{
		databaseContext:                databaseContext,
		ghostdagManager:                ghostdagManager,
		ghostdagStore:                  ghostdagStore,
		headerStore:                    headerStore,
		daaBlocksStore:                 daaBlocksStore,
		dagTopologyManager:             dagTopologyManager,
		dagTraversalManager:            dagTraversalManager,
		powMax:                         powMax,
		difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
		disableDifficultyAdjustment:    disableDifficultyAdjustment,
		targetTimePerBlock:             targetTimePerBlock,
		genesisHash:                    genesisHash,
	}
}

func (dm *difficultyManager) genesisBits() (uint32, error) {
	header, err := dm.headerStore.BlockHeader(dm.databaseContext, dm.genesisHash)
	if err != nil {
		return 0, err
	}

	return header.Bits(), nil
}

// RequiredDifficulty returns the difficulty required for some block
func (dm *difficultyManager) UpdateDAADataAndReturnDifficultyBits(blockHash *externalapi.DomainHash) (uint32, error) {
	// Fetch window of dag.difficultyAdjustmentWindowSize + 1 so we can have dag.difficultyAdjustmentWindowSize block intervals
	targetsWindow, windowHashes, err := dm.blockWindow(blockHash, dm.difficultyAdjustmentWindowSize+1)
	if err != nil {
		return 0, err
	}

	err = dm.updateDaaScoreAndAddedBlocks(blockHash, windowHashes)
	if err != nil {
		return 0, err
	}

	if dm.disableDifficultyAdjustment {
		return dm.genesisBits()
	}

	// We need at least 2 blocks to get a timestamp interval
	// We could instead clamp the timestamp difference to `targetTimePerBlock`,
	// but then everything will cancel out and we'll get the target from the last block, which will be the same as genesis.
	if len(targetsWindow) < 2 {
		return dm.genesisBits()
	}
	windowMinTimestamp, windowMaxTimeStamp, windowsMinIndex, _ := targetsWindow.minMaxTimestamps()
	// Remove the last block from the window so to calculate the average target of dag.difficultyAdjustmentWindowSize blocks
	targetsWindow.remove(windowsMinIndex)

	// Calculate new target difficulty as:
	// averageWindowTarget * (windowMinTimestamp / (targetTimePerBlock * windowSize))
	// The result uses integer division which means it will be slightly
	// rounded down.
	div := new(big.Int)
	newTarget := targetsWindow.averageTarget()
	newTarget.
		// We need to clamp the timestamp difference to 1 so that we'll never get a 0 target.
		Mul(newTarget, div.SetInt64(math.MaxInt64(windowMaxTimeStamp-windowMinTimestamp, 1))).
		Div(newTarget, div.SetInt64(dm.targetTimePerBlock.Milliseconds())).
		Div(newTarget, div.SetUint64(uint64(len(targetsWindow))))
	if newTarget.Cmp(dm.powMax) > 0 {
		return difficulty.BigToCompact(dm.powMax), nil
	}
	newTargetBits := difficulty.BigToCompact(newTarget)
	return newTargetBits, nil
}

func (dm *difficultyManager) updateDaaScoreAndAddedBlocks(blockHash *externalapi.DomainHash,
	windowHashes []*externalapi.DomainHash) error {

	onEnd := logger.LogAndMeasureExecutionTime(log, "updateDaaScoreAndAddedBlocks")
	defer onEnd()

	daaScore, addedBlocks, err := dm.calculateDaaScoreAndAddedBlocks(blockHash, windowHashes)
	if err != nil {
		return err
	}

	dm.daaBlocksStore.StageDAAScore(blockHash, daaScore)
	dm.daaBlocksStore.StageBlockDAAAddedBlocks(blockHash, addedBlocks)
	return nil
}

func (dm *difficultyManager) calculateDaaScoreAndAddedBlocks(blockHash *externalapi.DomainHash,
	windowHashes []*externalapi.DomainHash) (uint64, []*externalapi.DomainHash, error) {

	if blockHash.Equal(dm.genesisHash) {
		return 0, nil, nil
	}

	ghostdagData, err := dm.ghostdagStore.Get(dm.databaseContext, blockHash)
	if err != nil {
		return 0, nil, err
	}

	mergeSet := make(map[externalapi.DomainHash]struct{}, len(ghostdagData.MergeSet()))
	for _, hash := range ghostdagData.MergeSet() {
		mergeSet[*hash] = struct{}{}
	}

	// TODO: Consider optimizing by breaking the loop once you arrive to the
	// window block with blue work higher than all non-added merge set blocks.
	daaAddedBlocks := make([]*externalapi.DomainHash, 0, len(mergeSet))
	for _, hash := range windowHashes {
		if _, exists := mergeSet[*hash]; exists {
			daaAddedBlocks = append(daaAddedBlocks, hash)
			if len(daaAddedBlocks) == len(mergeSet) {
				break
			}
		}
	}

	selectedParentDAAScore, err := dm.daaBlocksStore.DAAScore(dm.databaseContext, ghostdagData.SelectedParent())
	if err != nil {
		return 0, nil, err
	}

	return selectedParentDAAScore + uint64(len(daaAddedBlocks)), daaAddedBlocks, nil
}
