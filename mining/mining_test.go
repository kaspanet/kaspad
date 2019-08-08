// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"container/heap"
	"errors"
	"math/rand"
	"testing"

	"github.com/daglabs/btcd/util/subnetworkid"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"

	"github.com/daglabs/btcd/util"
)

// TestTxFeePrioHeap ensures the priority queue for transaction fees and
// priorities works as expected.
func TestTxFeePrioHeap(t *testing.T) {
	// Create some fake priority items that exercise the expected sort
	// edge conditions.
	testItems := []*txPrioItem{
		{feePerKB: 5678},
		{feePerKB: 5678}, // Duplicate fee
		{feePerKB: 1234},
		{feePerKB: 10000}, // High fee
		{feePerKB: 0},     // Zero fee
	}

	// Add random data in addition to the edge conditions already manually
	// specified.
	randSeed := rand.Int63()
	defer func() {
		if t.Failed() {
			t.Logf("Random numbers using seed: %v", randSeed)
		}
	}()
	prng := rand.New(rand.NewSource(randSeed))
	for i := 0; i < 1000; i++ {
		testItems = append(testItems, &txPrioItem{
			feePerKB: uint64(prng.Float64() * util.SatoshiPerBitcoin),
		})
	}

	// Test sorting by fee per KB
	var highest *txPrioItem
	priorityQueue := newTxPriorityQueue(len(testItems))
	for _, prioItem := range testItems {
		if highest == nil || prioItem.feePerKB >= highest.feePerKB {
			highest = prioItem
		}
		heap.Push(priorityQueue, prioItem)
	}

	for i := 0; i < len(testItems); i++ {
		prioItem := heap.Pop(priorityQueue).(*txPrioItem)
		if prioItem.feePerKB > highest.feePerKB {
			t.Fatalf("fee sort: item (fee per KB: %v) "+
				"higher than than prev (fee per KB: %v)",
				prioItem.feePerKB, highest.feePerKB)
		}
		highest = prioItem
	}
}

func TestNewBlockTemplate(t *testing.T) {
	params := dagconfig.SimNetParams
	params.BlockCoinbaseMaturity = 0

	dag, teardownFunc, err := blockdag.DAGSetup("TestNewBlockTemplate", blockdag.Config{
		DAGParams: &params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	pkScript, err := txscript.NewScriptBuilder().AddOp(txscript.OpTrue).Script()
	if err != nil {
		t.Fatalf("Failed to create pkScript: %v", err)
	}

	policy := Policy{
		BlockMaxMass: 50000,
	}

	// First we create a block to have coinbase funds for the rest of the test.
	txSource := &fakeTxSource{
		txDescs: []*TxDesc{},
	}

	blockTemplateGenerator := NewBlkTmplGenerator(&policy,
		&params, txSource, dag, blockdag.NewMedianTime(), txscript.NewSigCache(100000))

	OpTrueAddr, err := OpTrueAddress(params.Prefix)
	if err != nil {
		t.Fatalf("OpTrueAddress: %s", err)
	}
	template1, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	if err != nil {
		t.Fatalf("NewBlockTemplate: %v", err)
	}

	isOrphan, delay, err := dag.ProcessBlock(util.NewBlock(template1.Block), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock: %v", err)
	}

	if delay != 0 {
		t.Fatalf("ProcessBlock: template1 " +
			"is too far in the future")
	}

	if isOrphan {
		t.Fatalf("ProcessBlock: template1 got unexpectedly orphan")
	}

	// We create another 4 blocks to in order to create more funds for tests.
	cbTxs := []*wire.MsgTx{template1.Block.Transactions[util.CoinbaseTransactionIndex]}
	for i := 0; i < 4; i++ {
		template, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
		if err != nil {
			t.Fatalf("NewBlockTemplate: %v", err)
		}
		isOrphan, delay, err = dag.ProcessBlock(util.NewBlock(template.Block), blockdag.BFNoPoWCheck)
		if err != nil {
			t.Fatalf("ProcessBlock: %v", err)
		}
		if delay != 0 {
			t.Fatalf("ProcessBlock: template " +
				"is too far in the future")
		}
		if isOrphan {
			t.Fatalf("ProcessBlock: template got unexpectedly orphan")
		}
		cbTxs = append(cbTxs, template.Block.Transactions[util.CoinbaseTransactionIndex])
	}

	// We want to check that the miner filters coinbase transaction
	cbTx, err := dag.NextBlockCoinbaseTransaction(nil, nil)
	if err != nil {
		t.Fatalf("createCoinbaseTx: %v", err)
	}

	signatureScript, err := txscript.PayToScriptHashSignatureScript(blockdag.OpTrueScript, nil)
	if err != nil {
		t.Fatalf("Error creating signature script: %s", err)
	}

	// tx is a regular transaction, and should not be filtered by the miner
	txIn := &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{
			TxID:  *cbTxs[0].TxID(),
			Index: 0,
		},
		Sequence:        wire.MaxTxInSequenceNum,
		SignatureScript: signatureScript,
	}
	txOut := &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	tx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})

	// We want to check that the miner filters non finalized transactions
	txIn = &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{
			TxID:  *cbTxs[1].TxID(),
			Index: 0,
		},
		Sequence:        0,
		SignatureScript: signatureScript,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	nonFinalizedTx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})
	nonFinalizedTx.LockTime = uint64(dag.ChainHeight() + 2)

	existingSubnetwork := &subnetworkid.SubnetworkID{0xff}
	nonExistingSubnetwork := &subnetworkid.SubnetworkID{0xfe}

	// We want to check that the miner filters transactions with non-existing subnetwork id. (It should first push it to the priority queue, and then ignore it)
	txIn = &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{
			TxID:  *cbTxs[2].TxID(),
			Index: 0,
		},
		Sequence:        0,
		SignatureScript: signatureScript,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	nonExistingSubnetworkTx := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut},
		nonExistingSubnetwork, 1, []byte{})

	// We want to check that the miner doesn't filters transactions that do not exceed the subnetwork gas limit
	txIn = &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{
			TxID:  *cbTxs[3].TxID(),
			Index: 0,
		},
		Sequence:        0,
		SignatureScript: signatureScript,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	subnetworkTx1 := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, existingSubnetwork, 1, []byte{})

	// We want to check that the miner filters transactions that exceed the subnetwork gas limit. (It should first push it to the priority queue, and then ignore it)
	txIn = &wire.TxIn{
		PreviousOutpoint: wire.Outpoint{
			TxID: *cbTxs[4].TxID(),
		},
		Sequence:        0,
		SignatureScript: signatureScript,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	subnetworkTx2 := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, existingSubnetwork,
		100, // Subnetwork gas limit is 90
		[]byte{})

	txSource.txDescs = []*TxDesc{
		{
			Tx: cbTx,
		},
		{
			Tx: util.NewTx(tx),
		},
		{
			Tx: util.NewTx(nonFinalizedTx),
		},
		{
			Tx: util.NewTx(subnetworkTx1),
		},
		{
			Tx: util.NewTx(subnetworkTx2),
		},
		{
			Tx: util.NewTx(nonExistingSubnetworkTx),
		},
	}

	// Here we check that the miner's priorty queue has the expected transactions after filtering.
	popReturnedUnexpectedValue := false
	expectedPops := map[daghash.TxID]bool{
		*tx.TxID():                      false,
		*subnetworkTx1.TxID():           false,
		*subnetworkTx2.TxID():           false,
		*nonExistingSubnetworkTx.TxID(): false,
	}
	var popPatch *monkey.PatchGuard
	popPatch = monkey.Patch((*txPriorityQueue).Pop, func(pq *txPriorityQueue) interface{} {
		popPatch.Unpatch()
		defer popPatch.Restore()

		item, ok := pq.Pop().(*txPrioItem)
		if _, expected := expectedPops[*item.tx.ID()]; expected && ok {
			expectedPops[*item.tx.ID()] = true
		} else {
			popReturnedUnexpectedValue = true
		}
		return item
	})
	defer popPatch.Unpatch()

	// Here we define nonExistingSubnetwork to be non-exist, and existingSubnetwork to have a gas limit of 90
	gasLimitPatch := monkey.Patch((*blockdag.SubnetworkStore).GasLimit, func(_ *blockdag.SubnetworkStore, subnetworkID *subnetworkid.SubnetworkID) (uint64, error) {
		if subnetworkID.IsEqual(nonExistingSubnetwork) {
			return 0, errors.New("not found")
		}
		return 90, nil
	})
	defer gasLimitPatch.Unpatch()

	template3, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	popPatch.Unpatch()
	gasLimitPatch.Unpatch()

	if err != nil {
		t.Errorf("NewBlockTemplate: unexpected error: %v", err)
	}

	if popReturnedUnexpectedValue {
		t.Errorf("(*txPriorityQueue).Pop returned unexpected value")
	}

	for id, popped := range expectedPops {
		if !popped {
			t.Errorf("tx %v was expected to pop, but wasn't", id)
		}
	}

	expectedTxs := map[daghash.TxID]bool{
		*tx.TxID():            false,
		*subnetworkTx1.TxID(): false,
	}

	for _, tx := range template3.Block.Transactions[util.CoinbaseTransactionIndex+1:] {
		id := *tx.TxID()
		if _, ok := expectedTxs[id]; !ok {
			t.Errorf("Unexpected tx %v in template3's candidate block", id)
		}
		expectedTxs[id] = true
	}

	for id, exists := range expectedTxs {
		if !exists {
			t.Errorf("tx %v was expected to be in template3's candidate block, but wasn't", id)
		}
	}
}
