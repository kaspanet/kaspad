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
		{feePerKB: 5678, priority: 3},
		{feePerKB: 5678, priority: 1},
		{feePerKB: 5678, priority: 1}, // Duplicate fee and prio
		{feePerKB: 5678, priority: 5},
		{feePerKB: 5678, priority: 2},
		{feePerKB: 1234, priority: 3},
		{feePerKB: 1234, priority: 1},
		{feePerKB: 1234, priority: 5},
		{feePerKB: 1234, priority: 5}, // Duplicate fee and prio
		{feePerKB: 1234, priority: 2},
		{feePerKB: 10000, priority: 0}, // Higher fee, smaller prio
		{feePerKB: 0, priority: 10000}, // Higher prio, lower fee
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
			priority: prng.Float64() * 100,
		})
	}

	// Test sorting by fee per KB then priority.
	var highest *txPrioItem
	priorityQueue := newTxPriorityQueue(len(testItems), true)
	for i := 0; i < len(testItems); i++ {
		prioItem := testItems[i]
		if highest == nil {
			highest = prioItem
		}
		if prioItem.feePerKB >= highest.feePerKB &&
			prioItem.priority > highest.priority {

			highest = prioItem
		}
		heap.Push(priorityQueue, prioItem)
	}

	for i := 0; i < len(testItems); i++ {
		prioItem := heap.Pop(priorityQueue).(*txPrioItem)
		if prioItem.feePerKB >= highest.feePerKB &&
			prioItem.priority > highest.priority {

			t.Fatalf("fee sort: item (fee per KB: %v, "+
				"priority: %v) higher than than prev "+
				"(fee per KB: %v, priority %v)",
				prioItem.feePerKB, prioItem.priority,
				highest.feePerKB, highest.priority)
		}
		highest = prioItem
	}

	// Test sorting by priority then fee per KB.
	highest = nil
	priorityQueue = newTxPriorityQueue(len(testItems), false)
	for i := 0; i < len(testItems); i++ {
		prioItem := testItems[i]
		if highest == nil {
			highest = prioItem
		}
		if prioItem.priority >= highest.priority &&
			prioItem.feePerKB > highest.feePerKB {

			highest = prioItem
		}
		heap.Push(priorityQueue, prioItem)
	}

	for i := 0; i < len(testItems); i++ {
		prioItem := heap.Pop(priorityQueue).(*txPrioItem)
		if prioItem.priority >= highest.priority &&
			prioItem.feePerKB > highest.feePerKB {

			t.Fatalf("priority sort: item (fee per KB: %v, "+
				"priority: %v) higher than than prev "+
				"(fee per KB: %v, priority %v)",
				prioItem.feePerKB, prioItem.priority,
				highest.feePerKB, highest.priority)
		}
		highest = prioItem
	}
}

func TestNewBlockTemplate(t *testing.T) {
	params := dagconfig.SimNetParams
	params.BlockRewardMaturity = 0

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
		BlockMaxSize:      50000,
		BlockPrioritySize: 750000,
		TxMinFreeFee:      util.Amount(0),
	}

	// First we create a block to have coinbase funds for the rest of the test.
	txSource := &fakeTxSource{
		txDescs: []*TxDesc{},
	}

	var createCoinbaseTxPatch *monkey.PatchGuard
	createCoinbaseTxPatch = monkey.Patch(CreateCoinbaseTx, func(params *dagconfig.Params, coinbaseScript []byte, nextBlockHeight uint64, addr util.Address) (*util.Tx, error) {
		createCoinbaseTxPatch.Unpatch()
		defer createCoinbaseTxPatch.Restore()
		tx, err := CreateCoinbaseTx(params, coinbaseScript, nextBlockHeight, addr)
		if err != nil {
			return nil, err
		}
		msgTx := tx.MsgTx()
		//Here we split the coinbase to 10 outputs, so we'll be able to use it in many transactions
		out := msgTx.TxOut[0]
		out.Value /= 10
		for i := 0; i < 9; i++ {
			msgTx.AddTxOut(&*out)
		}
		return tx, nil
	})
	defer createCoinbaseTxPatch.Unpatch()

	blockTemplateGenerator := NewBlkTmplGenerator(&policy,
		&params, txSource, dag, blockdag.NewMedianTime(), txscript.NewSigCache(100000))

	template1, err := blockTemplateGenerator.NewBlockTemplate(nil)
	createCoinbaseTxPatch.Unpatch()
	if err != nil {
		t.Fatalf("NewBlockTemplate: %v", err)
	}

	isOrphan, err := dag.ProcessBlock(util.NewBlock(template1.Block), blockdag.BFNoPoWCheck)
	if err != nil {
		t.Fatalf("ProcessBlock: %v", err)
	}

	if isOrphan {
		t.Fatalf("ProcessBlock: template1 got unexpectedly orphan")
	}

	cbScript, err := StandardCoinbaseScript(dag.Height()+1, 0)
	if err != nil {
		t.Fatalf("standardCoinbaseScript: %v", err)
	}

	// We want to check that the miner filters coinbase transaction
	cbTx, err := CreateCoinbaseTx(&params, cbScript, dag.Height()+1, nil)
	if err != nil {
		t.Fatalf("createCoinbaseTx: %v", err)
	}

	template1CbTx := template1.Block.Transactions[0]

	// tx is a regular transaction, and should not be filtered by the miner
	txIn := &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  *template1CbTx.TxID(),
			Index: 0,
		},
		Sequence: wire.MaxTxInSequenceNum,
	}
	txOut := &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	tx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})

	// We want to check that the miner filters non finalized transactions
	txIn = &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  *template1CbTx.TxID(),
			Index: 1,
		},
		Sequence: 0,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	nonFinalizedTx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})
	nonFinalizedTx.LockTime = uint64(dag.Height() + 2)

	existingSubnetwork := &subnetworkid.SubnetworkID{0xff}
	nonExistingSubnetwork := &subnetworkid.SubnetworkID{0xfe}

	// We want to check that the miner filters transactions with non-existing subnetwork id. (It should first push it to the priority queue, and then ignore it)
	txIn = &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  *template1CbTx.TxID(),
			Index: 2,
		},
		Sequence: 0,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	nonExistingSubnetworkTx := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut},
		nonExistingSubnetwork, 1, []byte{})

	// We want to check that the miner doesn't filters transactions that do not exceed the subnetwork gas limit
	txIn = &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  *template1CbTx.TxID(),
			Index: 3,
		},
		Sequence: 0,
	}
	txOut = &wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	}
	subnetworkTx1 := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, existingSubnetwork, 1, []byte{})

	// We want to check that the miner filters transactions that exceed the subnetwork gas limit. (It should first push it to the priority queue, and then ignore it)
	txIn = &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			TxID:  *template1CbTx.TxID(),
			Index: 4,
		},
		Sequence: 0,
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

	standardCoinbaseScriptErrString := "standardCoinbaseScript err"

	var standardCoinbaseScriptPatch *monkey.PatchGuard
	standardCoinbaseScriptPatch = monkey.Patch(StandardCoinbaseScript, func(nextBlockHeight uint64, extraNonce uint64) ([]byte, error) {
		return nil, errors.New(standardCoinbaseScriptErrString)
	})
	defer standardCoinbaseScriptPatch.Unpatch()

	// We want to check that NewBlockTemplate will fail if standardCoinbaseScript returns an error
	_, err = blockTemplateGenerator.NewBlockTemplate(nil)
	standardCoinbaseScriptPatch.Unpatch()

	if err == nil || err.Error() != standardCoinbaseScriptErrString {
		t.Errorf("expected an error \"%v\" but got \"%v\"", standardCoinbaseScriptErrString, err)
	}
	if err == nil {
		t.Errorf("expected an error but got <nil>")
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

	template2, err := blockTemplateGenerator.NewBlockTemplate(nil)
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

	for _, tx := range template2.Block.Transactions[2:] {
		id := *tx.TxID()
		if _, ok := expectedTxs[id]; !ok {
			t.Errorf("Unexpected tx %v in template2's candidate block", id)
		}
		expectedTxs[id] = true
	}

	for id, exists := range expectedTxs {
		if !exists {
			t.Errorf("tx %v was expected to be in template2's candidate block, but wasn't", id)
		}
	}
}
