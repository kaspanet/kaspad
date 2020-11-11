package miningmanager_test

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/kaspanet/kaspad/domain/txscript"
	infrastructuredatabase "github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
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

func createCoinbaseTransaction(scriptPublicKey []byte, value uint64) *externalapi.DomainTransaction {
	dummyTxOut := externalapi.DomainTransactionOutput{
		Value:           value,
		ScriptPublicKey: scriptPublicKey,
	}
	payload := make([]byte, 8)
	payloadHash := externalapi.DomainHash(*daghash.DoubleHashP(payload))
	transaction := &externalapi.DomainTransaction{
		Version:      appmessage.TxVersion,
		Inputs:       nil,
		Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
		SubnetworkID: subnetworks.SubnetworkIDCoinbase,
		Gas:          0,
		PayloadHash:  payloadHash,
		Payload:      payload,
		LockTime:     0,
		Fee:          1500,
		Mass:         1500,
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
			if err != nil {
				t.Fatalf("NewConsensus: %v", err)
			}
		}

		miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)
		transactions := make([]*externalapi.DomainTransaction, 10)
		for i := range transactions {
			transaction := createCoinbaseTransaction(scriptPublicKey, uint64(100000000+i))
			transactions[i] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: unexpected error: %v", err)
			}
		}

		miningManager.HandleNewBlockTransactions(transactions)
		block := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

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
			if err != nil {
				t.Fatalf("NewConsensus: %v", err)
			}
		}

		miningManager := miningManagerFactory.NewMiningManager(consensusInstance, constants.MaxMassAcceptedByBlock)
		transactions := make([]*externalapi.DomainTransaction, 10)
		for i := range transactions {
			transaction := createCoinbaseTransaction(scriptPublicKey, uint64(100000000+i))
			transactions[i] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: unexpected error: %v", err)
			}
		}

		miningManager.HandleNewBlockTransactions(transactions[0:5])
		block := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey,
		})
		if block == nil {
			t.Fatalf("GetBlockTemplate: failed building block")
		}

		for _, tx2 := range transactions[0:5] {
			for _, tx1 := range block.Transactions {
				if consensusserialization.TransactionID(tx1) == consensusserialization.TransactionID(tx2) {
					t.Fatalf("Spent tranasaction is still exisit")
				}
			}
		}
	})
}
