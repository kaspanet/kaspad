package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockwindowheapslicestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/daawindowstore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockparentbuilder"
	parentssanager "github.com/kaspanet/kaspad/domain/consensus/processes/parentsmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningproofmanager"
	"github.com/kaspanet/kaspad/util/staging"
	"io/ioutil"
	"os"
	"sync"

	"github.com/kaspanet/kaspad/domain/prefixmanager/prefix"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/daablocksstore"
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

	SkipAddingGenesis bool
}

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(config *Config, db infrastructuredatabase.Database, dbPrefix *prefix.Prefix) (
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
func (f *factory) NewConsensus(config *Config, db infrastructuredatabase.Database, dbPrefix *prefix.Prefix) (
	externalapi.Consensus, error) {

	dbManager := consensusdatabase.New(db)
	prefixBucket := consensusdatabase.MakeBucket(dbPrefix.Serialize())

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
	daaWindowStore := daawindowstore.New(prefixBucket, 10_000, preallocateCaches)
	acceptanceDataStore := acceptancedatastore.New(prefixBucket, 200, preallocateCaches)
	blockStore, err := blockstore.New(dbManager, prefixBucket, 200, preallocateCaches)
	if err != nil {
		return nil, err
	}
	blockHeaderStore, err := blockheaderstore.New(dbManager, prefixBucket, 10_000, preallocateCaches)
	if err != nil {
		return nil, err
	}

	blockStatusStore := blockstatusstore.New(prefixBucket, pruningWindowSizePlusFinalityDepthForCache, preallocateCaches)
	multisetStore := multisetstore.New(prefixBucket, 200, preallocateCaches)
	pruningStore := pruningstore.New(prefixBucket, 2, preallocateCaches)
	utxoDiffStore := utxodiffstore.New(prefixBucket, 200, preallocateCaches)
	consensusStateStore := consensusstatestore.New(prefixBucket, 10_000, preallocateCaches)

	headersSelectedTipStore := headersselectedtipstore.New(prefixBucket)
	finalityStore := finalitystore.New(prefixBucket, 200, preallocateCaches)
	headersSelectedChainStore := headersselectedchainstore.New(prefixBucket, pruningWindowSizeForCaches, preallocateCaches)
	daaBlocksStore := daablocksstore.New(prefixBucket, pruningWindowSizeForCaches, int(config.FinalityDepth()), preallocateCaches)
	windowHeapSliceStore := blockwindowheapslicestore.New(2000, preallocateCaches)

	newReachabilityDataStore := reachabilitydatastore.New(prefixBucket, pruningWindowSizePlusFinalityDepthForCache*2, preallocateCaches)
	blockRelationStores, reachabilityDataStores, ghostdagDataStores := dagStores(config, prefixBucket, pruningWindowSizePlusFinalityDepthForCache, pruningWindowSizeForCaches, preallocateCaches)
	oldReachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStores[0],
		reachabilityDataStores[0])
	isOldReachabilityInitialized, err := reachabilityDataStores[0].HasReachabilityData(dbManager, model.NewStagingArea(), model.VirtualGenesisBlockHash)
	if err != nil {
		return nil, err
	}

	newReachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStores[0],
		newReachabilityDataStore)
	reachabilityManager := newReachabilityManager
	if isOldReachabilityInitialized {
		reachabilityManager = oldReachabilityManager
	} else {
		for i := range reachabilityDataStores {
			reachabilityDataStores[i] = newReachabilityDataStore
		}
	}
	reachabilityDataStore := reachabilityDataStores[0]

	dagTopologyManagers, ghostdagManagers, dagTraversalManagers := f.dagProcesses(config, dbManager, blockHeaderStore, daaWindowStore, windowHeapSliceStore, blockRelationStores, reachabilityDataStores, ghostdagDataStores, isOldReachabilityInitialized)

	blockRelationStore := blockRelationStores[0]

	ghostdagDataStore := ghostdagDataStores[0]

	dagTopologyManager := dagTopologyManagers[0]
	ghostdagManager := ghostdagManagers[0]
	dagTraversalManager := dagTraversalManagers[0]

	// Processes
	parentsManager := parentssanager.New(config.GenesisHash, config.MaxBlockLevel)
	blockParentBuilder := blockparentbuilder.New(
		dbManager,
		blockHeaderStore,
		dagTopologyManager,
		parentsManager,
		reachabilityDataStore,
		pruningStore,

		config.GenesisHash,
		config.MaxBlockLevel,
	)
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
		config.GenesisHash,
		config.GenesisBlock.Header.Bits())
	coinbaseManager := coinbasemanager.New(
		dbManager,
		config.SubsidyGenesisReward,
		config.PreDeflationaryPhaseBaseSubsidy,
		config.CoinbasePayloadScriptPublicKeyMaxLength,
		config.GenesisHash,
		config.DeflationaryPhaseDaaScore,
		config.DeflationaryPhaseBaseSubsidy,
		dagTraversalManager,
		ghostdagDataStore,
		acceptanceDataStore,
		daaBlocksStore,
		blockStore,
		pruningStore,
		blockHeaderStore)
	headerTipsManager := headersselectedtipmanager.New(dbManager, dagTopologyManager, dagTraversalManager,
		ghostdagManager, headersSelectedTipStore, headersSelectedChainStore)
	genesisHash := config.GenesisHash
	finalityManager := finalitymanager.New(
		dbManager,
		dagTopologyManager,
		finalityStore,
		ghostdagDataStore,
		pruningStore,
		genesisHash,
		config.FinalityDepth())
	mergeDepthManager := mergedepthmanager.New(
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		finalityManager,
		ghostdagDataStore)
	consensusStateManager, err := consensusstatemanager.New(
		dbManager,
		config.MaxBlockParents,
		config.MergeSetSizeLimit,
		genesisHash,

		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		pastMedianTimeManager,
		transactionValidator,
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
		finalityManager,

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
		reachabilityDataStore,
		daaWindowStore,

		config.IsArchival,
		genesisHash,
		config.FinalityDepth(),
		config.PruningDepth(),
		config.EnableSanityCheckPruningUTXOSet,
		config.K,
		config.DifficultyAdjustmentWindowSize,
	)

	blockValidator := blockvalidator.New(
		config.PowMax,
		config.SkipProofOfWork,
		genesisHash,
		config.EnableNonNativeSubnetworks,
		config.MaxBlockMass,
		config.MergeSetSizeLimit,
		config.MaxBlockParents,
		config.TimestampDeviationTolerance,
		config.TargetTimePerBlock,
		config.IgnoreHeaderMass,
		config.MaxBlockLevel,

		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		transactionValidator,
		ghostdagManagers,
		dagTopologyManagers,
		dagTraversalManager,
		coinbaseManager,
		mergeDepthManager,
		reachabilityManager,
		finalityManager,
		blockParentBuilder,
		pruningManager,
		parentsManager,

		pruningStore,
		blockStore,
		ghostdagDataStores,
		blockHeaderStore,
		blockStatusStore,
		reachabilityDataStore,
		consensusStateStore,
		daaBlocksStore,
	)

	syncManager := syncmanager.New(
		dbManager,
		genesisHash,
		config.MergeSetSizeLimit,
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
		genesisHash,

		difficultyManager,
		pastMedianTimeManager,
		coinbaseManager,
		consensusStateManager,
		ghostdagManager,
		transactionValidator,
		finalityManager,
		blockParentBuilder,
		pruningManager,

		acceptanceDataStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
		daaBlocksStore,
	)

	blockProcessor := blockprocessor.New(
		genesisHash,
		config.TargetTimePerBlock,
		config.MaxBlockLevel,
		dbManager,
		consensusStateManager,
		pruningManager,
		blockValidator,
		dagTopologyManager,
		reachabilityManager,
		difficultyManager,
		pastMedianTimeManager,
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
		daaBlocksStore,
		daaWindowStore)

	pruningProofManager := pruningproofmanager.New(
		dbManager,
		dagTopologyManagers,
		ghostdagManagers,
		reachabilityManager,
		dagTraversalManagers,
		parentsManager,
		pruningManager,

		ghostdagDataStores,
		pruningStore,
		blockHeaderStore,
		blockStatusStore,
		finalityStore,
		consensusStateStore,
		blockRelationStore,
		reachabilityDataStore,

		genesisHash,
		config.K,
		config.PruningProofM,
		config.MaxBlockLevel,
	)

	c := &consensus{
		lock:            &sync.Mutex{},
		databaseContext: dbManager,

		genesisBlock: config.GenesisBlock,
		genesisHash:  config.GenesisHash,

		blockProcessor:        blockProcessor,
		blockBuilder:          blockBuilder,
		consensusStateManager: consensusStateManager,
		transactionValidator:  transactionValidator,
		syncManager:           syncManager,
		pastMedianTimeManager: pastMedianTimeManager,
		blockValidator:        blockValidator,
		coinbaseManager:       coinbaseManager,
		dagTopologyManagers:   dagTopologyManagers,
		dagTraversalManager:   dagTraversalManager,
		difficultyManager:     difficultyManager,
		ghostdagManagers:      ghostdagManagers,
		headerTipsManager:     headerTipsManager,
		mergeDepthManager:     mergeDepthManager,
		pruningManager:        pruningManager,
		reachabilityManager:   reachabilityManager,
		finalityManager:       finalityManager,
		pruningProofManager:   pruningProofManager,

		acceptanceDataStore:                 acceptanceDataStore,
		blockStore:                          blockStore,
		blockHeaderStore:                    blockHeaderStore,
		pruningStore:                        pruningStore,
		ghostdagDataStores:                  ghostdagDataStores,
		blockStatusStore:                    blockStatusStore,
		blockRelationStores:                 blockRelationStores,
		consensusStateStore:                 consensusStateStore,
		headersSelectedTipStore:             headersSelectedTipStore,
		multisetStore:                       multisetStore,
		reachabilityDataStore:               reachabilityDataStore,
		utxoDiffStore:                       utxoDiffStore,
		finalityStore:                       finalityStore,
		headersSelectedChainStore:           headersSelectedChainStore,
		daaBlocksStore:                      daaBlocksStore,
		blocksWithTrustedDataDAAWindowStore: daaWindowStore,
	}

	if isOldReachabilityInitialized {
		log.Infof("Migrating the database to the new reachability structure")
		stagingArea := model.NewStagingArea()
		dbTx, err := dbManager.Begin()
		if err != nil {
			return nil, err
		}

		log.Infof("Deleting the new reachability store (in case a previous migration failed)")
		err = newReachabilityDataStore.Delete(dbTx)
		if err != nil {
			return nil, err
		}

		err = dbTx.Commit()
		if err != nil {
			return nil, err
		}

		log.Infof("Initializing new reachability store")
		err = newReachabilityManager.Init(stagingArea)
		if err != nil {
			return nil, err
		}

		log.Infof("Committing changes")
		err = staging.CommitAllChanges(dbManager, stagingArea)
		if err != nil {
			return nil, err
		}

		err = c.pruningProofManager.RebuildReachability(newReachabilityDataStore)
		if err != nil {
			return nil, err
		}

		// Because we use reachabilityDataStores[0] for the migration indication, this is the only store we
		// need to delete in an atomic way. For the rest of the stores we don't need database transactions,
		// so we can delete them directly, hence saving memory.
		log.Infof("Deleting the old level 0 reachability store")
		dbTx, err = dbManager.Begin()
		if err != nil {
			return nil, err
		}

		err = reachabilityDataStores[0].Delete(dbTx)
		if err != nil {
			return nil, err
		}

		err = dbTx.Commit()
		if err != nil {
			return nil, err
		}

		for i, store := range reachabilityDataStores[1:] {
			log.Infof("Deleting the old level %d reachability store", i+1)
			err = store.Delete(dbManager)
			if err != nil {
				return nil, err
			}
		}

		log.Infof("Restarting the consensus")
		return f.NewConsensus(config, db, dbPrefix)
	}

	err = c.Init(config.SkipAddingGenesis)
	if err != nil {
		return nil, err
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

	testConsensusDBPrefix := &prefix.Prefix{}
	consensusAsInterface, err := f.NewConsensus(config, db, testConsensusDBPrefix)
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

func dagStores(config *Config,
	prefixBucket model.DBBucket,
	pruningWindowSizePlusFinalityDepthForCache, pruningWindowSizeForCaches int,
	preallocateCaches bool) ([]model.BlockRelationStore, []model.ReachabilityDataStore, []model.GHOSTDAGDataStore) {

	blockRelationStores := make([]model.BlockRelationStore, config.MaxBlockLevel+1)
	reachabilityDataStores := make([]model.ReachabilityDataStore, config.MaxBlockLevel+1)
	ghostdagDataStores := make([]model.GHOSTDAGDataStore, config.MaxBlockLevel+1)

	ghostdagDataCacheSize := pruningWindowSizeForCaches * 2
	if ghostdagDataCacheSize < config.DifficultyAdjustmentWindowSize {
		ghostdagDataCacheSize = config.DifficultyAdjustmentWindowSize
	}

	for i := 0; i <= config.MaxBlockLevel; i++ {
		prefixBucket := prefixBucket.Bucket([]byte{byte(i)})
		if i == 0 {
			blockRelationStores[i] = blockrelationstore.New(prefixBucket, pruningWindowSizePlusFinalityDepthForCache, preallocateCaches)
			reachabilityDataStores[i] = reachabilitydatastore.New(prefixBucket, pruningWindowSizePlusFinalityDepthForCache*2, preallocateCaches)
			ghostdagDataStores[i] = ghostdagdatastore.New(prefixBucket, ghostdagDataCacheSize, preallocateCaches)
		} else {
			blockRelationStores[i] = blockrelationstore.New(prefixBucket, 200, false)
			reachabilityDataStores[i] = reachabilitydatastore.New(prefixBucket, pruningWindowSizePlusFinalityDepthForCache, false)
			ghostdagDataStores[i] = ghostdagdatastore.New(prefixBucket, 200, false)
		}
	}

	return blockRelationStores, reachabilityDataStores, ghostdagDataStores
}

func (f *factory) dagProcesses(config *Config,
	dbManager model.DBManager,
	blockHeaderStore model.BlockHeaderStore,
	daaWindowStore model.BlocksWithTrustedDataDAAWindowStore,
	windowHeapSliceStore model.WindowHeapSliceStore,
	blockRelationStores []model.BlockRelationStore,
	reachabilityDataStores []model.ReachabilityDataStore,
	ghostdagDataStores []model.GHOSTDAGDataStore,
	isOldReachabilityInitialized bool) (
	[]model.DAGTopologyManager,
	[]model.GHOSTDAGManager,
	[]model.DAGTraversalManager,
) {

	reachabilityManagers := make([]model.ReachabilityManager, config.MaxBlockLevel+1)
	dagTopologyManagers := make([]model.DAGTopologyManager, config.MaxBlockLevel+1)
	ghostdagManagers := make([]model.GHOSTDAGManager, config.MaxBlockLevel+1)
	dagTraversalManagers := make([]model.DAGTraversalManager, config.MaxBlockLevel+1)

	newReachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStores[0],
		reachabilityDataStores[0])

	for i := 0; i <= config.MaxBlockLevel; i++ {
		if isOldReachabilityInitialized {
			reachabilityManagers[i] = reachabilitymanager.New(
				dbManager,
				ghostdagDataStores[i],
				reachabilityDataStores[i])
		} else {
			reachabilityManagers[i] = newReachabilityManager
		}

		dagTopologyManagers[i] = dagtopologymanager.New(
			dbManager,
			reachabilityManagers[i],
			blockRelationStores[i],
			ghostdagDataStores[i])

		ghostdagManagers[i] = f.ghostdagConstructor(
			dbManager,
			dagTopologyManagers[i],
			ghostdagDataStores[i],
			blockHeaderStore,
			config.K,
			config.GenesisHash)

		dagTraversalManagers[i] = dagtraversalmanager.New(
			dbManager,
			dagTopologyManagers[i],
			ghostdagDataStores[i],
			reachabilityManagers[i],
			ghostdagManagers[i],
			daaWindowStore,
			windowHeapSliceStore,
			config.GenesisHash,
			config.DifficultyAdjustmentWindowSize)
	}

	return dagTopologyManagers, ghostdagManagers, dagTraversalManagers
}
