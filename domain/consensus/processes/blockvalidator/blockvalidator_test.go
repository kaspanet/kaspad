package blockvalidator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstatusstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/mergedepthmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/transactionvalidator"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/merkle"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/kaspanet/kaspad/domain/dagconfig"
)

const difficultyForTest = uint32(0x207f83df)

type mocDifficultyManager struct {
}

func (mdf *mocDifficultyManager) RequiredDifficulty(blockHash *externalapi.DomainHash) (uint32, error) {
	return difficultyForTest, nil
}

func setupBlockValidator(dbManager model.DBManager, dagParams *dagconfig.Params) *blockValidator {
	// Data Structures
	acceptanceDataStore := acceptancedatastore.New()
	blockStore := blockstore.New()
	blockHeaderStore := blockheaderstore.New()
	blockRelationStore := blockrelationstore.New()
	blockStatusStore := blockstatusstore.New()
	reachabilityDataStore := reachabilitydatastore.New()
	ghostdagDataStore := ghostdagdatastore.New()

	// Processes
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
		ghostdagDataStore,
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
	difficultyManager := &mocDifficultyManager{}
	coinbaseManager := coinbasemanager.New(
		dbManager,
		ghostdagDataStore,
		acceptanceDataStore)
	genesisHash := externalapi.DomainHash(*dagParams.GenesisHash)
	mergeDepthManager := mergedepthmanager.New(
		dagParams.FinalityDepth(),
		dbManager,
		dagTopologyManager,
		dagTraversalManager,
		ghostdagDataStore)
	vlidator := New(
		dagParams.PowMax,
		true,
		&genesisHash,
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

	return vlidator.(*blockValidator)
}

func createBlock(header *externalapi.DomainBlockHeader,
	transactions []*externalapi.DomainTransaction) *externalapi.DomainBlock {
	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
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

		headerHash := hashserialization.HeaderHash(header)
		result[i] = headerHash
	}

	return result
}

func setupDBManager(dbName string) (model.DBManager, func(), error) {
	var err error
	tmpDir, err := ioutil.TempDir("", "setupDBManager")
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

func TestValidateInvalidBlockInternal(t *testing.T) {
	dbManager, teardownFunc, err := setupDBManager(t.Name())
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
		Bits:                 9999999,
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
	block := createBlock(blockHeader, transactions)
	blockHash := hashserialization.HeaderHash(block.Header)
	validator.blockStore.Stage(blockHash, block)
	validator.blockHeaderStore.Stage(blockHash, blockHeader)
	validator.ghostdagDataStore.Stage(blockHash, &model.BlockGHOSTDAGData{
		SelectedParent: &genesisHash,
		MergeSetBlues:  make([]*externalapi.DomainHash, 1000),
		MergeSetReds:   make([]*externalapi.DomainHash, 1),
	})
	blockWithoutTransactions := createBlock(blockHeader, nil)

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

	// checkParentsIncest
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
	err = validator.checkMergeSizeLimit(blockHash)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockTransactionsFinalized
	err = validator.checkBlockTransactionsFinalized(blockHash)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockSize
	err = validator.checkBlockSize(block)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// checkBlockContainsAtLeastOneTransaction
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

	// checkProofOfWork
	err = validator.checkProofOfWork(blockHeader)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}

	// validateDifficulty
	err = validator.validateDifficulty(blockHash)
	if err == nil {
		t.Fatalf("Waiting for error, but got: %s", err)
	}
}

func TestValidateValidBlock(t *testing.T) {
	dbManager, teardownFunc, err := SetupDBManager(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup DBManager instance: %v", err)
	}
	defer teardownFunc()

	validator := setupBlockValidator(dbManager, &dagconfig.SimnetParams)
	var time int64 = 0
	genesisHash := externalapi.DomainHash(*dagconfig.SimnetParams.GenesisHash)

	parentHashes := prepareParentHashes(10, []*externalapi.DomainHash{&genesisHash}, &time)
	for _, parentHash := range parentHashes {
		validator.blockHeaderStore.Stage(parentHash, nil)
	}
	sort.Slice(parentHashes, func(i, j int) bool {
		return hashes.Less(parentHashes[i], parentHashes[j])
	})

	transactions := make([]*externalapi.DomainTransaction, 1)
	inputs := make([]*externalapi.DomainTransactionInput, 1)
	for i := range inputs {
		inputs[i] = &externalapi.DomainTransactionInput{}
	}
	for i := range transactions {
		payload := make([]byte, 8)
		payloadHash := (externalapi.DomainHash)(*daghash.DoubleHashP(payload))

		transactions[i] = &externalapi.DomainTransaction{
			Version:      1,
			Inputs:       inputs,
			Outputs:      nil,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDRegistry,
			Gas:          0,
			PayloadHash:  payloadHash,
			Payload:      payload,
			Fee:          0,
			Mass:         0,
		}
	}

	coinbaseTransactions := make([]*externalapi.DomainTransaction, 1)
	for i := range coinbaseTransactions {
		payload := make([]byte, 30)
		payloadHash := (externalapi.DomainHash)(*daghash.DoubleHashP(payload))
		coinbaseTransactions[i] = &externalapi.DomainTransaction{
			Version:      1,
			Inputs:       nil,
			Outputs:      nil,
			LockTime:     0,
			SubnetworkID: subnetworks.SubnetworkIDCoinbase,
			Gas:          0,
			PayloadHash:  payloadHash,
			Payload:      payload,
			Fee:          0,
			Mass:         0,
		}
	}
	transactions = append(coinbaseTransactions, transactions...)

	blockHeader := &externalapi.DomainBlockHeader{
		Version:              1,
		ParentHashes:         parentHashes,
		HashMerkleRoot:       *merkle.CalculateHashMerkleRoot(transactions),
		AcceptedIDMerkleRoot: externalapi.DomainHash{},
		UTXOCommitment:       externalapi.DomainHash{},
		TimeInMilliseconds:   mstime.Now().UnixMilliseconds() + int64(len(parentHashes)),
		Bits:                 difficultyForTest,
	}

	blockWithThreeTx := createBlock(blockHeader, transactions)
	blockHash := hashserialization.HeaderHash(blockWithThreeTx.Header)
	validator.blockStore.Stage(blockHash, blockWithThreeTx)
	validator.blockHeaderStore.Stage(blockHash, blockHeader)
	validator.ghostdagDataStore.Stage(blockHash, &model.BlockGHOSTDAGData{
		SelectedParent: &genesisHash,
		MergeSetBlues:  make([]*externalapi.DomainHash, 1000),
		MergeSetReds:   make([]*externalapi.DomainHash, 1),
	})

	blockWithCoinbaseTx := createBlock(blockHeader, coinbaseTransactions)
	testedBlocks := []*externalapi.DomainBlock{blockWithCoinbaseTx, blockWithThreeTx}

	for _, block := range testedBlocks {
		blockHash := hashserialization.HeaderHash(block.Header)
		err = validator.ValidateHeaderInIsolation(blockHash)
		if err != nil {
			t.Fatalf("ValidateHeaderInIsolation: %v", err)
		}

		err = validator.ValidateBodyInIsolation(blockHash)
		if err != nil {
			t.Fatalf("ValidateBodyInIsolation: %v", err)
		}

		err = validator.ValidateProofOfWorkAndDifficulty(blockHash)
		if err != nil {
			t.Fatalf("ValidateProofOfWorkAndDifficulty: %v", err)
		}
	}
}
