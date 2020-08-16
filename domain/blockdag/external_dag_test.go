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

	"github.com/kaspanet/kaspad/domain/blockdag"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/domain/mining"
	"github.com/kaspanet/kaspad/domain/txscript"
	"github.com/kaspanet/kaspad/network/domainmessage"
	"github.com/kaspanet/kaspad/util"
)

// TestFinality checks that the finality mechanism works as expected.
// This is how the flow goes:
// 1) We build a chain of params.FinalityInterval blocks and call its tip altChainTip.
// 2) We build another chain (let's call it mainChain) of 2 * params.FinalityInterval
// blocks, which points to genesis, and then we check that the block in that
// chain with height of params.FinalityInterval is marked as finality point (This is
// very predictable, because the blue score of each new block in a chain is the
// parents plus one).
// 3) We make a new child to block with height (2 * params.FinalityInterval - 1)
// in mainChain, and we check that connecting it to the DAG
// doesn't affect the last finality point.
// 4) We make a block that points to genesis, and check that it
// gets rejected because its blue score is lower then the last finality
// point.
// 5) We make a block that points to altChainTip, and check that it
// gets rejected because it doesn't have the last finality point in
// its selected parent chain.
func TestFinality(t *testing.T) {
	params := dagconfig.SimnetParams
	params.K = 1
	params.FinalityDuration = 100 * params.TargetTimePerBlock
	dag, teardownFunc, err := blockdag.DAGSetup("TestFinality", true, blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	buildNodeToDag := func(parentHashes []*daghash.Hash) (*util.Block, error) {
		msgBlock, err := mining.PrepareBlockForTest(dag, parentHashes, nil, false)
		if err != nil {
			return nil, err
		}
		block := util.NewBlock(msgBlock)

		isOrphan, isDelayed, err := dag.ProcessBlock(block, blockdag.BFNoPoWCheck)
		if err != nil {
			return nil, err
		}
		if isDelayed {
			return nil, errors.Errorf("ProcessBlock: block " +
				"is too far in the future")
		}
		if isOrphan {
			return nil, errors.Errorf("ProcessBlock: unexpected returned orphan block")
		}

		return block, nil
	}

	genesis := util.NewBlock(params.GenesisBlock)
	currentNode := genesis

	// First we build a chain of params.FinalityInterval blocks for future use
	for i := uint64(0); i < dag.FinalityInterval(); i++ {
		currentNode, err = buildNodeToDag([]*daghash.Hash{currentNode.Hash()})
		if err != nil {
			t.Fatalf("TestFinality: buildNodeToDag unexpectedly returned an error: %v", err)
		}
	}

	altChainTip := currentNode

	// Now we build a new chain of 2 * params.FinalityInterval blocks, pointed to genesis, and
	// we expect the block with height 1 * params.FinalityInterval to be the last finality point
	currentNode = genesis
	for i := uint64(0); i < dag.FinalityInterval(); i++ {
		currentNode, err = buildNodeToDag([]*daghash.Hash{currentNode.Hash()})
		if err != nil {
			t.Fatalf("TestFinality: buildNodeToDag unexpectedly returned an error: %v", err)
		}
	}

	expectedFinalityPoint := currentNode

	for i := uint64(0); i < dag.FinalityInterval(); i++ {
		currentNode, err = buildNodeToDag([]*daghash.Hash{currentNode.Hash()})
		if err != nil {
			t.Fatalf("TestFinality: buildNodeToDag unexpectedly returned an error: %v", err)
		}
	}

	if !dag.LastFinalityPointHash().IsEqual(expectedFinalityPoint.Hash()) {
		t.Errorf("TestFinality: dag.lastFinalityPoint expected to be %v but got %v", expectedFinalityPoint, dag.LastFinalityPointHash())
	}

	// Here we check that even if we create a parallel tip (a new tip with
	// the same parents as the current one) with the same blue score as the
	// current tip, it still won't affect the last finality point.
	_, err = buildNodeToDag(currentNode.MsgBlock().Header.ParentHashes)
	if err != nil {
		t.Fatalf("TestFinality: buildNodeToDag unexpectedly returned an error: %v", err)
	}
	if !dag.LastFinalityPointHash().IsEqual(expectedFinalityPoint.Hash()) {
		t.Errorf("TestFinality: dag.lastFinalityPoint was unexpectly changed")
	}

	// Here we check that a block with lower blue score than the last finality
	// point will get rejected
	fakeCoinbaseTx, err := dag.NextBlockCoinbaseTransaction(nil, nil)
	if err != nil {
		t.Errorf("NextBlockCoinbaseTransaction: %s", err)
	}
	merkleRoot := blockdag.BuildHashMerkleTreeStore([]*util.Tx{fakeCoinbaseTx}).Root()
	beforeFinalityBlock := domainmessage.NewMsgBlock(&domainmessage.BlockHeader{
		Version:              0x10000000,
		ParentHashes:         []*daghash.Hash{genesis.Hash()},
		HashMerkleRoot:       merkleRoot,
		AcceptedIDMerkleRoot: &daghash.ZeroHash,
		UTXOCommitment:       &daghash.ZeroHash,
		Timestamp:            dag.SelectedTipHeader().Timestamp,
		Bits:                 genesis.MsgBlock().Header.Bits,
	})
	beforeFinalityBlock.AddTransaction(fakeCoinbaseTx.MsgTx())
	_, _, err = dag.ProcessBlock(util.NewBlock(beforeFinalityBlock), blockdag.BFNoPoWCheck)
	if err == nil {
		t.Errorf("TestFinality: buildNodeToDag expected an error but got <nil>")
	}
	var ruleErr blockdag.RuleError
	if errors.As(err, &ruleErr) {
		if ruleErr.ErrorCode != blockdag.ErrFinality {
			t.Errorf("TestFinality: buildNodeToDag expected an error with code %v but instead got %v", blockdag.ErrFinality, ruleErr.ErrorCode)
		}
	} else {
		t.Errorf("TestFinality: buildNodeToDag got unexpected error: %v", err)
	}

	// Here we check that a block that doesn't have the last finality point in
	// its selected parent chain will get rejected
	_, err = buildNodeToDag([]*daghash.Hash{altChainTip.Hash()})
	if err == nil {
		t.Errorf("TestFinality: buildNodeToDag expected an error but got <nil>")
	}
	if errors.As(err, &ruleErr) {
		if ruleErr.ErrorCode != blockdag.ErrFinality {
			t.Errorf("TestFinality: buildNodeToDag expected an error with code %v but instead got %v", blockdag.ErrFinality, ruleErr.ErrorCode)
		}
	} else {
		t.Errorf("TestFinality: buildNodeToDag got unexpected error: %v", ruleErr)
	}
}

// TestFinalityInterval tests that the finality interval is
// smaller then domainmessage.MaxInvPerMsg, so when a peer receives
// a getblocks message it should always be able to send
// all the necessary invs.
func TestFinalityInterval(t *testing.T) {
	netParams := []*dagconfig.Params{
		&dagconfig.MainnetParams,
		&dagconfig.TestnetParams,
		&dagconfig.DevnetParams,
		&dagconfig.RegressionNetParams,
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

			if dag.FinalityInterval() > domainmessage.MaxInvPerMsg {
				t.Errorf("FinalityInterval in %s should be lower or equal to domainmessage.MaxInvPerMsg", params.Name)
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

	block1, err := mining.PrepareBlockForTest(dag, []*daghash.Hash{params.GenesisHash}, nil, false)
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
	txIn := &domainmessage.TxIn{
		PreviousOutpoint: domainmessage.Outpoint{TxID: *cbTx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         domainmessage.MaxTxInSequenceNum,
	}
	txOut := &domainmessage.TxOut{
		ScriptPubKey: blockdag.OpTrueScript,
		Value:        uint64(1),
	}
	tx := domainmessage.NewNativeMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{txIn}, []*domainmessage.TxOut{txOut})

	chainedTxIn := &domainmessage.TxIn{
		PreviousOutpoint: domainmessage.Outpoint{TxID: *tx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         domainmessage.MaxTxInSequenceNum,
	}

	scriptPubKey, err := txscript.PayToScriptHashScript(blockdag.OpTrueScript)
	if err != nil {
		t.Fatalf("Failed to build public key script: %s", err)
	}
	chainedTxOut := &domainmessage.TxOut{
		ScriptPubKey: scriptPubKey,
		Value:        uint64(1),
	}
	chainedTx := domainmessage.NewNativeMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{chainedTxIn}, []*domainmessage.TxOut{chainedTxOut})

	block2, err := mining.PrepareBlockForTest(dag, []*daghash.Hash{block1.BlockHash()}, []*domainmessage.MsgTx{tx}, false)
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

	nonChainedTxIn := &domainmessage.TxIn{
		PreviousOutpoint: domainmessage.Outpoint{TxID: *cbTx.TxID(), Index: 0},
		SignatureScript:  signatureScript,
		Sequence:         domainmessage.MaxTxInSequenceNum,
	}
	nonChainedTxOut := &domainmessage.TxOut{
		ScriptPubKey: scriptPubKey,
		Value:        uint64(1),
	}
	nonChainedTx := domainmessage.NewNativeMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{nonChainedTxIn}, []*domainmessage.TxOut{nonChainedTxOut})

	block3, err := mining.PrepareBlockForTest(dag, []*daghash.Hash{block1.BlockHash()}, []*domainmessage.MsgTx{nonChainedTx}, false)
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
		var txs []*domainmessage.MsgTx
		if !previousBlock.IsGenesis() {
			previousCoinbaseTx := previousBlock.MsgBlock().Transactions[0]
			signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
			if err != nil {
				t.Fatalf("TestOrderInDiffFromAcceptanceData: Failed to build signature script: %s", err)
			}
			txIn := &domainmessage.TxIn{
				PreviousOutpoint: domainmessage.Outpoint{TxID: *previousCoinbaseTx.TxID(), Index: 0},
				SignatureScript:  signatureScript,
				Sequence:         domainmessage.MaxTxInSequenceNum,
			}
			txOut := &domainmessage.TxOut{
				ScriptPubKey: blockdag.OpTrueScript,
				Value:        uint64(1),
			}
			txs = append(txs, domainmessage.NewNativeMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{txIn}, []*domainmessage.TxOut{txOut}))
		}

		// Create the block
		msgBlock, err := mining.PrepareBlockForTest(dag, []*daghash.Hash{previousBlock.Hash()}, txs, false)
		if err != nil {
			t.Fatalf("TestOrderInDiffFromAcceptanceData: Failed to prepare block: %s", err)
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

	cbTxs := []*domainmessage.MsgTx{}
	for i := 0; i < 4; i++ {
		fundsBlock, err := mining.PrepareBlockForTest(dag, dag.TipHashes(), nil, false)
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

	tx1In := &domainmessage.TxIn{
		PreviousOutpoint: *domainmessage.NewOutpoint(cbTxs[0].TxID(), 0),
		Sequence:         domainmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	tx1Out := &domainmessage.TxOut{
		Value:        cbTxs[0].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	tx1 := domainmessage.NewSubnetworkMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{tx1In}, []*domainmessage.TxOut{tx1Out}, subnetworkID, 10000, []byte{})

	tx2In := &domainmessage.TxIn{
		PreviousOutpoint: *domainmessage.NewOutpoint(cbTxs[1].TxID(), 0),
		Sequence:         domainmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	tx2Out := &domainmessage.TxOut{
		Value:        cbTxs[1].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	tx2 := domainmessage.NewSubnetworkMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{tx2In}, []*domainmessage.TxOut{tx2Out}, subnetworkID, 10000, []byte{})

	// Here we check that we can't process a block that has transactions that exceed the gas limit
	overLimitBlock, err := mining.PrepareBlockForTest(dag, dag.TipHashes(), []*domainmessage.MsgTx{tx1, tx2}, true)
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

	overflowGasTxIn := &domainmessage.TxIn{
		PreviousOutpoint: *domainmessage.NewOutpoint(cbTxs[2].TxID(), 0),
		Sequence:         domainmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	overflowGasTxOut := &domainmessage.TxOut{
		Value:        cbTxs[2].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	overflowGasTx := domainmessage.NewSubnetworkMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{overflowGasTxIn}, []*domainmessage.TxOut{overflowGasTxOut},
		subnetworkID, math.MaxUint64, []byte{})

	// Here we check that we can't process a block that its transactions' gas overflows uint64
	overflowGasBlock, err := mining.PrepareBlockForTest(dag, dag.TipHashes(), []*domainmessage.MsgTx{tx1, overflowGasTx}, true)
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
	nonExistentSubnetworkTxIn := &domainmessage.TxIn{
		PreviousOutpoint: *domainmessage.NewOutpoint(cbTxs[3].TxID(), 0),
		Sequence:         domainmessage.MaxTxInSequenceNum,
		SignatureScript:  signatureScript,
	}
	nonExistentSubnetworkTxOut := &domainmessage.TxOut{
		Value:        cbTxs[3].TxOut[0].Value,
		ScriptPubKey: scriptPubKey,
	}
	nonExistentSubnetworkTx := domainmessage.NewSubnetworkMsgTx(domainmessage.TxVersion, []*domainmessage.TxIn{nonExistentSubnetworkTxIn},
		[]*domainmessage.TxOut{nonExistentSubnetworkTxOut}, nonExistentSubnetwork, 1, []byte{})

	nonExistentSubnetworkBlock, err := mining.PrepareBlockForTest(dag, dag.TipHashes(), []*domainmessage.MsgTx{nonExistentSubnetworkTx, overflowGasTx}, true)
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
	validBlock, err := mining.PrepareBlockForTest(dag, dag.TipHashes(), []*domainmessage.MsgTx{tx1}, true)
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
