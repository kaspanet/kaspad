package blockdag_test

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util/subnetworkid"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/testtools"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/util"
)

// TestFinalityInterval tests that the finality interval is
// smaller then appmessage.MaxInvPerMsg, so when a peer receives
// a getblocks message it should always be able to send
// all the necessary invs.
func TestFinalityInterval(t *testing.T) {
	netParams := []*dagconfig.Params{
		&dagconfig.MainnetParams,
		&dagconfig.TestnetParams,
		&dagconfig.DevnetParams,
		&dagconfig.SimnetParams,
	}
	for _, params := range netParams {
		func() {
			dag, teardownFunc, err := blockdag.DAGSetup("TestFinalityInterval", true, blockdag.Config{
				DAGParams: params,
			})
			if err != nil {
				t.Fatalf("Failed to setup dag instance for %s: %v", params.Name, err)
			}
			defer teardownFunc()

			if dag.FinalityInterval() > appmessage.MaxInvPerMsg {
				t.Errorf("FinalityInterval in %s should be lower or equal to appmessage.MaxInvPerMsg", params.Name)
			}
		}()
	}
}

// TestSubnetworkRegistry tests the full subnetwork registry flow
func TestSubnetworkRegistry(t *testing.T) {
	params := dagconfig.SimnetParams
	params.K = 1
	params.BlockCoinbaseMaturity = 0
	params.EnableNonNativeSubnetworks = true
	dag, teardownFunc, err := blockdag.DAGSetup("TestSubnetworkRegistry", true, blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	gasLimit := uint64(12345)
	subnetworkID, err := testtools.RegisterSubnetworkForTest(dag, &params, gasLimit)
	if err != nil {
		t.Fatalf("could not register network: %s", err)
	}
	limit, err := dag.GasLimit(subnetworkID)
	if err != nil {
		t.Fatalf("could not retrieve gas limit: %s", err)
	}
	if limit != gasLimit {
		t.Fatalf("unexpected gas limit. want: %d, got: %d", gasLimit, limit)
	}
}

func TestChainedTransactions(t *testing.T) {
	params := dagconfig.SimnetParams
	params.BlockCoinbaseMaturity = 0
	// Create a new database and dag instance to run tests against.
	dag, teardownFunc, err := blockdag.DAGSetup("TestChainedTransactions", true, blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup dag instance: %v", err)
	}
	defer teardownFunc()

	block1, err := blockdag.PrepareBlockForTest(dag, []*daghash.Hash{params.GenesisHash}, nil)
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}
	isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(block1), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock: %v", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: block1 " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: block1 got unexpectedly orphaned")
	}
	cbTx := block1.Transactions[0]

	signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
	if err != nil {
		t.Fatalf("Failed to build signature script: %s", err)
	}
	txIn := &appmessage.TxIn{
		PreviousOutpoint: appmessage.Outpoint{TxID: *cbTx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         appmessage.MaxTxInSequenceNum,
	}
	txOut := &appmessage.TxOut{
		ScriptPubKey: blockdag.OpTrueScript,
		Value:        uint64(1),
	}
	tx := appmessage.NewNativeMsgTx(appmessage.TxVersion, []*appmessage.TxIn{txIn}, []*appmessage.TxOut{txOut})

	chainedTxIn := &appmessage.TxIn{
		PreviousOutpoint: appmessage.Outpoint{TxID: *tx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         appmessage.MaxTxInSequenceNum,
	}

	scriptPubKey, err := txscript.PayToScriptHashScript(blockdag.OpTrueScript)
	if err != nil {
		t.Fatalf("Failed to build public key script: %s", err)
	}
	chainedTxOut := &appmessage.TxOut{
		ScriptPubKey: scriptPubKey,
		Value:        uint64(1),
	}
	chainedTx := appmessage.NewNativeMsgTx(appmessage.TxVersion, []*appmessage.TxIn{chainedTxIn}, []*appmessage.TxOut{chainedTxOut})

	block2, err := blockdag.PrepareBlockForTest(dag, []*daghash.Hash{block1.BlockHash()}, []*appmessage.MsgTx{tx})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Manually add a chained transaction to block2
	block2.Transactions = append(block2.Transactions, chainedTx)
	block2UtilTxs := make([]*util.Tx, len(block2.Transactions))
	for i, tx := range block2.Transactions {
		block2UtilTxs[i] = util.NewTx(tx)
	}
	block2.Header.HashMerkleRoot = blockdag.BuildHashMerkleTreeStore(block2UtilTxs).Root()

	//Checks that dag.ProcessBlock fails because we don't allow a transaction to spend another transaction from the same block
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(block2), blockdag.BFNoPoWCheck)
	if err == nil {
		t.Errorf("ProcessBlock expected an error")
	} else {
		var ruleErr blockdag.RuleError
		if ok := errors.As(err, &ruleErr); ok {
			if ruleErr.ErrorCode != blockdag.ErrMissingTxOut {
				t.Errorf("ProcessBlock expected an %v error code but got %v", blockdag.ErrMissingTxOut, ruleErr.ErrorCode)
			}
		} else {
			t.Errorf("ProcessBlock expected a blockdag.RuleError but got %v", err)
		}
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: block2 " +
			"is too far in the future")
	}
	if isOrphan {
		t.Errorf("ProcessBlock: block2 got unexpectedly orphaned")
	}

	nonChainedTxIn := &appmessage.TxIn{
		PreviousOutpoint: appmessage.Outpoint{TxID: *cbTx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         appmessage.MaxTxInSequenceNum,
	}
	nonChainedTxOut := &appmessage.TxOut{
		ScriptPubKey: scriptPubKey,
		Value:        uint64(1),
	}
	nonChainedTx := appmessage.NewNativeMsgTx(appmessage.TxVersion, []*appmessage.TxIn{nonChainedTxIn}, []*appmessage.TxOut{nonChainedTxOut})

	block3, err := blockdag.PrepareBlockForTest(dag, []*daghash.Hash{block1.BlockHash()}, []*appmessage.MsgTx{nonChainedTx})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	//Checks that dag.ProcessBlock doesn't fail because all of its transaction are dependant on transactions from previous blocks
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(block3), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Errorf("ProcessBlock: %v", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: block3 " +
			"is too far in the future")
	}
	if isOrphan {
		t.Errorf("ProcessBlock: block3 got unexpectedly orphaned")
	}
}

// TestOrderInDiffFromAcceptanceData makes sure that the order of transactions in
// dag.diffFromAcceptanceData is such that if txA is spent by txB then txA is processed
// before txB.
func TestOrderInDiffFromAcceptanceData(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	params := dagconfig.SimnetParams
	params.K = math.MaxUint8
	dag, teardownFunc, err := blockdag.DAGSetup("TestOrderInDiffFromAcceptanceData", true, blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	dag.TestSetCoinbaseMaturity(0)

	createBlock := func(previousBlock *util.Block) *util.Block {
		// Prepare a transaction that spends the previous block's coinbase transaction
		var txs []*appmessage.MsgTx
		if !previousBlock.IsGenesis() {
			previousCoinbaseTx := previousBlock.MsgBlock().Transactions[0]
			signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
			if err != nil {
				t.Fatalf("TestOrderInDiffFromAcceptanceData: Failed to build signature script: %s", err)
			}
			txIn := &appmessage.TxIn{
				PreviousOutpoint: appmessage.Outpoint{TxID: *previousCoinbaseTx.TxID(), Index: 0},
				SignatureScript:  signatureScript,
				Sequence:         appmessage.MaxTxInSequenceNum,
			}
			txOut := &appmessage.TxOut{
				ScriptPubKey: blockdag.OpTrueScript,
				Value:        uint64(1),
			}
			txs = append(txs, appmessage.NewNativeMsgTx(appmessage.TxVersion, []*appmessage.TxIn{txIn}, []*appmessage.TxOut{txOut}))
		}

		// Create the block
		msgBlock, err := blockdag.PrepareBlockForTest(dag, []*daghash.Hash{previousBlock.Hash()}, txs)
		if err != nil {
			t.Fatalf("TestOrderInDiffFromAcceptanceData: Failed to prepare block: %+v", err)
		}

		// Add the block to the DAG
		newBlock := util.NewBlock(msgBlock)
		isOrphan, isDelayed, err := dag.ProcessBlock(newBlock, blockdag.BFNoPoWCheck)
		if err != nil {
			t.Errorf("TestOrderInDiffFromAcceptanceData: %s", err)
		}
		if isDelayed {
			t.Fatalf("TestOrderInDiffFromAcceptanceData: block is too far in the future")
		}
		if isOrphan {
			t.Fatalf("TestOrderInDiffFromAcceptanceData: block got unexpectedly orphaned")
		}
		return newBlock
	}

	// Create two block chains starting from the genesis block. Every time a block is added
	// one of the chains is selected as the selected parent chain while all the blocks in
	// the other chain (and their transactions) get accepted by the new virtual. If the
	// transactions in the non-selected parent chain get processed in the wrong order then
	// diffFromAcceptanceData panics.
	blockAmountPerChain := 100
	chainATip := util.NewBlock(params.GenesisBlock)
	chainBTip := chainATip
	for i := 0; i < blockAmountPerChain; i++ {
		chainATip = createBlock(chainATip)
		chainBTip = createBlock(chainBTip)
	}
}

// TestGasLimit tests the gas limit rules
func TestGasLimit(t *testing.T) {
	params := dagconfig.SimnetParams
	params.K = 1
	params.BlockCoinbaseMaturity = 0
	params.EnableNonNativeSubnetworks = true
	dag, teardownFunc, err := blockdag.DAGSetup("TestSubnetworkRegistry", true, blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// First we prepare a subnetwork and a block with coinbase outputs to fund our tests
	gasLimit := uint64(12345)
	subnetworkID, err := testtools.RegisterSubnetworkForTest(dag, &params, gasLimit)
	if err != nil {
		t.Fatalf("could not register network: %s", err)
	}

	cbTxs := []*appmessage.MsgTx{}
	for i := 0; i < 4; i++ {
		fundsBlock, err := blockdag.PrepareBlockForTest(dag, dag.VirtualParentHashes(), nil)
		if err != nil {
			t.Fatalf("PrepareBlockForTest: %v", err)
		}
		isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(fundsBlock), blockdag.BFNoPoWCheck)
		if err != nil {
			t.Fatalf("ProcessBlock: %v", err)
		}
		if isDelayed {
			t.Fatalf("ProcessBlock: the funds block " +
				"is too far in the future")
		}
		if isOrphan {
			t.Fatalf("ProcessBlock: fundsBlock got unexpectedly orphan")
		}

		cbTxs = append(cbTxs, fundsBlock.Transactions[util.CoinbaseTransactionIndex])
	}

	signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
	if err != nil {
		t.Fatalf("Failed to build signature script: %s", err)
	}

	scriptPubKey, err := txscript.PayToScriptHashScript(blockdag.OpTrueScript)
	if err != nil {
		t.Fatalf("Failed to build public key script: %s", err)
	}

	tx1In := &appmessage.TxIn{
		PreviousOutpoint: *appmessage.NewOutpoint(cbTxs[0].TxID(), 0),
		Sequence:         appmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	tx1Out := &appmessage.TxOut{
		Value:        cbTxs[0].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	tx1 := appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, []*appmessage.TxIn{tx1In}, []*appmessage.TxOut{tx1Out}, subnetworkID, 10000, []byte{})

	tx2In := &appmessage.TxIn{
		PreviousOutpoint: *appmessage.NewOutpoint(cbTxs[1].TxID(), 0),
		Sequence:         appmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	tx2Out := &appmessage.TxOut{
		Value:        cbTxs[1].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	tx2 := appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, []*appmessage.TxIn{tx2In}, []*appmessage.TxOut{tx2Out}, subnetworkID, 10000, []byte{})

	// Here we check that we can't process a block that has transactions that exceed the gas limit
	overLimitBlock, err := blockdag.PrepareBlockForTest(dag, dag.VirtualParentHashes(), []*appmessage.MsgTx{tx1, tx2})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}
	isOrphan, isDelayed, err := dag.ProcessBlock(util.NewBlock(overLimitBlock), blockdag.BFNoPoWCheck)
	if err == nil {
		t.Fatalf("ProcessBlock expected to have an error in block that exceeds gas limit")
	}
	var ruleErr blockdag.RuleError
	if !errors.As(err, &ruleErr) {
		t.Fatalf("ProcessBlock expected a RuleError, but got %v", err)
	} else if ruleErr.ErrorCode != blockdag.ErrInvalidGas {
		t.Fatalf("ProcessBlock expected error code %s but got %s", blockdag.ErrInvalidGas, ruleErr.ErrorCode)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: overLimitBlock " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: overLimitBlock got unexpectedly orphan")
	}

	overflowGasTxIn := &appmessage.TxIn{
		PreviousOutpoint: *appmessage.NewOutpoint(cbTxs[2].TxID(), 0),
		Sequence:         appmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	overflowGasTxOut := &appmessage.TxOut{
		Value:        cbTxs[2].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	overflowGasTx := appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, []*appmessage.TxIn{overflowGasTxIn}, []*appmessage.TxOut{overflowGasTxOut},
		subnetworkID, math.MaxUint64, []byte{})

	// Here we check that we can't process a block that its transactions' gas overflows uint64
	overflowGasBlock, err := blockdag.PrepareBlockForTest(dag, dag.VirtualParentHashes(), []*appmessage.MsgTx{tx1, overflowGasTx})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(overflowGasBlock), blockdag.BFNoPoWCheck)
	if err == nil {
		t.Fatalf("ProcessBlock expected to have an error")
	}
	if !errors.As(err, &ruleErr) {
		t.Fatalf("ProcessBlock expected a RuleError, but got %v", err)
	} else if ruleErr.ErrorCode != blockdag.ErrInvalidGas {
		t.Fatalf("ProcessBlock expected error code %s but got %s", blockdag.ErrInvalidGas, ruleErr.ErrorCode)
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: overLimitBlock got unexpectedly orphan")
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: overflowGasBlock " +
			"is too far in the future")
	}

	nonExistentSubnetwork := &subnetworkid.SubnetworkID{123}
	nonExistentSubnetworkTxIn := &appmessage.TxIn{
		PreviousOutpoint: *appmessage.NewOutpoint(cbTxs[3].TxID(), 0),
		Sequence:         appmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	nonExistentSubnetworkTxOut := &appmessage.TxOut{
		Value:        cbTxs[3].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	nonExistentSubnetworkTx := appmessage.NewSubnetworkMsgTx(appmessage.TxVersion, []*appmessage.TxIn{nonExistentSubnetworkTxIn},
		[]*appmessage.TxOut{nonExistentSubnetworkTxOut}, nonExistentSubnetwork, 1, []byte{})

	nonExistentSubnetworkBlock, err := blockdag.PrepareBlockForTest(dag, dag.VirtualParentHashes(), []*appmessage.MsgTx{nonExistentSubnetworkTx, overflowGasTx})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}

	// Here we check that we can't process a block with a transaction from a non-existent subnetwork
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(nonExistentSubnetworkBlock), blockdag.BFNoPoWCheck)
	expectedErrStr := fmt.Sprintf("Error getting gas limit for subnetworkID '%s': subnetwork '%s' not found",
		nonExistentSubnetwork, nonExistentSubnetwork)
	if strings.Contains(err.Error(), expectedErrStr) {
		t.Fatalf("ProcessBlock expected error \"%v\" but got \"%v\"", expectedErrStr, err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: nonExistentSubnetworkBlock " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: nonExistentSubnetworkBlock got unexpectedly orphan")
	}

	// Here we check that we can process a block with a transaction that doesn't exceed the gas limit
	validBlock, err := blockdag.PrepareBlockForTest(dag, dag.VirtualParentHashes(), []*appmessage.MsgTx{tx1})
	if err != nil {
		t.Fatalf("PrepareBlockForTest: %v", err)
	}
	isOrphan, isDelayed, err = dag.ProcessBlock(util.NewBlock(validBlock), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock: %v", err)
	}
	if isDelayed {
		t.Fatalf("ProcessBlock: overLimitBlock " +
			"is too far in the future")
	}
	if isOrphan {
		t.Fatalf("ProcessBlock: overLimitBlock got unexpectedly orphan")
	}
}
