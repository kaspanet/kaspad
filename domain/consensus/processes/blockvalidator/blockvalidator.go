package blockvalidator

import (
	"math/big"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/difficulty"
)

// blockValidator exposes a set of validation classes, after which
// it's possible to determine whether either a block is valid
type blockValidator struct {
	powMax                      *big.Int
	skipPoW                     bool
	genesisHash                 *externalapi.DomainHash
	enableNonNativeSubnetworks  bool
	powMaxBits                  uint32
	maxBlockMass                uint64
	mergeSetSizeLimit           uint64
	maxBlockParents             externalapi.KType
	timestampDeviationTolerance int
	targetTimePerBlock          time.Duration

	databaseContext       model.DBReader
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	ghostdagManagers      []model.GHOSTDAGManager
	dagTopologyManagers   []model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager
	coinbaseManager       model.CoinbaseManager
	mergeDepthManager     model.MergeDepthManager
	pruningStore          model.PruningStore
	reachabilityManagers  []model.ReachabilityManager
	finalityManager       model.FinalityManager
	blockParentBuilder    model.BlockParentBuilder
	pruningManager        model.PruningManager

	blockStore          model.BlockStore
	ghostdagDataStore   model.GHOSTDAGDataStore
	blockHeaderStore    model.BlockHeaderStore
	blockStatusStore    model.BlockStatusStore
	reachabilityStore   model.ReachabilityDataStore
	consensusStateStore model.ConsensusStateStore
	daaBlocksStore      model.DAABlocksStore
}

// New instantiates a new BlockValidator
func New(powMax *big.Int,
	skipPoW bool,
	genesisHash *externalapi.DomainHash,
	enableNonNativeSubnetworks bool,
	maxBlockMass uint64,
	mergeSetSizeLimit uint64,
	maxBlockParents externalapi.KType,
	timestampDeviationTolerance int,
	targetTimePerBlock time.Duration,

	databaseContext model.DBReader,

	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	ghostdagManagers []model.GHOSTDAGManager,
	dagTopologyManagers []model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	coinbaseManager model.CoinbaseManager,
	mergeDepthManager model.MergeDepthManager,
	reachabilityManagers []model.ReachabilityManager,
	finalityManager model.FinalityManager,
	blockParentBuilder model.BlockParentBuilder,
	pruningManager model.PruningManager,

	pruningStore model.PruningStore,
	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore,
	reachabilityStore model.ReachabilityDataStore,
	consensusStateStore model.ConsensusStateStore,
	daaBlocksStore model.DAABlocksStore,
) model.BlockValidator {

	return &blockValidator{
		powMax:                     powMax,
		skipPoW:                    skipPoW,
		genesisHash:                genesisHash,
		enableNonNativeSubnetworks: enableNonNativeSubnetworks,
		powMaxBits:                 difficulty.BigToCompact(powMax),
		maxBlockMass:               maxBlockMass,
		mergeSetSizeLimit:          mergeSetSizeLimit,
		maxBlockParents:            maxBlockParents,

		timestampDeviationTolerance: timestampDeviationTolerance,
		targetTimePerBlock:          targetTimePerBlock,
		databaseContext:             databaseContext,
		difficultyManager:           difficultyManager,
		pastMedianTimeManager:       pastMedianTimeManager,
		transactionValidator:        transactionValidator,
		ghostdagManagers:            ghostdagManagers,
		dagTopologyManagers:         dagTopologyManagers,
		dagTraversalManager:         dagTraversalManager,
		coinbaseManager:             coinbaseManager,
		mergeDepthManager:           mergeDepthManager,
		reachabilityManagers:        reachabilityManagers,
		finalityManager:             finalityManager,
		blockParentBuilder:          blockParentBuilder,
		pruningManager:              pruningManager,

		pruningStore:        pruningStore,
		blockStore:          blockStore,
		ghostdagDataStore:   ghostdagDataStore,
		blockHeaderStore:    blockHeaderStore,
		blockStatusStore:    blockStatusStore,
		reachabilityStore:   reachabilityStore,
		consensusStateStore: consensusStateStore,
		daaBlocksStore:      daaBlocksStore,
	}
}
