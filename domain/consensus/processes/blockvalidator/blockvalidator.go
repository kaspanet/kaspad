package blockvalidator

import (
	"github.com/kaspanet/kaspad/util/difficulty"
	"math/big"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockValidator exposes a set of validation classes, after which
// it's possible to determine whether either a block is valid
type blockValidator struct {
	powMax                      *big.Int
	skipPoW                     bool
	genesisHash                 *externalapi.DomainHash
	enableNonNativeSubnetworks  bool
	powMaxBits                  uint32
	maxBlockSize                uint64
	mergeSetSizeLimit           uint64
	maxBlockParents             model.KType
	timestampDeviationTolerance int
	targetTimePerBlock          time.Duration

	databaseContext       model.DBReader
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	coinbaseManager       model.CoinbaseManager
	mergeDepthManager     model.MergeDepthManager
	pruningStore          model.PruningStore
	reachabilityManager   model.ReachabilityManager

	blockStore        model.BlockStore
	ghostdagDataStore model.GHOSTDAGDataStore
	blockHeaderStore  model.BlockHeaderStore
	blockStatusStore  model.BlockStatusStore
	reachabilityStore model.ReachabilityDataStore
}

// New instantiates a new BlockValidator
func New(powMax *big.Int,
	skipPoW bool,
	genesisHash *externalapi.DomainHash,
	enableNonNativeSubnetworks bool,
	maxBlockSize uint64,
	mergeSetSizeLimit uint64,
	maxBlockParents model.KType,
	timestampDeviationTolerance int,
	targetTimePerBlock time.Duration,

	databaseContext model.DBReader,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	coinbaseManager model.CoinbaseManager,
	mergeDepthManager model.MergeDepthManager,
	reachabilityManager model.ReachabilityManager,

	pruningStore model.PruningStore,

	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore,
	reachabilityStore model.ReachabilityDataStore,
) model.BlockValidator {

	return &blockValidator{
		powMax:                     powMax,
		skipPoW:                    skipPoW,
		genesisHash:                genesisHash,
		enableNonNativeSubnetworks: enableNonNativeSubnetworks,
		powMaxBits:                 difficulty.BigToCompact(powMax),
		maxBlockSize:               maxBlockSize,
		mergeSetSizeLimit:          mergeSetSizeLimit,
		maxBlockParents:            maxBlockParents,

		timestampDeviationTolerance: timestampDeviationTolerance,
		targetTimePerBlock:          targetTimePerBlock,
		databaseContext:             databaseContext,
		difficultyManager:           difficultyManager,
		pastMedianTimeManager:       pastMedianTimeManager,
		transactionValidator:        transactionValidator,
		ghostdagManager:             ghostdagManager,
		dagTopologyManager:          dagTopologyManager,
		dagTraversalManager:         dagTraversalManager,
		coinbaseManager:             coinbaseManager,
		mergeDepthManager:           mergeDepthManager,
		reachabilityManager:         reachabilityManager,

		pruningStore:      pruningStore,
		blockStore:        blockStore,
		ghostdagDataStore: ghostdagDataStore,
		blockHeaderStore:  blockHeaderStore,
		blockStatusStore:  blockStatusStore,
		reachabilityStore: reachabilityStore,
	}
}
