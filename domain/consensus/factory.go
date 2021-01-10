package consensus

import (
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headersselectedchainstore"
	"io/ioutil"
	"os"
	"sync"

	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/finalitymanager"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/finalitystore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
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
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
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

// Factory instantiates new Consensuses
type Factory interface {
	NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (externalapi.Consensus, error)
	NewTestConsensus(dagParams *dagconfig.Params, testName string) (
		tc testapi.TestConsensus, teardown func(keepDataDir bool), err error)
	NewTestConsensusWithDataDir(dagParams *dagconfig.Params, dataDir string) (
		tc testapi.TestConsensus, teardown func(keepDataDir bool), err error)
}

type factory struct{}

// NewFactory creates a new Consensus factory
func NewFactory() Factory {
	return &factory{}
}

// NewConsensus instantiates a new Consensus
func (f *factory) NewConsensus(dagParams *dagconfig.Params, db infrastructuredatabase.Database) (externalapi.Consensus, error) {
	dbManager := consensusdatabase.New(db)

	pruningWindowSizeForCaches := int(dagParams.PruningDepth())

	// Data Structures
	acceptanceDataStore := acceptancedatastore.New(200)
	blockStore, err := blockstore.New(dbManager, 200)
	if err != nil {
		return nil, err
	}
	blockHeaderStore, err := blockheaderstore.New(dbManager, 10_000)
	if err != nil {
		return nil, err
	}
	blockRelationStore := blockrelationstore.New(pruningWindowSizeForCaches)
	blockStatusStore := blockstatusstore.New(200)
	multisetStore := multisetstore.New(200)
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New(pruningWindowSizeForCaches)
	utxoDiffStore := utxodiffstore.New(200)
	consensusStateStore := consensusstatestore.New(10_000)
	ghostdagDataStore := ghostdagdatastore.New(pruningWindowSizeForCaches)
	headersSelectedTipStore := headersselectedtipstore.New()
	finalityStore := finalitystore.New(200)
	headersSelectedChainStore := headersselectedchainstore.New(pruningWindowSizeForCaches)

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
	ghostdagManager := ghostdagmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		blockHeaderStore,
		dagParams.K)
	dagTraversalManager := dagtraversalmanager.New(
		dbManager,
		dagTopologyManager,
		ghostdagDataStore,
		reachabilityDataStore,
		ghostdagManager)
	pastMedianTimeManager := pastmediantimemanager.New(
		dagParams.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore,
		ghostdagDataStore)
	transactionValidator := transactionvalidator.New(dagParams.BlockCoinbaseMaturity,
		dagParams.EnableNonNativeSubnetworks,
		dagParams.MassPerTxByte,
		dagParams.MassPerScriptPubKeyByte,
		dagParams.MassPerSigOp,
		dagParams.MaxCoinbasePayloadLength,
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
		dagParams.DisableDifficultyAdjustment,
		dagParams.TargetTimePerBlock,
		dagParams.GenesisHash)
	coinbaseManager := coinbasemanager.New(
		dbManager,
		dagParams.SubsidyReductionInterval,
		dagParams.BaseSubsidy,
		dagParams.CoinbasePayloadScriptPublicKeyMaxLength,
		ghostdagDataStore,
		acceptanceDataStore)
	headerTipsManager := headersselectedtipmanager.New(dbManager, dagTopologyManager, dagTraversalManager,
		ghostdagManager, headersSelectedTipStore, headersSelectedChainStore)
	genesisHash := dagParams.GenesisHash
	finalityManager := finalitymanager.New(
		dbManager,
		dagTopologyManager,
		finalityStore,
		ghostdagDataStore,
		genesisHash,
		dagParams.FinalityDepth())
	mergeDepthManager := mergedepthmanager.New(
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		finalityManager,
		ghostdagDataStore)
	blockValidator := blockvalidator.New(
		dagParams.PowMax,
		dagParams.SkipProofOfWork,
		genesisHash,
		dagParams.EnableNonNativeSubnetworks,
		dagParams.MaxBlockSize,
		dagParams.MergeSetSizeLimit,
		dagParams.MaxBlockParents,
		dagParams.TimestampDeviationTolerance,
		dagParams.TargetTimePerBlock,

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
	)
	consensusStateManager, err := consensusstatemanager.New(
		dbManager,
		dagParams.PruningDepth(),
		dagParams.MaxMassAcceptedByBlock,
		dagParams.MaxBlockParents,
		dagParams.MergeSetSizeLimit,
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
		pruningStore)
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
		genesisHash,
		dagParams.FinalityDepth(),
		dagParams.PruningDepth())

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
		acceptanceDataStore,
		blockRelationStore,
		multisetStore,
		ghostdagDataStore,
	)

	blockProcessor := blockprocessor.New(
		genesisHash,
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
		headersSelectedChainStore)

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
	}

	genesisInfo, err := c.GetBlockInfo(genesisHash)
	if err != nil {
		return nil, err
	}

	if !genesisInfo.Exists {
		_, err = c.ValidateAndInsertBlock(dagParams.GenesisBlock)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (f *factory) NewTestConsensus(dagParams *dagconfig.Params, testName string) (
	tc testapi.TestConsensus, teardown func(keepDataDir bool), err error) {

	dataDir, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}

	return f.NewTestConsensusWithDataDir(dagParams, dataDir)
}

func (f *factory) NewTestConsensusWithDataDir(dagParams *dagconfig.Params, dataDir string) (
	tc testapi.TestConsensus, teardown func(keepDataDir bool), err error) {

	db, err := ldb.NewLevelDB(dataDir)
	if err != nil {
		return nil, nil, err
	}
	consensusAsInterface, err := f.NewConsensus(dagParams, db)
	if err != nil {
		return nil, nil, err
	}

	consensusAsImplementation := consensusAsInterface.(*consensus)

	testConsensusStateManager := consensusstatemanager.NewTestConsensusStateManager(consensusAsImplementation.consensusStateManager)

	testTransactionValidator := transactionvalidator.NewTestTransactionValidator(consensusAsImplementation.transactionValidator)

	tstConsensus := &testConsensus{
		dagParams:                 dagParams,
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
			err := os.RemoveAll(dataDir)
			if err != nil {
				log.Errorf("Error removing data directory for test consensus: %s", err)
			}
		}
	}

	return tstConsensus, teardown, nil
}
