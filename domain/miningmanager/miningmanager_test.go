package miningmanager_test

import (
	"github.com/kaspanet/go-secp256k1"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/pkg/errors"
	"reflect"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"

	"testing"
)

const blockMaxMass uint64 = 10000000
const coinbaseIndex = 0

// TestValidateAndInsertTransaction verifies that valid transactions were successfully inserted into the memory pool.
func TestValidateAndInsertTransaction(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {
		params.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestValidateAndInsertTransaction")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		miningManager := miningFactory.NewMiningManager(tc, blockMaxMass, false)
		transactionsToInsert := make([]*externalapi.DomainTransaction, 10)
		for i := range transactionsToInsert {
			transactionsToInsert[i] = createTransactionWithUTXOEntry(t, params, i)
			err = miningManager.ValidateAndInsertTransaction(transactionsToInsert[i], true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}
		// The UTXOEntry was filled manually for those transactions, so the transactions won't be considered orphans.
		// Therefore, all the transactions expected to be contained in the mempool.
		transactionsFromMempool := miningManager.AllTransactions()
		if len(transactionsToInsert) != len(transactionsFromMempool) {
			t.Fatalf("Wrong number of transactions in mempool: expected: %d, got: %d", len(transactionsToInsert), len(transactionsFromMempool))
		}
		for _, transactionToInsert := range transactionsToInsert {
			if !contains(transactionToInsert, transactionsFromMempool) {
				t.Fatalf("Missing transaction %v in the mempool", transactionToInsert)
			}
		}

		// The parent's transaction was inserted by consensus(AddBlock), and we want to verify that
		// the transaction is not considered an orphan and inserted into the mempool.
		transactionNotAnOrphan, _, err := createParentAndChildrenTransaction(params, tc, 0)
		if err != nil {
			t.Fatalf("Error in createParentAndChildrenTransaction: %v", err)
		}
		err = miningManager.ValidateAndInsertTransaction(transactionNotAnOrphan, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		transactionsFromMempool = miningManager.AllTransactions()
		if !contains(transactionNotAnOrphan, transactionsFromMempool) {
			t.Fatalf("Missing transaction %v in the mempool", transactionNotAnOrphan)
		}
	})
}

//	TestInsertDoubleTransactionsToMempool verifies that an attempt to insert a transaction
//	more than once into the mempool will result in raising an appropriate error.
func TestInsertDoubleTransactionsToMempool(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestInsertDoubleTransactionsToMempool")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		miningManager := miningFactory.NewMiningManager(tc, blockMaxMass, false)
		transaction := createTransactionWithUTXOEntry(t, params, 0)
		err = miningManager.ValidateAndInsertTransaction(transaction, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		err = miningManager.ValidateAndInsertTransaction(transaction, true)
		if err == nil || !strings.Contains(err.Error(), "already have transaction") {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
	})
}

// TestHandleNewBlockTransactions verifies that all the transactions in the block were successfully removed from the memory pool.
func TestHandleNewBlockTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestHandleNewBlockTransactions")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		miningManager := miningFactory.NewMiningManager(tc, blockMaxMass, false)
		transactionsToInsert := make([]*externalapi.DomainTransaction, 10)
		for i := range transactionsToInsert[(coinbaseIndex + 1):] {
			transaction := createTransactionWithUTXOEntry(t, params, i)
			transactionsToInsert[i+1] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}

		_, err = miningManager.HandleNewBlockTransactions(transactionsToInsert[0:3])
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		mempoolTransactions := miningManager.AllTransactions()
		for _, RemovedTransaction := range transactionsToInsert[(coinbaseIndex + 1):3] {
			if contains(RemovedTransaction, mempoolTransactions) {
				t.Fatalf("This transaction shouldnt be in mempool: %v", RemovedTransaction)
			}
		}

		// There are no chained/double-spends transactions, and hence it is expected that all the other
		// transactions, will still be included in the mempool.
		mempoolTransactions = miningManager.AllTransactions()
		for i, transaction := range transactionsToInsert[3:] {
			if !contains(transaction, mempoolTransactions) {
				t.Fatalf("This transaction %d should be in mempool: %v", i, transaction)
			}
		}

		_, err = miningManager.HandleNewBlockTransactions(transactionsToInsert[2:])
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		if len(miningManager.AllTransactions()) != 0 {
			t.Fatalf("The mempool contains unexpected transactions: %v", miningManager.AllTransactions())
		}
	})
}

// TestDoubleSpends verifies that any transactions which are now double spends as a result of the block's new transactions
// will be removed from the mempool.
func TestDoubleSpends(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestDoubleSpends")
		if err != nil {
			t.Fatalf("Failed setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		miningManager := miningFactory.NewMiningManager(tc, blockMaxMass, false)
		transactionInTheMempool := createTransactionWithUTXOEntry(t, params, 0)
		err = miningManager.ValidateAndInsertTransaction(transactionInTheMempool, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		doubleSpendTransactionInTheBlock := createTransactionWithUTXOEntry(t, params, 0)
		doubleSpendTransactionInTheBlock.Inputs[0].PreviousOutpoint = transactionInTheMempool.Inputs[0].PreviousOutpoint
		blockTransactions := []*externalapi.DomainTransaction{nil, doubleSpendTransactionInTheBlock}
		_, err = miningManager.HandleNewBlockTransactions(blockTransactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		if contains(transactionInTheMempool, miningManager.AllTransactions()) {
			t.Fatalf("The transaction %v, shouldn't be in the mempool, since at least one output was already spent.", transactionInTheMempool)
		}
	})
}

// TestOrphanTransactions verifies that a transaction could be a part of a new block template, only if it's not an orphan.
func TestOrphanTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, params *dagconfig.Params) {

		params.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(params, false, "TestOrphanTransactions")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		miningManager := miningFactory.NewMiningManager(tc, blockMaxMass, false)
		parentTransactions, childTransactions, err := createArraysOfParentAndChildrenTransactions(params, tc)
		if err != nil {
			t.Fatalf("Error in createArraysOfParentAndChildrenTransactions: %v", err)
		}
		for _, orphanTransaction := range childTransactions {
			err = miningManager.ValidateAndInsertTransaction(orphanTransaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}
		transactionsMempool := miningManager.AllTransactions()
		for _, transaction := range transactionsMempool {
			if contains(transaction, childTransactions) {
				t.Fatalf("Error: an orphan transaction is exist in the mempool")
			}
		}

		block, err := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0},
			ExtraData:       nil})
		if err != nil {
			t.Fatalf("Failed get a block template: %v", err)
		}
		for _, transactionFromBlock := range block.Transactions[1:] {
			for _, orphanTransaction := range childTransactions {
				if consensushashing.TransactionID(transactionFromBlock) == consensushashing.TransactionID(orphanTransaction) {
					t.Fatalf("Tranasaction with unknown parents is exist in a block that was built from GetTemplate option.")
				}
			}
		}

		blockParentsTransactionsHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, parentTransactions)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}
		blockParentsTransactions, err := tc.GetBlock(blockParentsTransactionsHash)
		if err != nil {
			t.Fatalf("GetBlock: %v", err)
		}
		_, err = miningManager.HandleNewBlockTransactions(blockParentsTransactions.Transactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		transactionsMempool = miningManager.AllTransactions()
		for _, transaction := range transactionsMempool {
			if !contains(transaction, childTransactions) {
				t.Fatalf("Error: the transaction %v, should be in the mempool since its not oprhan anymore.", transaction)
			}
		}
		block, err = miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0},
			ExtraData:       nil})
		if err != nil {
			t.Fatalf("GetBlockTemplate: %v", err)
		}
		for _, transactionFromBlock := range block.Transactions[1:] {
			isContained := false
			for _, childTransaction := range childTransactions {
				if consensushashing.TransactionID(transactionFromBlock) == consensushashing.TransactionID(childTransaction) {
					isContained = true
					break
				}
			}
			if !isContained {
				t.Fatalf("Error: Unknown Transaction %v in a block.", transactionFromBlock)
			}
		}
	})
}

func createTransactionWithUTXOEntry(t *testing.T, params *dagconfig.Params, i int) *externalapi.DomainTransaction {
	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("Failed generate a private key: %v", err)
	}
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		t.Fatalf("Failed generate a public key: %v", err)
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		t.Fatalf("Failed serialize public key: %v", err)
	}
	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		t.Fatalf("Failed generate p2pkh address: %v", err)
	}
	scriptPublicKey, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("PayToAddrScript: %v", err)
	}
	prevOutTxID := externalapi.DomainTransactionID{}
	prevOutPoint := externalapi.DomainOutpoint{TransactionID: prevOutTxID, Index: uint32(i)}
	txInputWithMaxSequence := externalapi.DomainTransactionInput{
		PreviousOutpoint: prevOutPoint,
		SignatureScript:  []byte{},
		Sequence:         constants.SequenceLockTimeIsSeconds,
		UTXOEntry: utxo.NewUTXOEntry(
			100000000, // 1 KAS
			scriptPublicKey,
			true,
			uint64(5)),
	}
	txOut := externalapi.DomainTransactionOutput{
		Value:           10000,
		ScriptPublicKey: scriptPublicKey,
	}
	validTx := externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{&txInputWithMaxSequence},
		Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Fee:          289,
		Mass:         1,
		LockTime:     0}

	signatureScript, err := txscript.SignatureScript(&validTx, 0, scriptPublicKey, txscript.SigHashAll, privateKey)
	if err != nil {
		t.Fatalf("Failed to create a sigScript: %v", err)
	}
	validTx.Inputs[0].SignatureScript = signatureScript
	return &validTx
}

func createArraysOfParentAndChildrenTransactions(params *dagconfig.Params, tc testapi.TestConsensus) ([]*externalapi.DomainTransaction,
	[]*externalapi.DomainTransaction, error) {

	transactions := make([]*externalapi.DomainTransaction, 5)
	parentTransactions := make([]*externalapi.DomainTransaction, len(transactions))
	var err error
	for i := range transactions {
		parentTransactions[i], transactions[i], err = createParentAndChildrenTransaction(params, tc, i)
		if err != nil {
			return nil, nil, err
		}
	}
	return parentTransactions, transactions, nil
}

func createParentAndChildrenTransaction(params *dagconfig.Params, tc testapi.TestConsensus, i int) (*externalapi.DomainTransaction,
	*externalapi.DomainTransaction, error) {

	privateKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed generate private key: ")
	}
	publicKey, err := privateKey.SchnorrPublicKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed generate public key: ")
	}
	publicKeySerialized, err := publicKey.Serialize()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed serialize a public key:")
	}
	addr, err := util.NewAddressPubKeyHashFromPublicKey(publicKeySerialized[:], params.Prefix)
	if err != nil {
		return nil, nil, errors.Wrap(err, "NewAddressPubKeyHashFromPublicKey: ")
	}
	scriptPublicKey, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed generate a scriptPublicKey:")
	}

	firstBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "AddBlock: %v", err)
	}
	fundingBlockHashForParent, _, err := tc.AddBlock([]*externalapi.DomainHash{firstBlockHash}, nil, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "AddBlock: ")
	}
	fundingBlockForParent, err := tc.GetBlock(fundingBlockHashForParent)
	if err != nil {
		return nil, nil, errors.Wrap(err, "GetBlock: ")
	}
	fundingTransactionForParent := fundingBlockForParent.Transactions[transactionhelper.CoinbaseTransactionIndex]
	_, redeemScript := testutils.OpTrueScript()
	// Change the value to get different ID transactions each time.
	fundingTransactionForParent.Outputs[0].Value -= uint64(i)
	signatureScriptCheck, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
	if err != nil {
		return nil, nil, err
	}

	txInputForParent := externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(fundingTransactionForParent),
			Index: 0},
		SignatureScript: signatureScriptCheck,
		Sequence:        constants.SequenceLockTimeIsSeconds,
		UTXOEntry:       nil,
	}
	txOutForParent := externalapi.DomainTransactionOutput{
		Value:           10000,
		ScriptPublicKey: scriptPublicKey,
	}
	txParent := externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{&txInputForParent},
		Outputs:      []*externalapi.DomainTransactionOutput{&txOutForParent},
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Payload:      []byte{},
		Gas:          0,
		Mass:         1,
		LockTime:     0}

	txInputForChild := externalapi.DomainTransactionInput{
		PreviousOutpoint: externalapi.DomainOutpoint{TransactionID: *consensushashing.TransactionID(&txParent), Index: uint32(0)},
		SignatureScript:  []byte{},
		Sequence:         constants.SequenceLockTimeIsSeconds,
		UTXOEntry:        nil,
	}
	txOutForChild := externalapi.DomainTransactionOutput{
		Value:           9000,
		ScriptPublicKey: scriptPublicKey,
	}
	txChild := externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{&txInputForChild},
		Outputs:      []*externalapi.DomainTransactionOutput{&txOutForChild},
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Fee:          289,
		Mass:         1,
		LockTime:     0}
	signatureScript, err := txscript.SignatureScript(&txChild, 0, scriptPublicKey, txscript.SigHashAll, privateKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed creating a signatureScript")
	}
	txChild.Inputs[0].SignatureScript = signatureScript

	return &txParent, &txChild, nil
}

func contains(transaction *externalapi.DomainTransaction, transactions []*externalapi.DomainTransaction) bool {
	for _, candidateTransaction := range transactions {
		if reflect.DeepEqual(candidateTransaction, transaction) {
			return true
		}
	}
	return false
}
