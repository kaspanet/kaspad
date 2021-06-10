package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sync"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	daablocksstore "github.com/kaspanet/kaspad/domain/consensus/datastructures/daablocksstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/finalitystore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headersselectedchainstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headersselectedtipstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/multisetstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/pruningstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockbuilder"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/finalitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/headersselectedtipmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/mergedepthmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/syncmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
)

const (
	defaultTestLeveldbCacheSizeMiB = 8
	defaultPreallocateCaches       = true
	defaultTestPreallocateCaches   = false
)

// Config is the full config required to run consensus
type Config struct {
	dagconfig.Params
	// IsArchival tells the consensus if it should not prune old blocks
	IsArchival bool
	// EnableSanityCheckPruningUTXOSet checks the full pruning point utxo set against the commitment at every pruning movement
	EnableSanityCheckPruningUTXOSet bool
}

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(config *Config, db infrastructuredatabase.Database) (
		externalapi.Consensus, error)
	NewTestConsensus(config *Config, testName string) (
		tc testapi.TestConsensus, teardown func(keepDataDir bool), err error)

	SetTestDataDir(dataDir string)
	SetTestGHOSTDAGManager(ghostdagConstructor GHOSTDAGManagerConstructor)
	SetTestLevelDBCacheSize(cacheSizeMiB int)
	SetTestPreAllocateCache(preallocateCaches bool)
	SetTestPastMedianTimeManager(medianTimeConstructor PastMedianTimeManagerConstructor)
	SetTestDifficultyManager(difficultyConstructor DifficultyManagerConstructor)
}

type factory struct {
	dataDir                  string
	ghostdagConstructor      GHOSTDAGManagerConstructor
	pastMedianTimeConsructor PastMedianTimeManagerConstructor
	difficultyConstructor    DifficultyManagerConstructor
	cacheSizeMiB             *int
	preallocateCaches        *bool
}

// NewFactory creates a new Consensus factory
func NewFactory() Factory {
	return &factory{
		ghostdagConstructor:      ghostdagmanager.New,
		pastMedianTimeConsructor: pastmediantimemanager.New,
		difficultyConstructor:    difficultymanager.New,
	}
}

// NewConsensus instantiates a new Consensus
func (f *factory) NewConsensus(config *Config, db infrastructuredatabase.Database) (
	externalapi.Consensus, error) {

	dbManager := consensusdatabase.New(db)

	pruningWindowSizeForCaches := int(config.PruningDepth())

	var preallocateCaches bool
	if f.preallocateCaches != nil {
		preallocateCaches = *f.preallocateCaches
	} else {
		preallocateCaches = defaultPreallocateCaches
	}

	// This is used for caches that are used as part of deletePastBlocks that need to traverse until
	// the previous pruning point.
	pruningWindowSizePlusFinalityDepthForCache := int(config.PruningDepth() + config.FinalityDepth())

	// Data Structures
	acceptanceDataStore := acceptancedatastore.New(200, preallocateCaches)
	blockStore, err := blockstore.New(dbManager, 200, preallocateCaches)
	if err != nil {
		return nil, err
	}
	blockHeaderStore, err := blockheaderstore.New(dbManager, 10_000, preallocateCaches)
	if err != nil {
		return nil, err
	}
	blockRelationStore := blockrelationstore.New(pruningWindowSizePlusFinalityDepthForCache, preallocateCaches)

	blockStatusStore := blockstatusstore.New(pruningWindowSizePlusFinalityDepthForCache, preallocateCaches)
	multisetStore := multisetstore.New(200, preallocateCaches)
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New(pruningWindowSizePlusFinalityDepthForCache, preallocateCaches)
	utxoDiffStore := utxodiffstore.New(200, preallocateCaches)
	consensusStateStore := consensusstatestore.New(10_000, preallocateCaches)

	// Some tests artificially decrease the pruningWindowSize, thus making the GhostDagStore cache too small for a
	// a single DifficultyAdjustmentWindow. To alleviate this problem we make sure that the cache size is at least
	// dagParams.DifficultyAdjustmentWindowSize
	ghostdagDataCacheSize := pruningWindowSizeForCaches
	if ghostdagDataCacheSize < config.DifficultyAdjustmentWindowSize {
		ghostdagDataCacheSize = config.DifficultyAdjustmentWindowSize
	}
	ghostdagDataStore := ghostdagdatastore.New(ghostdagDataCacheSize, preallocateCaches)

	headersSelectedTipStore := headersselectedtipstore.New()
	finalityStore := finalitystore.New(200, preallocateCaches)
	headersSelectedChainStore := headersselectedchainstore.New(pruningWindowSizeForCaches, preallocateCaches)
	daaBlocksStore := daablocksstore.New(pruningWindowSizeForCaches, int(config.FinalityDepth()), preallocateCaches)

	// Processes
	reachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanager.New(
		dbManager,
		reachabilityManager,
		blockRelationStore,
		ghostdagDataStore)
	ghostdagManager := f.ghostdagConstructor(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		blockHeaderStore,
		config.K)
	dagTraversalManager := dagtraversalmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		reachabilityDataStore,
		ghostdagManager,
		consensusStateStore,
		config.GenesisHash)
	pastMedianTimeManager := f.pastMedianTimeConsructor(
		config.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore,
		ghostdagDataStore,
		config.GenesisHash)
	transactionValidator := transactionvalidator.New(config.BlockCoinbaseMaturity,
		config.EnableNonNativeSubnetworks,
		config.MassPerTxByte,
		config.MassPerScriptPubKeyByte,
		config.MassPerSigOp,
		config.MaxCoinbasePayloadLength,
		dbManager,
		pastMedianTimeManager,
		ghostdagDataStore,
		daaBlocksStore)
	difficultyManager := f.difficultyConstructor(
		dbManager,
		ghostdagManager,
		ghostdagDataStore,
		blockHeaderStore,
		daaBlocksStore,
		dagTopologyManager,
		dagTraversalManager,
		config.PowMax,
		config.DifficultyAdjustmentWindowSize,
		config.DisableDifficultyAdjustment,
		config.TargetTimePerBlock,
		config.GenesisHash)
	coinbaseManager := coinbasemanager.New(
		dbManager,
		config.SubsidyReductionInterval,
		config.BaseSubsidy,
		config.CoinbasePayloadScriptPublicKeyMaxLength,
		ghostdagDataStore,
		acceptanceDataStore,
		daaBlocksStore)
	headerTipsManager := headersselectedtipmanager.New(dbManager, dagTopologyManager, dagTraversalManager,
		ghostdagManager, headersSelectedTipStore, headersSelectedChainStore)
	genesisHash := config.GenesisHash
	finalityManager := finalitymanager.New(
		dbManager,
		dagTopologyManager,
		finalityStore,
		ghostdagDataStore,
		genesisHash,
		config.FinalityDepth())
	mergeDepthManager := mergedepthmanager.New(
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		finalityManager,
		ghostdagDataStore)
	blockValidator := blockvalidator.New(
		config.PowMax,
		config.SkipProofOfWork,
		genesisHash,
		config.EnableNonNativeSubnetworks,
		config.MaxBlockSize,
		config.MergeSetSizeLimit,
		config.MaxBlockParents,
		config.TimestampDeviationTolerance,
		config.TargetTimePerBlock,

		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		transactionValidator,
		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		coinbaseManager,
		mergeDepthManager,
		reachabilityManager,

		pruningStore,
		blockStore,
		ghostdagDataStore,
		blockHeaderStore,
		blockStatusStore,
		reachabilityDataStore,
		consensusStateStore,
	)
	consensusStateManager, err := consensusstatemanager.New(
		dbManager,
		config.PruningDepth(),
		config.MaxMassAcceptedByBlock,
		config.MaxBlockParents,
		config.MergeSetSizeLimit,
		genesisHash,

		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		pastMedianTimeManager,
		transactionValidator,
		blockValidator,
		reachabilityManager,
		coinbaseManager,
		mergeDepthManager,
		finalityManager,
		difficultyManager,

		blockStatusStore,
		ghostdagDataStore,
		consensusStateStore,
		multisetStore,
		blockStore,
		utxoDiffStore,
		blockRelationStore,
		acceptanceDataStore,
		blockHeaderStore,
		headersSelectedTipStore,
		pruningStore,
		daaBlocksStore)
	if err != nil {
		return nil, err
	}

	pruningManager := pruningmanager.New(
		dbManager,
		dagTraversalManager,
		dagTopologyManager,
		consensusStateManager,
		consensusStateStore,
		ghostdagDataStore,
		pruningStore,
		blockStatusStore,
		headersSelectedTipStore,
		multisetStore,
		acceptanceDataStore,
		blockStore,
		blockHeaderStore,
		utxoDiffStore,
		daaBlocksStore,
		config.IsArchival,
		genesisHash,
		config.FinalityDepth(),
		config.PruningDepth(),
		config.EnableSanityCheckPruningUTXOSet)

	syncManager := syncmanager.New(
		dbManager,
		genesisHash,
		dagTraversalManager,
		dagTopologyManager,
		ghostdagManager,
		pruningManager,

		ghostdagDataStore,
		blockStatusStore,
		blockHeaderStore,
		blockStore,
		pruningStore,
		headersSelectedChainStore)

	blockBuilder := blockbuilder.New(
		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		coinbaseManager,
		consensusStateManager,
		ghostdagManager,
		transactionValidator,

		acceptanceDataStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
	)

	blockProcessor := blockprocessor.New(
		genesisHash,
		config.TargetTimePerBlock,
		dbManager,
		consensusStateManager,
		pruningManager,
		blockValidator,
		dagTopologyManager,
		reachabilityManager,
		difficultyManager,
		pastMedianTimeManager,
		ghostdagManager,
		coinbaseManager,
		headerTipsManager,
		syncManager,

		acceptanceDataStore,
		blockStore,
		blockStatusStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
		consensusStateStore,
		pruningStore,
		reachabilityDataStore,
		utxoDiffStore,
		blockHeaderStore,
		headersSelectedTipStore,
		finalityStore,
		headersSelectedChainStore,
		daaBlocksStore)

	c := &consensus{
		lock:            &sync.Mutex{},
		databaseContext: dbManager,

		blockProcessor:        blockProcessor,
		blockBuilder:          blockBuilder,
		consensusStateManager: consensusStateManager,
		transactionValidator:  transactionValidator,
		syncManager:           syncManager,
		pastMedianTimeManager: pastMedianTimeManager,
		blockValidator:        blockValidator,
		coinbaseManager:       coinbaseManager,
		dagTopologyManager:    dagTopologyManager,
		dagTraversalManager:   dagTraversalManager,
		difficultyManager:     difficultyManager,
		ghostdagManager:       ghostdagManager,
		headerTipsManager:     headerTipsManager,
		mergeDepthManager:     mergeDepthManager,
		pruningManager:        pruningManager,
		reachabilityManager:   reachabilityManager,
		finalityManager:       finalityManager,

		acceptanceDataStore:       acceptanceDataStore,
		blockStore:                blockStore,
		blockHeaderStore:          blockHeaderStore,
		pruningStore:              pruningStore,
		ghostdagDataStore:         ghostdagDataStore,
		blockStatusStore:          blockStatusStore,
		blockRelationStore:        blockRelationStore,
		consensusStateStore:       consensusStateStore,
		headersSelectedTipStore:   headersSelectedTipStore,
		multisetStore:             multisetStore,
		reachabilityDataStore:     reachabilityDataStore,
		utxoDiffStore:             utxoDiffStore,
		finalityStore:             finalityStore,
		headersSelectedChainStore: headersSelectedChainStore,
		daaBlocksStore:            daaBlocksStore,
	}

	genesisInfo, err := c.GetBlockInfo(genesisHash)
	if err != nil {
		return nil, err
	}

	if !genesisInfo.Exists {
		if c.blockStore.Count(model.NewStagingArea()) > 0 {
			return nil, errors.Errorf(
				"expected genesis block %s is not found in the non-empty DAG \"%s\": wrong config or appdir?",
				genesisHash, config.Params.Name)
		}
		_, err = c.ValidateAndInsertBlock(config.GenesisBlock)
		if err != nil {
			return nil, err
		}
	}

	err = consensusStateManager.RecoverUTXOIfRequired()
	if err != nil {
		return nil, err
	}
	err = pruningManager.ClearImportedPruningPointData()
	if err != nil {
		return nil, err
	}
	err = pruningManager.UpdatePruningPointIfRequired()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (f *factory) NewTestConsensus(config *Config, testName string) (
	tc testapi.TestConsensus, teardown func(keepDataDir bool), err error) {
	datadir := f.dataDir
	if datadir == "" {
		datadir, err = ioutil.TempDir("", testName)
		if err != nil {
			return nil, nil, err
		}
	}
	var cacheSizeMiB int
	if f.cacheSizeMiB != nil {
		cacheSizeMiB = *f.cacheSizeMiB
	} else {
		cacheSizeMiB = defaultTestLeveldbCacheSizeMiB
	}
	if f.preallocateCaches == nil {
		f.SetTestPreAllocateCache(defaultTestPreallocateCaches)
	}
	db, err := ldb.NewLevelDB(datadir, cacheSizeMiB)
	if err != nil {
		return nil, nil, err
	}
	consensusAsInterface, err := f.NewConsensus(config, db)
	if err != nil {
		return nil, nil, err
	}

	consensusAsImplementation := consensusAsInterface.(*consensus)
	testConsensusStateManager := consensusstatemanager.NewTestConsensusStateManager(consensusAsImplementation.consensusStateManager)
	testTransactionValidator := transactionvalidator.NewTestTransactionValidator(consensusAsImplementation.transactionValidator)

	tstConsensus := &testConsensus{
		dagParams:                 &config.Params,
		consensus:                 consensusAsImplementation,
		database:                  db,
		testConsensusStateManager: testConsensusStateManager,
		testReachabilityManager: reachabilitymanager.NewTestReachabilityManager(consensusAsImplementation.
			reachabilityManager),
		testTransactionValidator: testTransactionValidator,
	}
	tstConsensus.testBlockBuilder = blockbuilder.NewTestBlockBuilder(consensusAsImplementation.blockBuilder, tstConsensus)
	teardown = func(keepDataDir bool) {
		db.Close()
		if !keepDataDir {
			err := os.RemoveAll(f.dataDir)
			if err != nil {
				log.Errorf("Error removing data directory for test consensus: %s", err)
			}
		}
	}
	return tstConsensus, teardown, nil
}

func (f *factory) SetTestDataDir(dataDir string) {
	f.dataDir = dataDir
}

func (f *factory) SetTestGHOSTDAGManager(ghostdagConstructor GHOSTDAGManagerConstructor) {
	f.ghostdagConstructor = ghostdagConstructor
}

func (f *factory) SetTestPastMedianTimeManager(medianTimeConstructor PastMedianTimeManagerConstructor) {
	f.pastMedianTimeConsructor = medianTimeConstructor
}

// SetTestDifficultyManager is a setter for the difficultyManager field on the factory.
func (f *factory) SetTestDifficultyManager(difficultyConstructor DifficultyManagerConstructor) {
	f.difficultyConstructor = difficultyConstructor
}

func (f *factory) SetTestLevelDBCacheSize(cacheSizeMiB int) {
	f.cacheSizeMiB = &cacheSizeMiB
}
func (f *factory) SetTestPreAllocateCache(preallocateCaches bool) {
	f.preallocateCaches = &preallocateCaches
}
