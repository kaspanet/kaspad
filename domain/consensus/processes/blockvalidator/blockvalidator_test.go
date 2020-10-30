package blockvalidator

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
	"github.com/kaspanet/kaspad/domain/consensus/processes/consensusstatemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/difficultymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pruningmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

func setupBlockValidator(dbManager model.DBManager, dagParams *dagconfig.Params) *blockValidator {
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
	genesisHash := externalapi.DomainHash(*dagParams.GenesisHash)
	validator := New(
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

	return validator.(*blockValidator)
}

func createBlock(header *externalapi.DomainBlockHeader,
	transactions []*externalapi.DomainTransaction) *externalapi.DomainBlock {
	headerHash := hashserialization.HeaderHash(header)
	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
		Hash:         headerHash,
	}
}

func prepareParentHashes(numOfBlocks int, parents []*externalapi.DomainHash, time *int64) []*externalapi.DomainHash {
	result := make([]*externalapi.DomainHash, numOfBlocks)
	for i := range result {
		*time++
		header := &externalapi.DomainBlockHeader{
			Version:              1,
			ParentHashes:         parents,
			HashMerkleRoot:       externalapi.DomainHash{},
			AcceptedIDMerkleRoot: externalapi.DomainHash{},
			UTXOCommitment:       externalapi.DomainHash{},
			TimeInMilliseconds:   *time,
			Bits:                 0,
		}

		result[i] = createBlock(header, nil).Hash
	}

	return result
}

func SetupDBManager(dbName string) (model.DBManager, func(), error) {
	var err error
	tmpDir, err := ioutil.TempDir("", "SetupDBManager")
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

	dbManager := consensusdatabase.New(db)
	return dbManager, teardown, err
}

func TestValidateHeaderInIsolation(t *testing.T) {
	dbManager, teardownFunc, err := SetupDBManager(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup DBManager instance: %v", err)
	}
	defer teardownFunc()

	validator := setupBlockValidator(dbManager, &dagconfig.SimnetParams)
	var time int64 = 0
	genesisHash := externalapi.DomainHash(*dagconfig.SimnetParams.GenesisHash)

	parentHashes := prepareParentHashes(20, []*externalapi.DomainHash{&genesisHash}, &time)
	blockHeader := &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       externalapi.DomainHash{},
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   mstime.Now().UnixMilliseconds() + int64(len(parentHashes)),
		Bits:                 0,
	}

	transactions := make([]*externalapi.DomainTransaction, 1000)
	inputs := make([]*externalapi.DomainTransactionInput, 100)
	for i := range inputs {
		inputs[i] = &externalapi.DomainTransactionInput{}
	}
	for i := range transactions {
		transactions[i] = &externalapi.DomainTransaction{
			Version:      1,
			Inputs:       inputs,
			Outputs:      nil,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			PayloadHash:  externalapi.DomainHash{},
			Payload:      make([]byte, 0),
			Fee:          0,
			Mass:         0,
		}
	}

	// create chained transctions
	previousOutpoint := &externalapi.DomainOutpoint{
		TransactionID: *hashserialization.TransactionID(transactions[0]),
	}
	chinedInput := &externalapi.DomainTransactionInput{
		PreviousOutpoint: *previousOutpoint,
	}
	chinedTransactions := make([]*externalapi.DomainTransaction, 1000)
	for i := range chinedTransactions {
		chinedTransactions[i] = &externalapi.DomainTransaction{
			Version:      1,
			Inputs:       []*externalapi.DomainTransactionInput{chinedInput},
			Outputs:      nil,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDNative,
			Gas:          0,
			PayloadHash:  externalapi.DomainHash{},
			Payload:      make([]byte, 0),
			Fee:          0,
			Mass:         0,
		}
	}
	transactions = append(transactions, chinedTransactions...)

	// create multiple coinbase transactions
	coinbaseTransactions := make([]*externalapi.DomainTransaction, 3)
	for i := range coinbaseTransactions {
		coinbaseTransactions[i] = &externalapi.DomainTransaction{
			Version:      1,
			Inputs:       inputs,
			Outputs:      nil,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDCoinbase,
			Gas:          0,
			PayloadHash:  externalapi.DomainHash{},
			Payload:      make([]byte, 0),
			Fee:          0,
			Mass:         0,
		}
	}
	transactions = append(transactions, coinbaseTransactions...)

	//subnetworks.SubnetworkIDCoinbase
	block := createBlock(blockHeader, transactions)

	// checkParentsLimit
	err = validator.checkParentsLimit(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockParentsOrder
	err = checkBlockParentsOrder(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockTransactionsFinalized
	err = validator.checkParentsIncest(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// validateMedianTime
	err = validator.validateMedianTime(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkMergeSizeLimit
	err = validator.checkMergeSizeLimit(block.Hash)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// validateMedianTime
	err = validator.validateMedianTime(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockTransactionsFinalized
	err = validator.checkBlockTransactionsFinalized(block.Hash)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockSize
	err = validator.checkBlockSize(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockContainsAtLeastOneTransaction
	blockWithoutTransactions := createBlock(blockHeader, nil)
	err = validator.checkBlockContainsAtLeastOneTransaction(blockWithoutTransactions)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkFirstBlockTransactionIsCoinbase
	err = validator.checkFirstBlockTransactionIsCoinbase(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockContainsOnlyOneCoinbase
	err = validator.checkBlockContainsOnlyOneCoinbase(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkTransactionsInIsolation
	err = validator.checkTransactionsInIsolation(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockHashMerkleRoot
	err = validator.checkBlockHashMerkleRoot(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockDuplicateTransactions
	err = validator.checkBlockDuplicateTransactions(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockDoubleSpends
	err = validator.checkBlockDoubleSpends(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockHasNoChainedTransactions
	err = validator.checkBlockHasNoChainedTransactions(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}
}
