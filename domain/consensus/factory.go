package consensus

import (
	"io/ioutil"
	"os"
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"

	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headertipsstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/multisetstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/pruningstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/utxodiffstore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockbuilder"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/headertipsmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/mergedepthmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/syncmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
)

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (externalapi.Consensus, error)
	NewTestConsensus(dagParams *dagconfig.Params, testName string) (tc testapi.TestConsensus, teardown func(), err error)
}

type factory struct{}

// NewFactory creates a new Consensus factory
func NewFactory() Factory {
	return &factory{}
}

// NewConsensus instantiates a new Consensus
func (f *factory) NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (externalapi.Consensus, error) {
	dbManager := consensusdatabase.New(db)

	// Data Structures
	storeCacheSize := int(dagParams.FinalityDepth())
	acceptanceDataStore := acceptancedatastore.New(storeCacheSize)
	blockStore, err := blockstore.New(dbManager, storeCacheSize)
	if err != nil {
		return nil, err
	}
	blockHeaderStore, err := blockheaderstore.New(dbManager, storeCacheSize)
	if err != nil {
		return nil, err
	}
	blockRelationStore := blockrelationstore.New(storeCacheSize)
	blockStatusStore := blockstatusstore.New(storeCacheSize)
	multisetStore := multisetstore.New(storeCacheSize)
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New(storeCacheSize)
	utxoDiffStore := utxodiffstore.New(storeCacheSize)
	consensusStateStore := consensusstatestore.New(storeCacheSize)
	ghostdagDataStore := ghostdagdatastore.New(storeCacheSize)
	headerTipsStore := headertipsstore.New()

	// Processes
	reachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStore,
		reachabilityDataStore)
	dagTopologyManager := dagtopologymanager.New(
		dbManager,
		reachabilityManager,
		blockRelationStore)
	ghostdagManager := ghostdagmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		model.KType(dagParams.K))
	dagTraversalManager := dagtraversalmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		ghostdagManager)
	pastMedianTimeManager := pastmediantimemanager.New(
		dagParams.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore,
		ghostdagDataStore)
	transactionValidator := transactionvalidator.New(dagParams.BlockCoinbaseMaturity,
		dagParams.EnableNonNativeSubnetworks,
		dbManager,
		pastMedianTimeManager,
		ghostdagDataStore)
	difficultyManager := difficultymanager.New(
		dbManager,
		ghostdagManager,
		ghostdagDataStore,
		blockHeaderStore,
		dagTopologyManager,
		dagTraversalManager,
		dagParams.PowMax,
		dagParams.DifficultyAdjustmentWindowSize,
		dagParams.TargetTimePerBlock)
	coinbaseManager := coinbasemanager.New(
		dbManager,
		ghostdagDataStore,
		acceptanceDataStore)
	headerTipsManager := headertipsmanager.New(dbManager, dagTopologyManager, ghostdagManager, headerTipsStore)
	genesisHash := dagParams.GenesisHash
	mergeDepthManager := mergedepthmanager.New(
		dagParams.FinalityDepth(),
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		ghostdagDataStore)
	blockValidator := blockvalidator.New(
		dagParams.PowMax,
		dagParams.SkipProofOfWork,
		genesisHash,
		dagParams.EnableNonNativeSubnetworks,
		dagParams.DisableDifficultyAdjustment,
		dagParams.DifficultyAdjustmentWindowSize,

		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		transactionValidator,
		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,
		coinbaseManager,
		mergeDepthManager,

		blockStore,
		ghostdagDataStore,
		blockHeaderStore,
		blockStatusStore,
	)
	consensusStateManager, err := consensusstatemanager.New(
		dbManager,
		dagParams.FinalityDepth(),
		dagParams.PruningDepth(),
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

		blockStatusStore,
		ghostdagDataStore,
		consensusStateStore,
		multisetStore,
		blockStore,
		utxoDiffStore,
		blockRelationStore,
		acceptanceDataStore,
		blockHeaderStore,
		headerTipsStore)
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
		multisetStore,
		acceptanceDataStore,
		blockStore,
		utxoDiffStore,
		genesisHash,
		dagParams.FinalityDepth(),
		dagParams.PruningDepth())

	syncManager := syncmanager.New(
		dbManager,
		genesisHash,
		dagParams.TargetTimePerBlock.Milliseconds(),
		dagTraversalManager,
		dagTopologyManager,
		ghostdagManager,
		consensusStateManager,

		ghostdagDataStore,
		blockStatusStore,
		blockHeaderStore,
		headerTipsStore,
		blockStore)

	blockBuilder := blockbuilder.New(
		dbManager,
		difficultyManager,
		pastMedianTimeManager,
		coinbaseManager,
		consensusStateManager,
		ghostdagManager,
		acceptanceDataStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
	)

	blockProcessor := blockprocessor.New(
		dagParams,
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
		headerTipsStore)

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

		acceptanceDataStore:   acceptanceDataStore,
		blockStore:            blockStore,
		blockHeaderStore:      blockHeaderStore,
		pruningStore:          pruningStore,
		ghostdagDataStore:     ghostdagDataStore,
		blockStatusStore:      blockStatusStore,
		blockRelationStore:    blockRelationStore,
		consensusStateStore:   consensusStateStore,
		headerTipsStore:       headerTipsStore,
		multisetStore:         multisetStore,
		reachabilityDataStore: reachabilityDataStore,
		utxoDiffStore:         utxoDiffStore,
	}

	genesisInfo, err := c.GetBlockInfo(genesisHash)
	if err != nil {
		return nil, err
	}

	if !genesisInfo.Exists {
		err = c.ValidateAndInsertBlock(dagParams.GenesisBlock)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (f *factory) NewTestConsensus(dagParams *dagconfig.Params, testName string) (
	tc testapi.TestConsensus, teardown func(), err error) {

	testDatabaseDir, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}
	db, err := ldb.NewLevelDB(testDatabaseDir)
	if err != nil {
		return nil, nil, err
	}
	consensusAsInterface, err := f.NewConsensus(dagParams, db)
	if err != nil {
		return nil, nil, err
	}

	consensusAsImplementation := consensusAsInterface.(*consensus)

	testConsensusStateManager := consensusstatemanager.NewTestConsensusStateManager(consensusAsImplementation.consensusStateManager)

	tstConsensus := &testConsensus{
		consensus:                 consensusAsImplementation,
		testConsensusStateManager: testConsensusStateManager,
		testReachabilityManager: reachabilitymanager.NewTestReachabilityManager(consensusAsImplementation.
			reachabilityManager),
	}
	tstConsensus.testBlockBuilder = blockbuilder.NewTestBlockBuilder(consensusAsImplementation.blockBuilder, tstConsensus)
	teardown = func() {
		db.Close()
		os.RemoveAll(testDatabaseDir)
	}

	return tstConsensus, teardown, nil
}
