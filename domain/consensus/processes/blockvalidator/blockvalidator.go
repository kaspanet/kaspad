package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
	"math/big"
)

// blockValidator exposes a set of validation classes, after which
// it's possible to determine whether either a block is valid
type blockValidator struct {
	powMax                         *big.Int
	skipPoW                        bool
	genesisHash                    *externalapi.DomainHash
	enableNonNativeSubnetworks     bool
	disableDifficultyAdjustment    bool
	powMaxBits                     uint32
	difficultyAdjustmentWindowSize uint64
	finalityDepth                  uint64

	databaseContext       model.DBReader
	consensusStateManager model.ConsensusStateManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager

	blockStore        model.BlockStore
	ghostdagDataStore model.GHOSTDAGDataStore
	blockHeaderStore  model.BlockHeaderStore
}

// New instantiates a new BlockValidator
func New(powMax *big.Int,
	skipPoW bool,
	genesisHash *externalapi.DomainHash,
	enableNonNativeSubnetworks bool,
	disableDifficultyAdjustment bool,
	difficultyAdjustmentWindowSize uint64,
	finalityDepth uint64,
	databaseContext model.DBReader,

	consensusStateManager model.ConsensusStateManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,

	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockHeaderStore model.BlockHeaderStore) model.BlockValidator {

	return &blockValidator{
		powMax:                         powMax,
		skipPoW:                        skipPoW,
		genesisHash:                    genesisHash,
		enableNonNativeSubnetworks:     enableNonNativeSubnetworks,
		disableDifficultyAdjustment:    disableDifficultyAdjustment,
		powMaxBits:                     util.BigToCompact(powMax),
		difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
		finalityDepth:                  finalityDepth,
		databaseContext:                databaseContext,
		consensusStateManager:          consensusStateManager,
		difficultyManager:              difficultyManager,
		pastMedianTimeManager:          pastMedianTimeManager,
		transactionValidator:           transactionValidator,
		ghostdagManager:                ghostdagManager,
		dagTopologyManager:             dagTopologyManager,
		dagTraversalManager:            dagTraversalManager,
		blockStore:                     blockStore,
		ghostdagDataStore:              ghostdagDataStore,
		blockHeaderStore:               blockHeaderStore,
	}
}
