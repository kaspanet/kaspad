package blockprocessor_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
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
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func SetupBlockProcessorWithDB(dbName string, dagParams *dagconfig.Params) (model.BlockProcessor, func(), error) {
	var err error
	tmpDir, err := ioutil.TempDir("", "SetupBlockProcessorWithDB")
	if err != nil {
		return nil, nil, errors.Errorf("error creating temp dir: %s", err)
	}

	dbPath := filepath.Join(tmpDir, dbName)
	_ = os.RemoveAll(dbPath)
	db, err := ldb.NewLevelDB(dbPath)
	if err != nil {
		return nil, nil, err
	}

	originalLDBOptions := ldb.Options
	ldb.Options = func() *opt.Options {
		return nil
	}

	teardown := func() {
		db.Close()
		ldb.Options = originalLDBOptions
		os.RemoveAll(dbPath)
	}

	blockProcessor := SetupBlockProcessor(db, dagParams)

	return blockProcessor, teardown, err
}

func SetupBlockProcessor(db database.Database, dagParams *dagconfig.Params) model.BlockProcessor {
	acceptanceDataStore := acceptancedatastore.New()
	blockStore := blockstore.New()
	blockHeaderStore := blockheaderstore.New()
	blockRelationStore := blockrelationstore.New()
	blockStatusStore := blockstatusstore.New()
	multisetStore := multisetstore.New()
	pruningStore := pruningstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	utxoDiffStore := utxodiffstore.New()
	consensusStateStore := consensusstatestore.New()
	ghostdagDataStore := ghostdagdatastore.New()

	dbManager := consensusdatabase.New(db)

	reachabilityManager := reachabilitymanager.New(
		dbManager,
		ghostdagDataStore,
		blockRelationStore,
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
		ghostdagDataStore)
	pruningManager := pruningmanager.New(
		dagTraversalManager,
		dagTopologyManager,
		pruningStore,
		blockStatusStore,
		consensusStateStore)
	consensusStateManager := consensusstatemanager.New(
		dbManager,
		dagParams,
		ghostdagManager,
		dagTopologyManager,
		pruningManager,
		blockStatusStore,
		ghostdagDataStore,
		consensusStateStore,
		multisetStore,
		blockStore,
		utxoDiffStore,
		blockRelationStore,
		acceptanceDataStore,
		blockHeaderStore)
	difficultyManager := difficultymanager.New(
		ghostdagManager)
	pastMedianTimeManager := pastmediantimemanager.New(
		dagParams.TimestampDeviationTolerance,
		dbManager,
		dagTraversalManager,
		blockHeaderStore)
	transactionValidator := transactionvalidator.New(dagParams.BlockCoinbaseMaturity,
		dbManager,
		pastMedianTimeManager,
		ghostdagDataStore)
	coinbaseManager := coinbasemanager.New(
		ghostdagDataStore,
		acceptanceDataStore)
	genesisHash := externalapi.DomainHash(*dagParams.GenesisHash)
	blockValidator := blockvalidator.New(
		dagParams.PowMax,
		false,
		&genesisHash,
		dagParams.EnableNonNativeSubnetworks,
		dagParams.DisableDifficultyAdjustment,
		dagParams.DifficultyAdjustmentWindowSize,
		uint64(dagParams.FinalityDuration/dagParams.TargetTimePerBlock),

		dbManager,
		consensusStateManager,
		difficultyManager,
		pastMedianTimeManager,
		transactionValidator,
		ghostdagManager,
		dagTopologyManager,
		dagTraversalManager,

		blockStore,
		ghostdagDataStore,
		blockHeaderStore,
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
		blockHeaderStore)

	return blockProcessor
}

func TestBlockProcessor(t *testing.T) {
	createChain := func(t *testing.T, numOfBlocks int) (blockProcessor model.BlockProcessor, идщслы []*externalapi.DomainBlock, teardownFunc func()) {
		blockProcessor, teardownFunc, err := SetupBlockProcessorWithDB(t.Name(), &dagconfig.SimnetParams)
		if err != nil {
			t.Fatalf("Failed to setup blockProcessor instance: %v", err)
		}

		blocks := make([]*externalapi.DomainBlock, numOfBlocks)
		for i := range blocks {
			block, err := blockProcessor.BuildBlock(nil, nil)
			if err != nil {
				t.Fatalf("error in BuildBlock: %+v", err)
			}

			err = blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}

			blocks[i] = block
		}

		return blockProcessor, blocks, teardownFunc
	}

	t.Run("Test create and process block", func(t *testing.T) {
		blockProcessor, teardownFunc, err := SetupBlockProcessorWithDB(t.Name(), &dagconfig.SimnetParams)
		if err != nil {
			t.Fatalf("Failed to setup blockProcessor instance: %v", err)
		}
		defer teardownFunc()

		// create block
		block, err := blockProcessor.BuildBlock(nil, nil)
		if err != nil {
			t.Fatalf("error in BuildBlock: %+v", err)
		}

		// process block
		err = blockProcessor.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
		}
	})

	t.Run("Test create and process 11 blocks", func(t *testing.T) {
		blockProcessor, teardownFunc, err := SetupBlockProcessorWithDB(t.Name(), &dagconfig.SimnetParams)
		if err != nil {
			t.Fatalf("Failed to setup blockProcessor instance: %v", err)
		}
		defer teardownFunc()

		// create 11 blocks
		blocks := make([]*externalapi.DomainBlock, 11)
		for i := range blocks {
			block, err := blockProcessor.BuildBlock(nil, nil)
			if err != nil {
				t.Fatalf("error in BuildBlock: %+v", err)
			}
			blocks[i] = block
		}

		// process 11 blocks
		for _, block := range blocks {
			err = blockProcessor.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})

	t.Run("Test create and double process same block", func(t *testing.T) {
		blockProcessor, teardownFunc, err := SetupBlockProcessorWithDB(t.Name(), &dagconfig.SimnetParams)
		if err != nil {
			t.Fatalf("Failed to setup blockProcessor instance: %v", err)
		}
		defer teardownFunc()

		// create and double process same block
		block, err := blockProcessor.BuildBlock(nil, nil)
		if err != nil {
			t.Fatalf("error in BuildBlock: %+v", err)
		}
		err = blockProcessor.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
		}
		err = blockProcessor.ValidateAndInsertBlock(block)
		if err != nil {
			t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
		}
	})

	t.Run("Test add a chain of 50 blocks to a chain of 49 blocks", func(t *testing.T) {
		// create a chain A of 50 blocks
		_, chainABlocks, teardownChainAFunc := createChain(t, 50)
		// destory chain A
		teardownChainAFunc()

		// create a chain B of 49 blocks
		cahinB, _, teardownChainBFunc := createChain(t, 49)
		defer teardownChainBFunc()

		// add chain A blocks to chain B
		for i := range chainABlocks {
			block := chainABlocks[i]
			err := cahinB.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})

	t.Run("Test add a chain of 86401 blocks to a chain of 86401 blocks", func(t *testing.T) {
		// create a chain A of 86401 blocks
		_, chainABlocks, teardownChainAFunc := createChain(t, 86401)
		// destory chain A
		teardownChainAFunc()

		// create a chain B of 86401 blocks
		chainB, _, teardownChainBFunc := createChain(t, 86401)
		defer teardownChainBFunc()

		// add chain A blocks to chain B
		for i := range chainABlocks {
			block := chainABlocks[i]
			err := chainB.ValidateAndInsertBlock(block)
			if err != nil {
				t.Fatalf("error in ValidateAndInsertBlock: %+v", err)
			}
		}
	})
}
