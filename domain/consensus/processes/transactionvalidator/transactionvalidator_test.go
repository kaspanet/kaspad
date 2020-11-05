package transactionvalidator

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	consensusdatabase "github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockheaderstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/blockrelationstore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/ghostdagdatastore"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures/reachabilitydatastore"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtopologymanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/dagtraversalmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/ghostdagmanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/pastmediantimemanager"
	"github.com/kaspanet/kaspad/domain/consensus/processes/reachabilitymanager"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/infrastructure/db/database/ldb"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func setupTransactionValidator(dbManager model.DBManager, dagParams *dagconfig.Params) *transactionValidator {
	// Data Structures
	blockHeaderStore := blockheaderstore.New()
	blockRelationStore := blockrelationstore.New()
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
	vlidator := New(dagParams.BlockCoinbaseMaturity,
		dbManager,
		pastMedianTimeManager,
		ghostdagDataStore)

	return vlidator.(*transactionValidator)
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

func TestValidateTransactionInIsolation(t *testing.T) {
	prevOutTxID := &externalapi.DomainTransactionID{}
	dummyPrevOut := externalapi.DomainOutpoint{TransactionID: *prevOutTxID, Index: 1}
	dummySigScript := bytes.Repeat([]byte{0x00}, 65)
	dummyTxIn := externalapi.DomainTransactionInput{
		PreviousOutpoint: dummyPrevOut,
		SignatureScript:  dummySigScript,
		Sequence:         appmessage.MaxTxInSequenceNum,
	}
	addrHash := [20]byte{0x01}
	addr, err := util.NewAddressPubKeyHash(addrHash[:], util.Bech32PrefixKaspaTest)
	if err != nil {
		t.Fatalf("NewAddressPubKeyHash: unexpected error: %v", err)
	}
	dummyScriptPublicKey, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("PayToAddrScript: unexpected error: %v", err)
	}
	dummyTxOut := externalapi.DomainTransactionOutput{
		Value:           100000000, // 1 KAS
		ScriptPublicKey: dummyScriptPublicKey,
	}

	dummyLargeTxOut := externalapi.DomainTransactionOutput{
		Value:           util.MaxSompi + 1,
		ScriptPublicKey: dummyScriptPublicKey,
	}

	payload := make([]byte, 8)
	payloadHash := externalapi.DomainHash(*daghash.DoubleHashP(payload))

	largePayload := make([]byte, constants.MaxCoinbasePayloadLength+1)

	tests := []struct {
		name    string
		tx      *externalapi.DomainTransaction
		isValid bool
	}{
		{
			name: "Valid transaction",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: true,
		},
		{
			name: "checkTransactionInputCount",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkTransactionAmountRanges",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyLargeTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkDuplicateTransactionInputs",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn, &dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
		{
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDCoinbase,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      largePayload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkTransactionPayloadHash",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDCoinbase,
				Gas:          0,
				PayloadHash:  externalapi.DomainHash{},
				Payload:      largePayload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkGasInBuiltInOrNativeTransactions",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          1,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkSubnetworkRegistryTransaction",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      nil,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkNativeTransactionPayload",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDNative,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
		{
			name: "checkTransactionSubnetwork",
			tx: &externalapi.DomainTransaction{
				Version:      appmessage.TxVersion,
				Inputs:       []*externalapi.DomainTransactionInput{&dummyTxIn},
				Outputs:      []*externalapi.DomainTransactionOutput{&dummyTxOut},
				SubnetworkID: subnetworks.SubnetworkIDRegistry,
				Gas:          0,
				PayloadHash:  payloadHash,
				Payload:      payload,
				LockTime:     0},
			isValid: false,
		},
	}

	dbManager, teardownFunc, err := setupDBManager(t.Name())
	if err != nil {
		t.Fatalf("Failed to setup DBManager instance: %v", err)
	}
	defer teardownFunc()

	validator := setupTransactionValidator(dbManager, &dagconfig.SimnetParams)

	for _, test := range tests {
		err := validator.ValidateTransactionInIsolation(test.tx)
		if test.isValid {
			if err != nil {
				t.Fatalf("ValidateTransactionInIsolation %v: %v", test.name, err)
			}
		} else {
			if err == nil {
				t.Fatalf("ValidateTransactionInIsolation:%v: Waiting for error, but got : %v", test.name, err)
			}
		}
	}
}
