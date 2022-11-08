package miningmanager_test

import (
	"github.com/kaspanet/kaspad/cmd/kaspawallet/libkaspawallet"
	"github.com/kaspanet/kaspad/domain/consensusreference"
	"github.com/kaspanet/kaspad/domain/miningmanager/model"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/version"
	"reflect"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/domain/miningmanager/mempool"

	"github.com/kaspanet/kaspad/domain/consensus"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/model/testapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/domain/consensus/utils/testutils"
	"github.com/kaspanet/kaspad/domain/consensus/utils/transactionhelper"
	"github.com/kaspanet/kaspad/domain/consensus/utils/txscript"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/miningmanager"
	"github.com/pkg/errors"
)

// TestValidateAndInsertTransaction verifies that valid transactions were successfully inserted into the mempool.
func TestValidateAndInsertTransaction(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestValidateAndInsertTransaction")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		transactionsToInsert := make([]*externalapi.DomainTransaction, 10)
		for i := range transactionsToInsert {
			transactionsToInsert[i] = createTransactionWithUTXOEntry(t, i, 0)
			_, err = miningManager.ValidateAndInsertTransaction(transactionsToInsert[i], false, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}
		// The UTXOEntry was filled manually for those transactions, so the transactions won't be considered orphans.
		// Therefore, all the transactions expected to be contained in the mempool.
		transactionsFromMempool, _ := miningManager.AllTransactions(true, false)
		if len(transactionsToInsert) != len(transactionsFromMempool) {
			t.Fatalf("Wrong number of transactions in mempool: expected: %d, got: %d", len(transactionsToInsert), len(transactionsFromMempool))
		}
		for _, transactionToInsert := range transactionsToInsert {
			if !contains(transactionToInsert, transactionsFromMempool) {
				t.Fatalf("Missing transaction %s in the mempool", consensushashing.TransactionID(transactionToInsert))
			}
		}

		// The parent's transaction was inserted by consensus(AddBlock), and we want to verify that
		// the transaction is not considered an orphan and inserted into the mempool.
		transactionNotAnOrphan, err := createChildAndParentTxsAndAddParentToConsensus(tc)
		if err != nil {
			t.Fatalf("Error in createParentAndChildrenTransaction: %v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(transactionNotAnOrphan, false, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		transactionsFromMempool, _ = miningManager.AllTransactions(true, false)
		if !contains(transactionNotAnOrphan, transactionsFromMempool) {
			t.Fatalf("Missing transaction %s in the mempool", consensushashing.TransactionID(transactionNotAnOrphan))
		}
	})
}

func TestImmatureSpend(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestValidateAndInsertTransaction")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		tx := createTransactionWithUTXOEntry(t, 0, consensusConfig.GenesisBlock.Header.DAAScore())
		_, err = miningManager.ValidateAndInsertTransaction(tx, false, false)
		txRuleError := &mempool.TxRuleError{}
		if !errors.As(err, txRuleError) || txRuleError.RejectCode != mempool.RejectImmatureSpend {
			t.Fatalf("Unexpected error %+v", err)
		}
		transactionsFromMempool, _ := miningManager.AllTransactions(true, false)
		if contains(tx, transactionsFromMempool) {
			t.Fatalf("Mempool contains a transaction with immature coinbase")
		}
	})
}

// TestInsertDoubleTransactionsToMempool verifies that an attempt to insert a transaction
// more than once into the mempool will result in raising an appropriate error.
func TestInsertDoubleTransactionsToMempool(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestInsertDoubleTransactionsToMempool")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		transaction := createTransactionWithUTXOEntry(t, 0, 0)
		_, err = miningManager.ValidateAndInsertTransaction(transaction, false, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(transaction, false, true)
		if err == nil || !strings.Contains(err.Error(), "is already in the mempool") {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
	})
}

// TestDoubleSpendInMempool verifies that an attempt to insert a transaction double-spending
// another transaction already in  the mempool will result in raising an appropriate error.
func TestDoubleSpendInMempool(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestDoubleSpendInMempool")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		transaction, err := createChildAndParentTxsAndAddParentToConsensus(tc)
		if err != nil {
			t.Fatalf("Error creating transaction: %+v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(transaction, false, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}

		doubleSpendingTransaction := transaction.Clone()
		doubleSpendingTransaction.ID = nil
		doubleSpendingTransaction.Outputs[0].Value-- // do some minor change so that txID is different

		_, err = miningManager.ValidateAndInsertTransaction(doubleSpendingTransaction, false, true)
		if err == nil || !strings.Contains(err.Error(), "already spent by transaction") {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
	})
}

// TestHandleNewBlockTransactions verifies that all the transactions in the block were successfully removed from the mempool.
func TestHandleNewBlockTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestHandleNewBlockTransactions")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		transactionsToInsert := make([]*externalapi.DomainTransaction, 10)
		for i := range transactionsToInsert {
			transaction := createTransactionWithUTXOEntry(t, i, 0)
			transactionsToInsert[i] = transaction
			_, err = miningManager.ValidateAndInsertTransaction(transaction, false, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}

		const partialLength = 3
		blockWithFirstPartOfTheTransactions := append([]*externalapi.DomainTransaction{nil}, transactionsToInsert[0:partialLength]...)
		blockWithRestOfTheTransactions := append([]*externalapi.DomainTransaction{nil}, transactionsToInsert[partialLength:]...)
		_, err = miningManager.HandleNewBlockTransactions(blockWithFirstPartOfTheTransactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		mempoolTransactions, _ := miningManager.AllTransactions(true, false)
		for _, removedTransaction := range blockWithFirstPartOfTheTransactions {
			if contains(removedTransaction, mempoolTransactions) {
				t.Fatalf("This transaction shouldnt be in mempool: %s", consensushashing.TransactionID(removedTransaction))
			}
		}

		// There are no chained/double-spends transactions, and hence it is expected that all the other
		// transactions, will still be included in the mempool.
		mempoolTransactions, _ = miningManager.AllTransactions(true, false)
		for _, transaction := range blockWithRestOfTheTransactions[transactionhelper.CoinbaseTransactionIndex+1:] {
			if !contains(transaction, mempoolTransactions) {
				t.Fatalf("This transaction %s should be in mempool.", consensushashing.TransactionID(transaction))
			}
		}
		// Handle all the other transactions.
		_, err = miningManager.HandleNewBlockTransactions(blockWithRestOfTheTransactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		mempoolTransactions, _ = miningManager.AllTransactions(true, false)
		if len(mempoolTransactions) != 0 {
			blockIDs := domainBlocksToBlockIds(mempoolTransactions)
			t.Fatalf("The mempool contains unexpected transactions: %s", blockIDs)
		}
	})
}

func domainBlocksToBlockIds(blocks []*externalapi.DomainTransaction) []*externalapi.DomainTransactionID {
	blockIDs := make([]*externalapi.DomainTransactionID, len(blocks))
	for i := range blockIDs {
		blockIDs[i] = consensushashing.TransactionID(blocks[i])
	}
	return blockIDs
}

// TestDoubleSpendWithBlock verifies that any transactions which are now double spends as a result of the block's new transactions
// will be removed from the mempool.
func TestDoubleSpendWithBlock(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestDoubleSpendWithBlock")
		if err != nil {
			t.Fatalf("Failed setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		transactionInTheMempool := createTransactionWithUTXOEntry(t, 0, 0)
		_, err = miningManager.ValidateAndInsertTransaction(transactionInTheMempool, false, true)
		if err != nil {
			t.Fatalf("ValidateAndInsertTransaction: %v", err)
		}
		doubleSpendTransactionInTheBlock := createTransactionWithUTXOEntry(t, 0, 0)
		doubleSpendTransactionInTheBlock.Inputs[0].PreviousOutpoint = transactionInTheMempool.Inputs[0].PreviousOutpoint
		blockTransactions := []*externalapi.DomainTransaction{nil, doubleSpendTransactionInTheBlock}
		_, err = miningManager.HandleNewBlockTransactions(blockTransactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %v", err)
		}
		mempoolTransactions, _ := miningManager.AllTransactions(true, false)
		if contains(transactionInTheMempool, mempoolTransactions) {
			t.Fatalf("The transaction %s, shouldn't be in the mempool, since at least one "+
				"output was already spent.", consensushashing.TransactionID(transactionInTheMempool))
		}
	})
}

// TestOrphanTransactions verifies that a transaction could be a part of a new block template, only if it's not an orphan.
func TestOrphanTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestOrphanTransactions")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))
		// Before each parent transaction, We will add two blocks by consensus in order to fund the parent transactions.
		parentTransactions, childTransactions, err := createArraysOfParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("Error in createArraysOfParentAndChildrenTransactions: %v", err)
		}
		for _, orphanTransaction := range childTransactions {
			_, err = miningManager.ValidateAndInsertTransaction(orphanTransaction, false, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}
		transactionsMempool, _ := miningManager.AllTransactions(true, false)
		for _, transaction := range transactionsMempool {
			if contains(transaction, childTransactions) {
				t.Fatalf("Error: an orphan transaction is exist in the mempool")
			}
		}

		block, _, err := miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
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

		blockParentsTransactions, _, err := tc.GetBlock(blockParentsTransactionsHash)
		if err != nil {
			t.Fatalf("GetBlock: %v", err)
		}
		_, err = miningManager.HandleNewBlockTransactions(blockParentsTransactions.Transactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %+v", err)
		}
		transactionsMempool, _ = miningManager.AllTransactions(true, false)
		if len(transactionsMempool) != len(childTransactions) {
			t.Fatalf("Expected %d transactions in the mempool but got %d", len(childTransactions), len(transactionsMempool))
		}

		for _, transaction := range transactionsMempool {
			if !contains(transaction, childTransactions) {
				t.Fatalf("Error: the transaction %s, should be in the mempool since its not "+
					"oprhan anymore.", consensushashing.TransactionID(transaction))
			}
		}
		block, _, err = miningManager.GetBlockTemplate(&externalapi.DomainCoinbaseData{
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
				t.Fatalf("Error: Unknown Transaction %s in a block.", consensushashing.TransactionID(transactionFromBlock))
			}
		}
	})
}

func TestHighPriorityTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestDoubleSpendWithBlock")
		if err != nil {
			t.Fatalf("Failed setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		mempoolConfig := mempool.DefaultConfig(&consensusConfig.Params)
		mempoolConfig.MaximumTransactionCount = 1
		mempoolConfig.MaximumOrphanTransactionCount = 1
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempoolConfig)

		// Create 3 pairs of transaction parent-and-child pairs: 1 low priority and 2 high priority
		lowPriorityParentTransaction, lowPriorityChildTransaction, err := createParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("error creating low-priority transaction pair: %+v", err)
		}
		firstHighPriorityParentTransaction, firstHighPriorityChildTransaction, err := createParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("error creating first high-priority transaction pair: %+v", err)
		}
		secondHighPriorityParentTransaction, secondHighPriorityChildTransaction, err := createParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("error creating second high-priority transaction pair: %+v", err)
		}

		// Submit all the children, make sure the 2 highPriority ones remain in the orphan pool
		_, err = miningManager.ValidateAndInsertTransaction(lowPriorityChildTransaction, false, true)
		if err != nil {
			t.Fatalf("error submitting low-priority transaction: %+v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(firstHighPriorityChildTransaction, true, true)
		if err != nil {
			t.Fatalf("error submitting first high-priority transaction: %+v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(secondHighPriorityChildTransaction, true, true)
		if err != nil {
			t.Fatalf("error submitting second high-priority transaction: %+v", err)
		}
		// There's no API to check what stayed in the orphan pool, but we'll find it out when we begin to unorphan

		// Submit all the parents.
		// Low priority transaction will only accept the parent, since the child was evicted from orphanPool
		lowPriorityAcceptedTransactions, err :=
			miningManager.ValidateAndInsertTransaction(lowPriorityParentTransaction, false, true)
		if err != nil {
			t.Fatalf("error submitting low-priority transaction: %+v", err)
		}
		expectedLowPriorityAcceptedTransactions := []*externalapi.DomainTransaction{lowPriorityParentTransaction}
		if !reflect.DeepEqual(lowPriorityAcceptedTransactions, expectedLowPriorityAcceptedTransactions) {
			t.Errorf("Expected only lowPriorityParent (%v) to be in lowPriorityAcceptedTransactions, but got %v",
				consensushashing.TransactionIDs(expectedLowPriorityAcceptedTransactions),
				consensushashing.TransactionIDs(lowPriorityAcceptedTransactions))
		}

		// Both high priority transactions should accept parent and child

		// Insert firstHighPriorityParentTransaction
		firstHighPriorityAcceptedTransactions, err :=
			miningManager.ValidateAndInsertTransaction(firstHighPriorityParentTransaction, true, true)
		if err != nil {
			t.Fatalf("error submitting first high-priority transaction: %+v", err)
		}
		expectedFirstHighPriorityAcceptedTransactions :=
			[]*externalapi.DomainTransaction{firstHighPriorityParentTransaction, firstHighPriorityChildTransaction}
		if !reflect.DeepEqual(firstHighPriorityAcceptedTransactions, expectedFirstHighPriorityAcceptedTransactions) {
			t.Errorf(
				"Expected both firstHighPriority transaction (%v) to be in firstHighPriorityAcceptedTransactions, but got %v",
				consensushashing.TransactionIDs(firstHighPriorityAcceptedTransactions),
				consensushashing.TransactionIDs(expectedFirstHighPriorityAcceptedTransactions))
		}
		// Insert secondHighPriorityParentTransaction
		secondHighPriorityAcceptedTransactions, err :=
			miningManager.ValidateAndInsertTransaction(secondHighPriorityParentTransaction, true, true)
		if err != nil {
			t.Fatalf("error submitting second high-priority transaction: %+v", err)
		}
		expectedSecondHighPriorityAcceptedTransactions :=
			[]*externalapi.DomainTransaction{secondHighPriorityParentTransaction, secondHighPriorityChildTransaction}
		if !reflect.DeepEqual(secondHighPriorityAcceptedTransactions, expectedSecondHighPriorityAcceptedTransactions) {
			t.Errorf(
				"Expected both secondHighPriority transaction (%v) to be in secondHighPriorityAcceptedTransactions, but got %v",
				consensushashing.TransactionIDs(secondHighPriorityAcceptedTransactions),
				consensushashing.TransactionIDs(expectedSecondHighPriorityAcceptedTransactions))
		}
	})
}

func TestRevalidateHighPriorityTransactions(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestRevalidateHighPriorityTransactions")
		if err != nil {
			t.Fatalf("Failed setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		mempoolConfig := mempool.DefaultConfig(&consensusConfig.Params)
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempoolConfig)

		// Create two valid transactions that double-spend each other (childTransaction1, childTransaction2)
		parentTransaction, childTransaction1, err := createParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("Error creating parentTransaction and childTransaction1: %+v", err)
		}
		tips, err := tc.Tips()
		if err != nil {
			t.Fatalf("Error getting tips: %+v", err)
		}

		fundingBlock, _, err := tc.AddBlock(tips, nil, []*externalapi.DomainTransaction{parentTransaction})
		if err != nil {
			t.Fatalf("Error getting function block: %+v", err)
		}

		childTransaction2 := childTransaction1.Clone()
		childTransaction2.Outputs[0].Value-- // decrement value to change id

		// Mine 1 block with confirming childTransaction1 and 2 blocks confirming childTransaction2, so that
		// childTransaction2 is accepted
		tip1, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock}, nil,
			[]*externalapi.DomainTransaction{childTransaction1})
		if err != nil {
			t.Fatalf("Error adding tip1: %+v", err)
		}
		tip2, _, err := tc.AddBlock([]*externalapi.DomainHash{fundingBlock}, nil,
			[]*externalapi.DomainTransaction{childTransaction2})
		if err != nil {
			t.Fatalf("Error adding tip2: %+v", err)
		}
		_, _, err = tc.AddBlock([]*externalapi.DomainHash{tip2}, nil, nil)
		if err != nil {
			t.Fatalf("Error mining on top of tip2: %+v", err)
		}

		// Add to mempool transaction that spends childTransaction2 (as high priority)
		spendingTransaction, err := testutils.CreateTransaction(childTransaction2, 1000)
		if err != nil {
			t.Fatalf("Error creating spendingTransaction: %+v", err)
		}
		_, err = miningManager.ValidateAndInsertTransaction(spendingTransaction, true, false)
		if err != nil {
			t.Fatalf("Error inserting spendingTransaction: %+v", err)
		}

		// Revalidate, to make sure spendingTransaction is still valid
		validTransactions, err := miningManager.RevalidateHighPriorityTransactions()
		if err != nil {
			t.Fatalf("Error from first RevalidateHighPriorityTransactions: %+v", err)
		}
		if len(validTransactions) != 1 || !validTransactions[0].Equal(spendingTransaction) {
			t.Fatalf("Expected to have spendingTransaction as only validTransaction returned from "+
				"RevalidateHighPriorityTransactions, but got %v instead", validTransactions)
		}

		// Mine 2 more blocks on top of tip1, to re-org out childTransaction1, thus making spendingTransaction invalid
		for i := 0; i < 2; i++ {
			tip1, _, err = tc.AddBlock([]*externalapi.DomainHash{tip1}, nil, nil)
			if err != nil {
				t.Fatalf("Error mining on top of tip1: %+v", err)
			}
		}

		// Make sure spendingTransaction is still in mempool
		mempoolTransactions, _ := miningManager.AllTransactions(true, false)
		if len(mempoolTransactions) != 1 || !mempoolTransactions[0].Equal(spendingTransaction) {
			t.Fatalf("Expected to have spendingTransaction as only validTransaction returned from "+
				"RevalidateHighPriorityTransactions, but got %v instead", validTransactions)
		}

		// Revalidate again, this time validTransactions should be empty
		validTransactions, err = miningManager.RevalidateHighPriorityTransactions()
		if err != nil {
			t.Fatalf("Error from first RevalidateHighPriorityTransactions: %+v", err)
		}
		if len(validTransactions) != 0 {
			t.Fatalf("Expected to have empty validTransactions, but got %v instead", validTransactions)
		}
		// And also AllTransactions should be empty as well
		mempoolTransactions, _ = miningManager.AllTransactions(true, false)
		if len(mempoolTransactions) != 0 {
			t.Fatalf("Expected to have empty allTransactions, but got %v instead", mempoolTransactions)
		}
	})
}

// TestModifyBlockTemplate verifies that modifying a block template changes coinbase data correctly.
func TestModifyBlockTemplate(t *testing.T) {
	testutils.ForAllNets(t, true, func(t *testing.T, consensusConfig *consensus.Config) {
		consensusConfig.BlockCoinbaseMaturity = 0
		factory := consensus.NewFactory()
		tc, teardown, err := factory.NewTestConsensus(consensusConfig, "TestModifyBlockTemplate")
		if err != nil {
			t.Fatalf("Error setting up TestConsensus: %+v", err)
		}
		defer teardown(false)

		miningFactory := miningmanager.NewFactory()
		tcAsConsensus := tc.(externalapi.Consensus)
		tcAsConsensusPointer := &tcAsConsensus
		consensusReference := consensusreference.NewConsensusReference(&tcAsConsensusPointer)
		miningManager := miningFactory.NewMiningManager(consensusReference, &consensusConfig.Params, mempool.DefaultConfig(&consensusConfig.Params))

		// Create some complex transactions. Logic taken from TestOrphanTransactions

		// Before each parent transaction, We will add two blocks by consensus in order to fund the parent transactions.
		parentTransactions, childTransactions, err := createArraysOfParentAndChildrenTransactions(tc)
		if err != nil {
			t.Fatalf("Error in createArraysOfParentAndChildrenTransactions: %v", err)
		}
		for _, orphanTransaction := range childTransactions {
			_, err = miningManager.ValidateAndInsertTransaction(orphanTransaction, false, true)
			if err != nil {
				t.Fatalf("ValidateAndInsertTransaction: %v", err)
			}
		}
		transactionsMempool, _ := miningManager.AllTransactions(true, false)
		for _, transaction := range transactionsMempool {
			if contains(transaction, childTransactions) {
				t.Fatalf("Error: an orphan transaction is exist in the mempool")
			}
		}

		emptyCoinbaseData := &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0},
			ExtraData:       nil}
		block, _, err := miningManager.GetBlockTemplate(emptyCoinbaseData)
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

		// Run the purpose of this test, compare modified block templates
		sweepCompareModifiedTemplateToBuilt(t, consensusConfig, miningManager.GetBlockTemplateBuilder())

		// Create some more complex blocks and transactions. Logic taken from TestOrphanTransactions
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

		blockParentsTransactions, _, err := tc.GetBlock(blockParentsTransactionsHash)
		if err != nil {
			t.Fatalf("GetBlock: %v", err)
		}
		_, err = miningManager.HandleNewBlockTransactions(blockParentsTransactions.Transactions)
		if err != nil {
			t.Fatalf("HandleNewBlockTransactions: %+v", err)
		}
		transactionsMempool, _ = miningManager.AllTransactions(true, false)
		if len(transactionsMempool) != len(childTransactions) {
			t.Fatalf("Expected %d transactions in the mempool but got %d", len(childTransactions), len(transactionsMempool))
		}

		for _, transaction := range transactionsMempool {
			if !contains(transaction, childTransactions) {
				t.Fatalf("Error: the transaction %s, should be in the mempool since its not "+
					"oprhan anymore.", consensushashing.TransactionID(transaction))
			}
		}
		block, _, err = miningManager.GetBlockTemplate(emptyCoinbaseData)
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
				t.Fatalf("Error: Unknown Transaction %s in a block.", consensushashing.TransactionID(transactionFromBlock))
			}
		}

		// Run the purpose of this test, compare modified block templates
		sweepCompareModifiedTemplateToBuilt(t, consensusConfig, miningManager.GetBlockTemplateBuilder())

		// Create a real coinbase to use
		coinbaseUsual, err := generateNewCoinbase(consensusConfig.Prefix, opUsual)
		if err != nil {
			t.Fatalf("Generate coinbase: %v.", err)
		}
		var emptyTransactions []*externalapi.DomainTransaction

		// Create interesting DAG structures and rerun the template comparisons
		tips, err = tc.Tips()
		if err != nil {
			t.Fatalf("Tips: %v.", err)
		}
		// Create a fork
		_, _, err = tc.AddBlock(tips[:1], coinbaseUsual, emptyTransactions)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}
		chainTip, _, err := tc.AddBlock(tips[:1], coinbaseUsual, emptyTransactions)
		if err != nil {
			t.Fatalf("AddBlock: %v", err)
		}

		sweepCompareModifiedTemplateToBuilt(t, consensusConfig, miningManager.GetBlockTemplateBuilder())

		// Create some blue blocks
		for i := externalapi.KType(0); i < consensusConfig.K-2; i++ {
			chainTip, _, err = tc.AddBlock([]*externalapi.DomainHash{chainTip}, coinbaseUsual, emptyTransactions)
			if err != nil {
				t.Fatalf("AddBlock: %v", err)
			}
		}

		sweepCompareModifiedTemplateToBuilt(t, consensusConfig, miningManager.GetBlockTemplateBuilder())

		// Mine more such that we have a merged red
		for i := externalapi.KType(0); i < consensusConfig.K; i++ {
			chainTip, _, err = tc.AddBlock([]*externalapi.DomainHash{chainTip}, coinbaseUsual, emptyTransactions)
			if err != nil {
				t.Fatalf("AddBlock: %v", err)
			}
		}
		blockTemplate, err := miningManager.GetBlockTemplateBuilder().BuildBlockTemplate(emptyCoinbaseData)
		if err != nil {
			t.Fatalf("BuildBlockTemplate: %v", err)
		}
		if !blockTemplate.CoinbaseHasRedReward {
			t.Fatalf("Expected block template to have red reward")
		}

		sweepCompareModifiedTemplateToBuilt(t, consensusConfig, miningManager.GetBlockTemplateBuilder())
	})
}

func sweepCompareModifiedTemplateToBuilt(
	t *testing.T, consensusConfig *consensus.Config, builder model.BlockTemplateBuilder) {
	for i := 0; i < 4; i++ {
		// Run a few times to get more randomness
		compareModifiedTemplateToBuilt(t, consensusConfig, builder, opUsual, opUsual)
		compareModifiedTemplateToBuilt(t, consensusConfig, builder, opECDSA, opECDSA)
	}
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opTrue, opUsual)
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opUsual, opTrue)
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opECDSA, opUsual)
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opUsual, opECDSA)
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opEmpty, opUsual)
	compareModifiedTemplateToBuilt(t, consensusConfig, builder, opUsual, opEmpty)
}

type opType uint8

const (
	opUsual opType = iota
	opECDSA
	opTrue
	opEmpty
)

func compareModifiedTemplateToBuilt(
	t *testing.T, consensusConfig *consensus.Config, builder model.BlockTemplateBuilder,
	firstCoinbaseOp, secondCoinbaseOp opType) {
	coinbase1, err := generateNewCoinbase(consensusConfig.Params.Prefix, firstCoinbaseOp)
	if err != nil {
		t.Fatalf("Failed to generate new coinbase: %v", err)
	}
	coinbase2, err := generateNewCoinbase(consensusConfig.Params.Prefix, secondCoinbaseOp)
	if err != nil {
		t.Fatalf("Failed to generate new coinbase: %v", err)
	}

	// Build a fresh template for coinbase2 as a reference
	expectedTemplate, err := builder.BuildBlockTemplate(coinbase2)
	if err != nil {
		t.Fatalf("Failed to build block template: %v", err)
	}
	// Modify to coinbase1
	modifiedTemplate, err := builder.ModifyBlockTemplate(coinbase1, expectedTemplate.Clone())
	if err != nil {
		t.Fatalf("Failed to modify block template: %v", err)
	}
	// And modify back to coinbase2
	modifiedTemplate, err = builder.ModifyBlockTemplate(coinbase2, modifiedTemplate.Clone())
	if err != nil {
		t.Fatalf("Failed to modify block template: %v", err)
	}

	// Make sure timestamps are equal before comparing the hash
	mutableHeader := modifiedTemplate.Block.Header.ToMutable()
	mutableHeader.SetTimeInMilliseconds(expectedTemplate.Block.Header.TimeInMilliseconds())
	modifiedTemplate.Block.Header = mutableHeader.ToImmutable()

	// Assert hashes are equal
	expectedTemplateHash := consensushashing.BlockHash(expectedTemplate.Block)
	modifiedTemplateHash := consensushashing.BlockHash(modifiedTemplate.Block)
	if !expectedTemplateHash.Equal(modifiedTemplateHash) {
		t.Fatalf("Expected block hashes %s, %s to be equal", expectedTemplateHash, modifiedTemplateHash)
	}
}

func generateNewCoinbase(addressPrefix util.Bech32Prefix, op opType) (*externalapi.DomainCoinbaseData, error) {
	if op == opTrue {
		scriptPublicKey, _ := testutils.OpTrueScript()
		return &externalapi.DomainCoinbaseData{
			ScriptPublicKey: scriptPublicKey, ExtraData: []byte(version.Version()),
		}, nil
	}
	if op == opEmpty {
		return &externalapi.DomainCoinbaseData{
			ScriptPublicKey: &externalapi.ScriptPublicKey{Script: nil, Version: 0},
			ExtraData:       nil,
		}, nil
	}
	_, publicKey, err := libkaspawallet.CreateKeyPair(op == opECDSA)
	if err != nil {
		return nil, err
	}
	var address string
	if op == opECDSA {
		addressPublicKeyECDSA, err := util.NewAddressPublicKeyECDSA(publicKey, addressPrefix)
		if err != nil {
			return nil, err
		}
		address = addressPublicKeyECDSA.EncodeAddress()
	} else {
		addressPublicKey, err := util.NewAddressPublicKey(publicKey, addressPrefix)
		if err != nil {
			return nil, err
		}
		address = addressPublicKey.EncodeAddress()
	}
	payAddress, err := util.DecodeAddress(address, addressPrefix)
	if err != nil {
		return nil, err
	}
	scriptPublicKey, err := txscript.PayToAddrScript(payAddress)
	if err != nil {
		return nil, err
	}
	return &externalapi.DomainCoinbaseData{
		ScriptPublicKey: scriptPublicKey, ExtraData: []byte(version.Version()),
	}, nil
}

func createTransactionWithUTXOEntry(t *testing.T, i int, daaScore uint64) *externalapi.DomainTransaction {
	prevOutTxID := externalapi.DomainTransactionID{}
	prevOutPoint := externalapi.DomainOutpoint{TransactionID: prevOutTxID, Index: uint32(i)}
	scriptPublicKey, redeemScript := testutils.OpTrueScript()
	signatureScript, err := txscript.PayToScriptHashSignatureScript(redeemScript, nil)
	if err != nil {
		t.Fatalf("PayToScriptHashSignatureScript: %v", err)
	}
	txInput := externalapi.DomainTransactionInput{
		PreviousOutpoint: prevOutPoint,
		SignatureScript:  signatureScript,
		Sequence:         constants.MaxTxInSequenceNum,
		UTXOEntry: utxo.NewUTXOEntry(
			100000000, // 1 KAS
			scriptPublicKey,
			true,
			daaScore),
	}
	txOut := externalapi.DomainTransactionOutput{
		Value:           10000,
		ScriptPublicKey: scriptPublicKey,
	}
	tx := externalapi.DomainTransaction{
		Version:      constants.MaxTransactionVersion,
		Inputs:       []*externalapi.DomainTransactionInput{&txInput},
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

func createParentAndChildrenTransactions(tc testapi.TestConsensus) (txParent *externalapi.DomainTransaction,
	txChild *externalapi.DomainTransaction, err error) {

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
	fundingBlockForParent, _, err := tc.GetBlock(fundingBlockHashForParent)
	if err != nil {
		return nil, nil, errors.Wrap(err, "GetBlock: ")
	}
	fundingTransactionForParent := fundingBlockForParent.Transactions[transactionhelper.CoinbaseTransactionIndex]
	txParent, err = testutils.CreateTransaction(fundingTransactionForParent, 1000)
	if err != nil {
		return nil, nil, err
	}
	txChild, err = testutils.CreateTransaction(txParent, 1000)
	if err != nil {
		return nil, nil, err
	}
	return txParent, txChild, nil
}

func createChildAndParentTxsAndAddParentToConsensus(tc testapi.TestConsensus) (*externalapi.DomainTransaction, error) {
	firstBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{tc.DAGParams().GenesisHash}, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "AddBlock: %v", err)
	}
	ParentBlockHash, _, err := tc.AddBlock([]*externalapi.DomainHash{firstBlockHash}, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "AddBlock: ")
	}
	ParentBlock, _, err := tc.GetBlock(ParentBlockHash)
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
