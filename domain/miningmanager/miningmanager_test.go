package miningmanager_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/pkg/errors"
	"strings"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"

	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/dagconfig"

	"testing"
)

const blockMaxMass uint64 = 10000000

// TestValidateAndInsertTransaction verifies that valid transactions were successfully inserted into the mempool.
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
			transactionsToInsert[i] = createTransactionWithUTXOEntry(t, i)
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
		transactionNotAnOrphan, err := createChildTxWhenParentTxWasAddedByConsensus(params, tc)
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

// TestInsertDoubleTransactionsToMempool verifies that an attempt to insert a transaction
// more than once into the mempool will result in raising an appropriate error.
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
		transaction := createTransactionWithUTXOEntry(t, 0)
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

// TestHandleNewBlockTransactions verifies that all the transactions in the block were successfully removed from the mempool.
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
		for i := range transactionsToInsert[(transactionhelper.CoinbaseTransactionIndex + 1):] {
			transaction := createTransactionWithUTXOEntry(t, i)
			transactionsToInsert[i+1] = transaction
			err = miningManager.ValidateAndInsertTransaction(transaction, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}

		const partialLength = 3
		_, err = miningManager.HandleNewBlockTransactions(transactionsToInsert[0:partialLength])
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		mempoolTransactions := miningManager.AllTransactions()
		for _, removedTransaction := range transactionsToInsert[(transactionhelper.CoinbaseTransactionIndex + 1):partialLength] {
			if contains(removedTransaction, mempoolTransactions) {
				t.Fatalf("This transaction shouldnt be in mempool: %v", removedTransaction)
			}
		}

		// There are no chained/double-spends transactions, and hence it is expected that all the other
		// transactions, will still be included in the mempool.
		mempoolTransactions = miningManager.AllTransactions()
		for i, transaction := range transactionsToInsert[partialLength:] {
			if !contains(transaction, mempoolTransactions) {
				t.Fatalf("This transaction %d should be in mempool: %v", i, transaction)
			}
		}
		// The first index considers as coinbase, therefore in order that all the transactions will insert into
		// the mempool, we will start one index less (partialLength - 1).
		// Handle all the other transactions in the transactionsToInsert array.
		_, err = miningManager.HandleNewBlockTransactions(transactionsToInsert[(partialLength - 1):])
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
		transactionInTheMempool := createTransactionWithUTXOEntry(t, 0)
		err = miningManager.ValidateAndInsertTransaction(transactionInTheMempool, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		doubleSpendTransactionInTheBlock := createTransactionWithUTXOEntry(t, 0)
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
		// Before each parent transaction, We will add two blocks by consensus in order to fund the parent transactions.
		parentTransactions, childTransactions, err := createArraysOfParentAndChildrenTransactions(tc)
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
		tips, err := tc.Tips()
		if err != nil {
			t.Fatalf("Tips: %v.", err)
		}
		blockParentsTransactionsHash, _, err := tc.AddBlock(tips, nil, parentTransactions)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		_, _, err = tc.AddBlock([]*externalapi.DomainHash{blockParentsTransactionsHash}, nil, nil)
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
		if len(transactionsMempool) != len(childTransactions) {
			t.Fatalf("Expected %d transactions in the mempool but got %d", len(childTransactions), len(transactionsMempool))
		}

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
				if *consensushashing.TransactionID(transactionFromBlock) == *consensushashing.TransactionID(childTransaction) {
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

func createTransactionWithUTXOEntry(t *testing.T, i int) *externalapi.DomainTransaction {
	prevOutTxID := externalapi.DomainTransactionID{}
	prevOutPoint := externalapi.DomainOutpoint{TransactionID: prevOutTxID, Index: uint32(i)}
	scriptPublicKey, redeemScript := testutils.OpTrueScript()
	signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
	if err != nil {
		t.Fatalf("PayToScriptHashSignatureScript: %v", err)
	}
	txInputWithMaxSequence := externalapi.DomainTransactionInput{
		PreviousOutpoint: prevOutPoint,
		SignatureScript:  signatureScript,
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
	tx := externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{&txInputWithMaxSequence},
		Outputs:      []*externalapi.DomainTransactionOutput{&txOut},
		SubnetworkID: subnetworks.SubnetworkIDNative,
		Gas:          0,
		Fee:          289,
		Mass:         1,
		LockTime:     0}

	return &tx
}

func createArraysOfParentAndChildrenTransactions(tc testapi.TestConsensus) ([]*externalapi.DomainTransaction,
	[]*externalapi.DomainTransaction, error) {

	const numOfTransactions = 5
	transactions := make([]*externalapi.DomainTransaction, numOfTransactions)
	parentTransactions := make([]*externalapi.DomainTransaction, len(transactions))
	var err error
	for i := range transactions {
		parentTransactions[i], transactions[i], err = createParentAndChildrenTransactions(tc)
		if err != nil {
			return nil, nil, err
		}
	}
	return parentTransactions, transactions, nil
}

func createParentAndChildrenTransactions(tc testapi.TestConsensus) (*externalapi.DomainTransaction,
	*externalapi.DomainTransaction, error) {

	// We will add two blocks by consensus before the parent transactions, in order to fund the parent transactions.
	tips, err := tc.Tips()
	if err != nil {
		return nil, nil, err
	}

	_, _, err = tc.AddBlock(tips, nil, nil)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "AddBlock: %v", err)
	}

	tips, err = tc.Tips()
	if err != nil {
		return nil, nil, err
	}

	fundingBlockHashForParent, _, err := tc.AddBlock(tips, nil, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "AddBlock: ")
	}
	fundingBlockForParent, err := tc.GetBlock(fundingBlockHashForParent)
	if err != nil {
		return nil, nil, errors.Wrap(err, "GetBlock: ")
	}
	fundingTransactionForParent := fundingBlockForParent.Transactions[transactionhelper.CoinbaseTransactionIndex]
	txParent, err := testutils.CreateTransaction(fundingTransactionForParent, 1000)
	if err != nil {
		return nil, nil, err
	}
	txChild, err := testutils.CreateTransaction(txParent, 1000)
	if err != nil {
		return nil, nil, err
	}
	return txParent, txChild, nil
}

func createChildTxWhenParentTxWasAddedByConsensus(params *dagconfig.Params, tc testapi.TestConsensus) (*externalapi.DomainTransaction, error) {

	firstBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{params.GenesisHash}, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "AddBlock: %v", err)
	}
	ParentBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{firstBlockHash}, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "AddBlock: ")
	}
	ParentBlock, err := tc.GetBlock(ParentBlockHash)
	if err != nil {
		return nil, errors.Wrap(err, "GetBlock: ")
	}
	parentTransaction := ParentBlock.Transactions[transactionhelper.CoinbaseTransactionIndex]
	txChild, err := testutils.CreateTransaction(parentTransaction, 1000)
	if err != nil {
		return nil, err
	}
	return txChild, nil
}

func contains(transaction *externalapi.DomainTransaction, transactions []*externalapi.DomainTransaction) bool {
	for _, candidateTransaction := range transactions {
		if candidateTransaction.Equal(transaction) {
			return true
		}
	}
	return false
}
