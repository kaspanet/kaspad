package miningmanager_test

import (
	"bytes"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
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

func createCoinbaseTransaction(t *testing.T, scriptPublicKey []byte, value uint64) *externalapi.DomainTransaction {
	dummyTxOut := externalapi.DomainTransactionOutput{
		Value:           value,
		ScriptPublicKey: scriptPublicKey,
	}
	payload, err := coinbase.SerializeCoinbasePayload(1, &externalapi.DomainCoinbaseData{
		ScriptPublicKey:scriptPublicKey,
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
	}

	return transaction
}

func createTransaction(inputs []*externalapi.DomainTransactionInput, scriptPublicKey []byte, value uint64) *externalapi.DomainTransaction {
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
	}

	return transaction
}

func TestMiningManager(t *testing.T) {
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
		block := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

		// Check 10 transactions are not exist
		for _, tx2 := range transactions {
			for _, tx1 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					t.Fatalf("Spent tranasaction is still exisit")
				}
			}
		}
	})

	t.Run("Spending some transactions", func(t *testing.T) {
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

		// Spending first 5 transactions
		miningManager.HandleNewBlockTransactions(transactions[0:5])
		block := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

		// Check first 5 transactions are not exist
		for _, tx2 := range transactions[0:5] {
			for _, tx1 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					t.Fatalf("Spent tranasaction is still exisit")
				}
			}
		}
	})

	t.Run("Spending transactions with unknown parents", func(t *testing.T) {
		consensusInstance, err := consensusFactory.NewConsensus(dagParams, db)
		if err != nil {
			t.Fatalf("NewConsensus: %v", err)
		}

		miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)
		transactions := make([]*externalapi.DomainTransaction, 10)
		parentTransactions := make([]*externalapi.DomainTransaction, len(transactions))
		for i := range parentTransactions {
			parentTransaction := createCoinbaseTransaction(t, scriptPublicKey, uint64(100000000+i))
			parentTransactions[i] = parentTransaction
		}

		// Insert transactions with unknown parents
		for i := range transactions {
			parentTransaction := parentTransactions[i]
			txIn := externalapi.DomainTransactionInput{
				PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensusserialization.TransactionID(parentTransaction), Index: 1},
				SignatureScript:  bytes.Repeat([]byte{0x00}, 65),
				Sequence:         appmessage.MaxTxInSequenceNum,
			}
			transaction := createTransaction([]*externalapi.DomainTransactionInput{&txIn}, scriptPublicKey, uint64(10+i))
			transactions[i] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: unexpected error: %v", err)
			}
		}

		// Check transactions with unknown parents
		block := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

		for _, tx1 := range transactions {
			for _, tx2 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					t.Fatalf("Tranasaction with unknown parents is exisit")
				}
			}
		}

		// Add the missing parents
		for _, parentTransaction := range parentTransactions {
			err = miningManager.ValidateAndInsertTransaction(parentTransaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: unexpected error: %v", err)
			}
		}
		block = miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

		numberOfFoundTransactions := 0
		for _, tx1 := range transactions {
			for _, tx2 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					numberOfFoundTransactions++
					break
				}
			}
		}

		if len(transactions) != numberOfFoundTransactions{
			t.Fatalf("Not all transactions are exist, expected %v, but got %v", len(transactions), numberOfFoundTransactions)
		}
	})
}
