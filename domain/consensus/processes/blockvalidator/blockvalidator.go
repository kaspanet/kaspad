package blockvalidator

import (
	"math/big"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
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

	databaseContext       model.DBReader
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	coinbaseManager       model.CoinbaseManager
	mergeDepthManager     model.MergeDepthManager
	pruningManager        model.PruningManager

	blockStore        model.BlockStore
	ghostdagDataStore model.GHOSTDAGDataStore
	blockHeaderStore  model.BlockHeaderStore
	blockStatusStore  model.BlockStatusStore
}

// New instantiates a new BlockValidator
func New(powMax *big.Int,
	skipPoW bool,
	genesisHash *externalapi.DomainHash,
	enableNonNativeSubnetworks bool,
	disableDifficultyAdjustment bool,
	difficultyAdjustmentWindowSize uint64,
	databaseContext model.DBReader,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	coinbaseManager model.CoinbaseManager,
	mergeDepthManager model.MergeDepthManager,
	pruningManager model.PruningManager,

	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore) model.BlockValidator {

	return &blockValidator{
		powMax:                         powMax,
		skipPoW:                        skipPoW,
		genesisHash:                    genesisHash,
		enableNonNativeSubnetworks:     enableNonNativeSubnetworks,
		disableDifficultyAdjustment:    disableDifficultyAdjustment,
		powMaxBits:                     util.BigToCompact(powMax),
		difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize,
		databaseContext:                databaseContext,
		difficultyManager:              difficultyManager,
		pastMedianTimeManager:          pastMedianTimeManager,
		transactionValidator:           transactionValidator,
		ghostdagManager:                ghostdagManager,
		dagTopologyManager:             dagTopologyManager,
		dagTraversalManager:            dagTraversalManager,
		coinbaseManager:                coinbaseManager,
		mergeDepthManager:              mergeDepthManager,
		pruningManager:                 pruningManager,

		blockStore:        blockStore,
		ghostdagDataStore: ghostdagDataStore,
		blockHeaderStore:  blockHeaderStore,
		blockStatusStore:  blockStatusStore,
	}
}
