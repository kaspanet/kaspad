package blockprocessor_test

import (
	"io/ioutil"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/go-secp256k1"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/consensusstatestore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/multisetstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/pruningstore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"	
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/utxodiffstore"	
	"github.com/kaspanet/kaspad/domain/consensus/model"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockprocessor"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockvalidator"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"		
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"		
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"	
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/processes/headersselectedtipmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/mergedepthmanager"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/finalitystore"
	"github.com/kaspanet/kaspad/domain/consensus/processes/finalitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/syncmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/blockbuilder"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/headersselectedtipstore"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"

)

type testContext struct {
	GenesisHash				*externalapi.DomainHash
	DbManager				model.DBManager
	PruningManager			model.PruningManager
	BlockValidator			model.BlockValidator
	DagTopologyManager		model.DAGTopologyManager
	ReachabilityManager		model.ReachabilityManager
	DifficultyManager		model.DifficultyManager
	PastMedianTimeManager	model.PastMedianTimeManager
	GhostdagManager			model.GHOSTDAGManager
	CoinbaseManager			model.CoinbaseManager
	HeaderTipsManager		model.HeadersSelectedTipManager
	SyncManager				model.SyncManager
	AcceptanceDataStore		model.AcceptanceDataStore
	BlockStore				model.BlockStore
	BlockStatusStore		model.BlockStatusStore
	BlockRelationStore		model.BlockRelationStore
	MultisetStore			model.MultisetStore
	GhostdagDataStore		model.GHOSTDAGDataStore
	ConsensusStateStore		model.ConsensusStateStore
	PruningStore			model.PruningStore
	ReachabilityDataStore	model.ReachabilityDataStore
	UtxoDiffStore			model.UTXODiffStore
	BlockHeaderStore		model.BlockHeaderStore
	HeadersSelectedTipStore	model.HeaderSelectedTipStore
	FinalityStore			model.FinalityStore
	ConsensusStateManager	model.ConsensusStateManager
	GenesisBlock			*externalapi.DomainBlock
}

func setupTestContext(dagParams *dagconfig.Params, dbManager model.DBManager) (*testContext, error) {
	// Data Structures
	acceptanceDataStore := acceptancedatastore.New(200)
	blockStore, err := blockstore.New(dbManager, 200)
	if err != nil {
		return nil, err
	}
	blockHeaderStore, err := blockheaderstore.New(dbManager, 200)
	if err != nil {
		return nil, err
	}
	blockRelationStore := blockrelationstore.New(200)
	blockStatusStore := blockstatusstore.New(200)
	multisetStore := multisetstore.New(200)
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New(200)
	utxoDiffStore := utxodiffstore.New(200)
	consensusStateStore := consensusstatestore.New()
	ghostdagDataStore := ghostdagdatastore.New(200)
	headersSelectedTipStore := headersselectedtipstore.New()
	finalityStore := finalitystore.New(200)
	
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
	consensusStateManager, _ := consensusstatemanager.New(
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
	headerTipsManager := headersselectedtipmanager.New(dbManager, dagTopologyManager, ghostdagManager, headersSelectedTipStore)
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
		pruningStore)

	return &testContext{
		GenesisHash: genesisHash,
		DbManager: dbManager,
		ConsensusStateManager: consensusStateManager,
		PruningManager: pruningManager,
		BlockValidator: blockValidator,
		DagTopologyManager: dagTopologyManager,
		ReachabilityManager: reachabilityManager,
		DifficultyManager: difficultyManager,
		PastMedianTimeManager: pastMedianTimeManager,
		GhostdagManager: ghostdagManager,
		CoinbaseManager: coinbaseManager,
		HeaderTipsManager: headerTipsManager,
		SyncManager: syncManager,
		AcceptanceDataStore: acceptanceDataStore,
		BlockStore: blockStore,
		BlockStatusStore: blockStatusStore,
		BlockRelationStore: blockRelationStore,
		MultisetStore: multisetStore,
		GhostdagDataStore: ghostdagDataStore,
		ConsensusStateStore: consensusStateStore,
		PruningStore: pruningStore,
		ReachabilityDataStore: reachabilityDataStore,
		UtxoDiffStore: utxoDiffStore,
		BlockHeaderStore: blockHeaderStore,
		HeadersSelectedTipStore: headersSelectedTipStore,
		FinalityStore: finalityStore,
		GenesisBlock: dagParams.GenesisBlock,
	}, nil
}

func SetupBlockBuilder(testContext *testContext) (model.BlockBuilder) {
	blockBuilder := blockbuilder.New(
		testContext.DbManager,
		testContext.DifficultyManager,
		testContext.PastMedianTimeManager,
		testContext.CoinbaseManager,
		testContext.ConsensusStateManager,
		testContext.GhostdagManager,
		testContext.AcceptanceDataStore,
		testContext.BlockRelationStore,
		testContext.MultisetStore,
		testContext.GhostdagDataStore,
	)

	return blockBuilder
}

func SetupBlockProcessor(testContext *testContext) (model.BlockProcessor, error) {
	blockProcessor := blockprocessor.New(
		testContext.GenesisHash,
		testContext.DbManager,
		testContext.ConsensusStateManager,
		testContext.PruningManager,
		testContext.BlockValidator,
		testContext.DagTopologyManager,
		testContext.ReachabilityManager,
		testContext.DifficultyManager,
		testContext.PastMedianTimeManager,
		testContext.GhostdagManager,
		testContext.CoinbaseManager,
		testContext.HeaderTipsManager,
		testContext.SyncManager,
		testContext.AcceptanceDataStore,
		testContext.BlockStore,
		testContext.BlockStatusStore,
		testContext.BlockRelationStore,
		testContext.MultisetStore,
		testContext.GhostdagDataStore,
		testContext.ConsensusStateStore,
		testContext.PruningStore,
		testContext.ReachabilityDataStore,
		testContext.UtxoDiffStore,
		testContext.BlockHeaderStore,
		testContext.HeadersSelectedTipStore,
		testContext.FinalityStore)

		_, err := blockProcessor.ValidateAndInsertBlock(testContext.GenesisBlock)
		if err != nil {
			return nil, err
		}
	return blockProcessor, nil
}


func SetupLDB(testName string) (*ldb.LevelDB, func(), error) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		return nil, nil, err
	}
	ldb, err := ldb.NewLevelDB(path)
	if err != nil {
		return nil, nil, err
	}

	teardownFunc := func() {
		err = ldb.Close()
	}

	return ldb, teardownFunc, nil
}

func SetupDBManager(testName string) (model.DBManager, func(), error) {
	ldb, teardownFunc, err := SetupLDB(testName)
	bManager := consensusdatabase.New(ldb)
	return bManager, teardownFunc, err
}

func generatePubKey() (*secp256k1.SchnorrKeyPair, *secp256k1.SchnorrPublicKey, error) {
	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := privKey.SchnorrPublicKey()
	if err != nil {
		return nil, nil, err
	}
	return privKey, publicKey, nil
}

func generateScriptPublicKey(publicKey *secp256k1.SchnorrPublicKey) ([]byte, error) {
	serializedPubKey, err := publicKey.Serialize()
	if err != nil {
		return nil, err
	}
	address, err := util.NewAddressPubKeyHashFromPublicKey(util.Hash160(serializedPubKey[:]), util.Bech32PrefixKaspaTest)
	if err != nil {
		return nil, err
	}
	scriptPublicKey, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}
	return scriptPublicKey, nil
}

func NewCoinbaseData(scriptPublicKey []byte) (*externalapi.DomainCoinbaseData, error) {
	return &externalapi.DomainCoinbaseData {
		ScriptPublicKey: scriptPublicKey,
	}, nil
}

func SetupCoinbaseData() (*externalapi.DomainCoinbaseData, error) {
	_, publicKey, err := generatePubKey()
	if err != nil {
		return nil, err
	}
	scriptPublicKey, err := generateScriptPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	return NewCoinbaseData(scriptPublicKey)
}
											
func TestBlockProcessor(t *testing.T) {
	createChain := func(t *testing.T, numOfBlocks int) (model.BlockProcessor, []*externalapi.DomainBlock, func()) {
		params := &dagconfig.TestnetParams
		params.SkipProofOfWork = true

		dbManager, teardownFunc, err := SetupDBManager("TestBlockProcessor")
		if err != nil {
			t.Fatalf("Error setting up DBManager: %+v", err)
		}
		testContext, err := setupTestContext(params, dbManager)
		if err != nil {
			t.Fatalf("Error setting up TestContext: %+v", err)
		}
		blockBuilder := SetupBlockBuilder(testContext)
		blockProcessor, err := SetupBlockProcessor(testContext)
		if err != nil {
			t.Fatalf("Error setting up BlockProcessor: %+v", err)
		}

		coinbaseData := externalapi.DomainCoinbaseData{}
		blocks := make([]*externalapi.DomainBlock, numOfBlocks)
		for i := range blocks {
			block, err := blockBuilder.BuildBlock(&coinbaseData, nil)
			if err != nil {
				t.Fatalf("error in BuildBlock: %+v", err)
			}
			_, err = blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
			blocks[i] = block
		}
		return blockProcessor, blocks, teardownFunc
	}

	t.Run("Test create and process block", func(t *testing.T) {
		params := &dagconfig.TestnetParams
		params.SkipProofOfWork = true
		dbManager, teardownFunc, err := SetupDBManager("TestBlockProcessor")
		if err != nil {
			t.Fatalf("Error setting up DBManager: %+v", err)
		}
		testContext, err := setupTestContext(params, dbManager)
		if err != nil {
			t.Fatalf("Error setting up TestContext: %+v", err)
		}
		blockBuilder := SetupBlockBuilder(testContext)
		blockProcessor, err := SetupBlockProcessor(testContext)
		if err != nil {
			t.Fatalf("Error setting up BlockProcessor: %+v", err)
		}
		defer teardownFunc()
		
		// create block
		coinbaseData := &externalapi.DomainCoinbaseData{}
		block, err := blockBuilder.BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("error in BuildBlock: %+v", err)
		}
		
		// process block
		_, err = blockProcessor.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
		}
	})
	
	t.Run("Test create and process 11 blocks", func(t *testing.T) {
		params := &dagconfig.TestnetParams
		params.SkipProofOfWork = true
		dbManager, teardownFunc, err := SetupDBManager("TestBlockProcessor")
		if err != nil {
			t.Fatalf("Error setting up DBManager: %+v", err)
		}
		testContext, err := setupTestContext(params, dbManager)
		if err != nil {
			t.Fatalf("Error setting up TestContext: %+v", err)
		}
		blockBuilder := SetupBlockBuilder(testContext)
		blockProcessor, err := SetupBlockProcessor(testContext)
		if err != nil {
			t.Fatalf("Error setting up BlockProcessor: %+v", err)
		}
		defer teardownFunc()

		// create 11 blocks
		blocks := make([]*externalapi.DomainBlock, 11)
		for i := range blocks {
			coinbaseData , _ := SetupCoinbaseData()
			block, err := blockBuilder.BuildBlock(coinbaseData, nil)
			if err != nil {
				t.Fatalf("error in BuildBlock: %+v", err)
			}
			blocks[i] = block
		}

		// process 11 blocks
		for _, block := range blocks {
			_, err = blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})
	
	t.Run("Test create and double process same block", func(t *testing.T) {
		params := &dagconfig.TestnetParams
		params.SkipProofOfWork = true
		dbManager, teardownFunc, err := SetupDBManager("TestBlockProcessor")
		if err != nil {
			t.Fatalf("Error setting up DBManager: %+v", err)
		}
		testContext, err := setupTestContext(params, dbManager)
		if err != nil {
			t.Fatalf("Error setting up TestContext: %+v", err)
		}
		blockBuilder := SetupBlockBuilder(testContext)
		blockProcessor, err := SetupBlockProcessor(testContext)
		if err != nil {
			t.Fatalf("Error setting up BlockProcessor: %+v", err)
		}
		defer teardownFunc()

		coinbaseData := &externalapi.DomainCoinbaseData{}
		block, err := blockBuilder.BuildBlock(coinbaseData, nil)
		if err != nil {
			t.Fatalf("error in BuildBlock: %+v", err)
		}
		_, err = blockProcessor.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
		}
		_, err = blockProcessor.ValidateAndInsertBlock(block)
		if err == nil {
			t.Fatalf("error in ValidateAndInsertBlock: should be ErrDuplicateBlock")
		}
	})

	t.Run("Test add a chain of 50 blocks to a chain of 49 blocks", func(t *testing.T) {
		// create a chain A of 50 blocks
		_, chainABlocks, teardownChainAFunc := createChain(t, 50)
		// destory chain A
		teardownChainAFunc()
		
		// create a chain B of 49 blocks
		blockProcessor, _, teardownChainBFunc := createChain(t, 49)
		defer teardownChainBFunc()
		
		// add chain A blocks to chain B
		for i := range chainABlocks {
			block := chainABlocks[i]
			_, err := blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})
	
	t.Run("Test add a chain of 100 blocks to a chain of 100 blocks", func(t *testing.T) { //TODO: timeout
		// create a chain A of 100 blocks
		_, chainABlocks, teardownChainAFunc := createChain(t, 100)
		// destory chain A
		teardownChainAFunc()
		
		// create a chain B of 100 blocks
		blockProcessor, _, teardownChainBFunc := createChain(t, 100)
		defer teardownChainBFunc()
		
		// add chain A blocks to chain B
		for i := range chainABlocks {
			block := chainABlocks[i]
			_, err := blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})
}
