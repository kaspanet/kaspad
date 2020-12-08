package blockvalidator

import (
	"math/big"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util"
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
	consensusStateManager model.ConsensusStateManager

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
	pruningStore model.PruningStore,

	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	blockHeaderStore model.BlockHeaderStore,
	blockStatusStore model.BlockStatusStore) model.BlockValidator {

	return &blockValidator{
		powMax:                      powMax,
		skipPoW:                     skipPoW,
		genesisHash:                 genesisHash,
		enableNonNativeSubnetworks:  enableNonNativeSubnetworks,
		powMaxBits:                  util.BigToCompact(powMax),
		maxBlockSize:                maxBlockSize,
		mergeSetSizeLimit:           mergeSetSizeLimit,
		maxBlockParents:             maxBlockParents,
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
		pruningStore:                pruningStore,

		blockStore:        blockStore,
		ghostdagDataStore: ghostdagDataStore,
		blockHeaderStore:  blockHeaderStore,
		blockStatusStore:  blockStatusStore,
	}
}
