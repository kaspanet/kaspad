package coinbasemanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus"
	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/acceptancedatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/coinbasemanager"
	"github.com/kaspanet/kaspad/domain/consensus/utils/coinbase"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// VirtualBlockHash is a marker hash for the virtual block
var VirtualBlockHash = &externalapi.DomainHash{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
}

func createCoinbaseTransaction(t *testing.T, scriptPublicKey []byte, value uint64) *externalapi.DomainTransaction {
	dummyTxOut := externalapi.DomainTransactionOutput{
		Value:           value,
		ScriptPublicKey: scriptPublicKey,
	}
	payload, err := coinbase.SerializeCoinbasePayload(1, &externalapi.DomainCoinbaseData{
		ScriptPublicKey: scriptPublicKey,
	})
	if err != nil {
		t.Fatalf("SerializeCoinbasePayload: %v", err)
	}

	payloadHash := hashes.HashData(payload)
	transaction := &externalapi.DomainTransaction{
		Version:      constants.TransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		PayloadHash:  *payloadHash,
		Payload:      payload,
		Fee:          1000,
		Mass:         1,
	}

	return transaction
}

func setupDBForTest(dbName string) (infrastructuredatabase.Database, func(), error) {
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

	return db, teardown, err
}

func createTransaction(inputs []*externalapi.DomainTransactionInput,
	scriptPublicKey []byte, value uint64) *externalapi.DomainTransaction {
	dummyTxOut := externalapi.DomainTransactionOutput{
		Value:           value,
		ScriptPublicKey: scriptPublicKey,
	}

	transaction := &externalapi.DomainTransaction{
		Version:      constants.TransactionVersion,
		Inputs:       inputs,
		Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
		LockTime:     0,
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Fee:          1000,
	}

	return transaction
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

func createBlock(header *externalapi.DomainBlockHeader,
	transactions []*externalapi.DomainTransaction) *externalapi.DomainBlock {
	return &externalapi.DomainBlock{
		Header:       header,
		Transactions: transactions,
	}
}

func TestExpectedCoinbaseTransaction(t *testing.T) {
	dagParams := &dagconfig.SimnetParams
	consensusFactory := consensus.NewFactory()
	miningManagerFactory := miningmanager.NewFactory()
	db, teardownFunc, err := setupDBForTest(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup db: %v", err)
	}
	defer teardownFunc()

	miningAddrHash := [20]byte{0x99}
	miningAddr, err := util.NewAddressPubKeyHash(miningAddrHash[:], util.Bech32PrefixKaspaTest)
	if err != nil {
		t.Fatalf("NewAddressPubKeyHash: unexpected error: %v", err)
	}
	scriptPublicKey, err := txscript.PayToAddrScript(miningAddr)

	ghostdagDataStore := ghostdagdatastore.New()
	acceptanceDataStore := acceptancedatastore.New()

	dbManager := consensusdatabase.New(db)

	coinbaseData := &externalapi.DomainCoinbaseData{ScriptPublicKey: scriptPublicKey}

	t.Run("Spending all transactions", func(t *testing.T) {
		consensusInstance, err := consensusFactory.NewConsensus(dagParams, db)
		if err != nil {
			t.Fatalf("NewConsensus: %v", err)
		}

		// Insert 10 transactions
		miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)
		transactions := make([]*externalapi.DomainTransaction, 10)
		for i := range transactions {
			transaction := createCoinbaseTransaction(t, scriptPublicKey, uint64(100000000+i))
			transactions[i] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: unexpected error: %v", err)
			}
		}

		// Spending 10 transactions
		miningManager.HandleNewBlockTransactions(transactions)
		block, err := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if err != nil {
			t.Fatalf("GetBlockTemplate: failed building block with error: %v", err)
		}

		coinbaseManager := coinbasemanager.New(
			dbManager,
			ghostdagDataStore,
			acceptanceDataStore)

		_, err = coinbaseManager.ExpectedCoinbaseTransaction(VirtualBlockHash,
			coinbaseData) //expectedDomainTransaction
		if err != nil {
			t.Fatalf("DomainTransaction: failed getting expected coinbase transaction:error-%v", err)
		}
		// Check 10 transactions are not exist
		for _, tx2 := range transactions {
			for _, tx1 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					t.Logf("Spent tranasaction is still exisit")
				}
			}
		}
	})

}
