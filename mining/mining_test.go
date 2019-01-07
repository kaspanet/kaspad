// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"container/heap"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/txscript"
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

type fakeTxSource struct {
	txDescs []*TxDesc
}

func (txs *fakeTxSource) LastUpdated() time.Time {
	return time.Unix(0, 0)
}

func (txs *fakeTxSource) MiningDescs() []*TxDesc {
	return txs.txDescs
}

func (txs *fakeTxSource) HaveTransaction(hash *daghash.Hash) bool {
	for _, desc := range txs.txDescs {
		if *desc.Tx.Hash() == *hash {
			return true
		}
	}
	return false
}

func TestSomething(t *testing.T) {
	params := &dagconfig.SimNetParams

	// Use a hard coded key pair for deterministic results.
	keyBytes, err := hex.DecodeString("700868df1838811ffbdf918fb482c1f7e" +
		"ad62db4b97bd7012c23e726485e577d")
	if err != nil {
		t.Fatalf("hex.DecodeString: %v", err)
	}
	signKey, signPub := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)

	// Generate associated pay-to-script-hash address and resulting payment
	// script.
	pubKeyBytes := signPub.SerializeCompressed()
	payPubKeyAddr, err := util.NewAddressPubKey(pubKeyBytes, params.Prefix)
	if err != nil {
		t.Fatalf("NewAddressPubKey: %v", err)
	}
	pkHashAddr := payPubKeyAddr.AddressPubKeyHash()
	pkScript, err := txscript.PayToAddrScript(pkHashAddr)
	payAddr, err := util.NewAddressPubKeyHash(
		util.Hash160(pubKeyBytes), util.Bech32PrefixDAGTest)

	dag, teardownFunc, err := blockdag.DAGSetup("TestSomething", blockdag.Config{
		DAGParams: params,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()
	policy := Policy{
		BlockMaxSize:      50000,
		BlockPrioritySize: 750000,
		TxMinFreeFee:      util.Amount(1000),
	}

	txSource := &fakeTxSource{
		txDescs: []*TxDesc{},
	}

	blockTemplateGenerator := NewBlkTmplGenerator(&policy,
		params, txSource, dag, blockdag.NewMedianTime(), txscript.NewSigCache(100000))

	template1, err := blockTemplateGenerator.NewBlockTemplate(payAddr)
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

	cbScript, err := standardCoinbaseScript(dag.Height()+1, 0)
	if err != nil {
		t.Fatalf("standardCoinbaseScript: %v", err)
	}

	cbTx, err := createCoinbaseTx(params, cbScript, dag.Height()+1, nil)
	if err != nil {
		t.Fatalf("createCoinbaseTx: %v", err)
	}

	template1CbTx := template1.Block.Transactions[0]

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  template1CbTx.TxHash(),
			Index: 0,
		},
		Sequence: wire.MaxTxInSequenceNum,
	})
	tx.AddTxOut(&wire.TxOut{
		PkScript: pkScript,
		Value:    1,
	})

	// Sign the new transaction.
	tx.TxIn[0].SignatureScript, err = txscript.SignatureScript(tx, 0, pkScript,
		txscript.SigHashAll, signKey, true)
	if err != nil {
		t.Fatalf("SignatureScript: %v", err)
	}

	txSource.txDescs = []*TxDesc{
		{
			Tx: cbTx,
		},
		{
			Tx: util.NewTx(tx),
		},
	}

	functionCalledAsExpected := false

	tests := []struct {
		target              interface{}
		replacement         interface{}
		expectsError        bool
		expectedErrorString string
	}{
		{
			target:       nil,
			replacement:  nil,
			expectsError: false,
		},
		{
			target: standardCoinbaseScript,
			replacement: func(nextBlockHeight int32, extraNonce uint64) ([]byte, error) {
				functionCalledAsExpected = true
				return nil, errors.New("standardCoinbaseScript err")
			},
			expectsError:        true,
			expectedErrorString: "standardCoinbaseScript err",
		},
		{
			target: log.Tracef,
			replacement: func() func(format string, params ...interface{}) {
				i := 0
				return func(format string, params ...interface{}) {
					if i == 0 {
						functionCalledAsExpected = true
					}
					switch i {
					case 0:
						if fmt.Sprintf(format, params...) != fmt.Sprintf("Skipping coinbase tx %s", cbTx.Hash()) {
							functionCalledAsExpected = false
						}
					}
					i++
				}
			}(),
			expectsError: false,
		},
	}

	for i, test := range tests {
		func() {
			functionCalledAsExpected = false
			if test.target != nil {
				guard := monkey.Patch(test.target, test.replacement)
				defer guard.Unpatch()
			} else {
				functionCalledAsExpected = true
			}
			_, err = blockTemplateGenerator.NewBlockTemplate(payAddr)

			if !functionCalledAsExpected {
				t.Errorf("Test %v: function wasn't called as expected", i)
			}

			if !test.expectsError {
				if err != nil {
					t.Errorf("Test %v: unexpected error: %v", i, err)
				}
			} else {
				if err == nil || err.Error() != test.expectedErrorString {
					t.Errorf("Test %v: expected an error \"%v\" but got \"%v\"", i, test.expectedErrorString, err)
				}
				if err == nil {
					t.Errorf("Test %v: expected an error but got <nil>", i)
				}
			}
		}()
	}
}
