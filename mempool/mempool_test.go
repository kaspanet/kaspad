// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mempool

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/util/testtools"

	"bou.ke/monkey"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/blockdag/indexers"
	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/txscript"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

// fakeChain is used by the pool harness to provide generated test utxos and
// a current faked chain height to the pool callbacks.  This, in turn, allows
// transactions to appear as though they are spending completely valid utxos.
type fakeChain struct {
	sync.RWMutex
	currentHeight  int32
	medianTimePast time.Time
}

// BestHeight returns the current height associated with the fake chain
// instance.
func (s *fakeChain) BestHeight() int32 {
	s.RLock()
	height := s.currentHeight
	s.RUnlock()
	return height
}

// SetHeight sets the current height associated with the fake chain instance.
func (s *fakeChain) SetHeight(height int32) {
	s.Lock()
	s.currentHeight = height
	s.Unlock()
}

// MedianTimePast returns the current median time past associated with the fake
// chain instance.
func (s *fakeChain) MedianTimePast() time.Time {
	s.RLock()
	mtp := s.medianTimePast
	s.RUnlock()
	return mtp
}

// SetMedianTimePast sets the current median time past associated with the fake
// chain instance.
func (s *fakeChain) SetMedianTimePast(mtp time.Time) {
	s.Lock()
	s.medianTimePast = mtp
	s.Unlock()
}

func calcSequenceLock(tx *util.Tx,
	utxoSet blockdag.UTXOSet) (*blockdag.SequenceLock, error) {

	return &blockdag.SequenceLock{
		Seconds:     -1,
		BlockHeight: -1,
	}, nil
}

// spendableOutpoint is a convenience type that houses a particular utxo and the
// amount associated with it.
type spendableOutpoint struct {
	outPoint wire.OutPoint
	amount   util.Amount
}

// txOutToSpendableOutpoint returns a spendable outpoint given a transaction and index
// of the output to use.  This is useful as a convenience when creating test
// transactions.
func txOutToSpendableOutpoint(tx *util.Tx, outputNum uint32) spendableOutpoint {
	return spendableOutpoint{
		outPoint: wire.OutPoint{TxID: *tx.ID(), Index: outputNum},
		amount:   util.Amount(tx.MsgTx().TxOut[outputNum].Value),
	}
}

// poolHarness provides a harness that includes functionality for creating and
// signing transactions as well as a fake chain that provides utxos for use in
// generating valid transactions.
type poolHarness struct {
	// signKey is the signing key used for creating transactions throughout
	// the tests.
	//
	// payAddr is the p2sh address for the signing key and is used for the
	// payment address throughout the tests.
	signKey     *btcec.PrivateKey
	payAddr     util.Address
	payScript   []byte
	chainParams *dagconfig.Params

	chain  *fakeChain
	txPool *TxPool
}

// CreateCoinbaseTx returns a coinbase transaction with the requested number of
// outputs paying an appropriate subsidy based on the passed block height to the
// address associated with the harness.  It automatically uses a standard
// signature script that starts with the block height that is required by
// version 2 blocks.
func (p *poolHarness) CreateCoinbaseTx(blockHeight int32, numOutputs uint32) (*util.Tx, error) {
	// Create standard coinbase script.
	extraNonce := int64(0)
	coinbaseScript, err := txscript.NewScriptBuilder().
		AddInt64(int64(blockHeight)).AddInt64(extraNonce).Script()
	if err != nil {
		return nil, err
	}

	txIns := []*wire.TxIn{&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		PreviousOutPoint: *wire.NewOutPoint(&daghash.TxID{},
			wire.MaxPrevOutIndex),
		SignatureScript: coinbaseScript,
		Sequence:        wire.MaxTxInSequenceNum,
	}}

	txOuts := []*wire.TxOut{}
	totalInput := blockdag.CalcBlockSubsidy(blockHeight, p.chainParams)
	amountPerOutput := totalInput / uint64(numOutputs)
	remainder := totalInput - amountPerOutput*uint64(numOutputs)
	for i := uint32(0); i < numOutputs; i++ {
		// Ensure the final output accounts for any remainder that might
		// be left from splitting the input amount.
		amount := amountPerOutput
		if i == numOutputs-1 {
			amount = amountPerOutput + remainder
		}
		txOuts = append(txOuts, &wire.TxOut{
			PkScript: p.payScript,
			Value:    amount,
		})
	}

	return util.NewTx(wire.NewMsgTx(wire.TxVersion, txIns, txOuts, nil, 0, nil)), nil
}

// CreateSignedTxForSubnetwork creates a new signed transaction that consumes the provided
// inputs and generates the provided number of outputs by evenly splitting the
// total input amount.  All outputs will be to the payment script associated
// with the harness and all inputs are assumed to do the same.
func (p *poolHarness) CreateSignedTxForSubnetwork(inputs []spendableOutpoint, numOutputs uint32, subnetworkID *subnetworkid.SubnetworkID, gas uint64) (*util.Tx, error) {
	// Calculate the total input amount and split it amongst the requested
	// number of outputs.
	var totalInput util.Amount
	for _, input := range inputs {
		totalInput += input.amount
	}
	amountPerOutput := uint64(totalInput) / uint64(numOutputs)
	remainder := uint64(totalInput) - amountPerOutput*uint64(numOutputs)

	txIns := []*wire.TxIn{}
	for _, input := range inputs {
		txIns = append(txIns, &wire.TxIn{
			PreviousOutPoint: input.outPoint,
			SignatureScript:  nil,
			Sequence:         wire.MaxTxInSequenceNum,
		})
	}

	txOuts := []*wire.TxOut{}
	for i := uint32(0); i < numOutputs; i++ {
		// Ensure the final output accounts for any remainder that might
		// be left from splitting the input amount.
		amount := amountPerOutput
		if i == numOutputs-1 {
			amount = amountPerOutput + remainder
		}
		txOuts = append(txOuts, &wire.TxOut{
			PkScript: p.payScript,
			Value:    amount,
		})
	}

	tx := wire.NewMsgTx(wire.TxVersion, txIns, txOuts, subnetworkID, gas, []byte{})

	// Sign the new transaction.
	for i := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, p.payScript,
			txscript.SigHashAll, p.signKey, true)
		if err != nil {
			return nil, err
		}
		tx.TxIn[i].SignatureScript = sigScript
	}

	return util.NewTx(tx), nil
}

// CreateSignedTx creates a new signed transaction that consumes the provided
// inputs and generates the provided number of outputs by evenly splitting the
// total input amount.  All outputs will be to the payment script associated
// with the harness and all inputs are assumed to do the same.
func (p *poolHarness) CreateSignedTx(inputs []spendableOutpoint, numOutputs uint32) (*util.Tx, error) {
	return p.CreateSignedTxForSubnetwork(inputs, numOutputs, subnetworkid.SubnetworkIDNative, 0)
}

// CreateTxChain creates a chain of zero-fee transactions (each subsequent
// transaction spends the entire amount from the previous one) with the first
// one spending the provided outpoint.  Each transaction spends the entire
// amount of the previous one and as such does not include any fees.
func (p *poolHarness) CreateTxChain(firstOutput spendableOutpoint, numTxns uint32) ([]*util.Tx, error) {
	txChain := make([]*util.Tx, 0, numTxns)
	prevOutPoint := firstOutput.outPoint
	spendableAmount := firstOutput.amount
	for i := uint32(0); i < numTxns; i++ {
		// Create the transaction using the previous transaction output
		// and paying the full amount to the payment address associated
		// with the harness.
		txIn := &wire.TxIn{
			PreviousOutPoint: prevOutPoint,
			SignatureScript:  nil,
			Sequence:         wire.MaxTxInSequenceNum,
		}
		txOut := &wire.TxOut{
			PkScript: p.payScript,
			Value:    uint64(spendableAmount),
		}
		tx := wire.NewMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, nil, 0, nil)

		// Sign the new transaction.
		sigScript, err := txscript.SignatureScript(tx, 0, p.payScript,
			txscript.SigHashAll, p.signKey, true)
		if err != nil {
			return nil, err
		}
		tx.TxIn[0].SignatureScript = sigScript

		txChain = append(txChain, util.NewTx(tx))

		// Next transaction uses outputs from this one.
		prevOutPoint = wire.OutPoint{TxID: tx.TxID(), Index: 0}
	}

	return txChain, nil
}

// newPoolHarness returns a new instance of a pool harness initialized with a
// fake chain and a TxPool bound to it that is configured with a policy suitable
// for testing.  Also, the fake chain is populated with the returned spendable
// outputs so the caller can easily create new valid transactions which build
// off of it.
func newPoolHarness(dagParams *dagconfig.Params, numOutputs uint32, dbName string) (*poolHarness, []spendableOutpoint, func(), error) {
	// Use a hard coded key pair for deterministic results.
	keyBytes, err := hex.DecodeString("700868df1838811ffbdf918fb482c1f7e" +
		"ad62db4b97bd7012c23e726485e577d")
	if err != nil {
		return nil, nil, nil, err
	}
	signKey, signPub := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)

	// Generate associated pay-to-script-hash address and resulting payment
	// script.
	pubKeyBytes := signPub.SerializeCompressed()
	payPubKeyAddr, err := util.NewAddressPubKey(pubKeyBytes, dagParams.Prefix)
	if err != nil {
		return nil, nil, nil, err
	}
	payAddr := payPubKeyAddr.AddressPubKeyHash()
	pkScript, err := txscript.PayToAddrScript(payAddr)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create a new database and chain instance to run tests against.
	dag, teardownFunc, err := blockdag.DAGSetup(dbName, blockdag.Config{
		DAGParams:    dagParams,
		SubnetworkID: subnetworkid.SubnetworkIDSupportsAll,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Failed to setup DAG instance: %v", err)
	}
	defer func() {
		if err != nil {
			teardownFunc()
		}
	}()

	// Create a new fake chain and harness bound to it.
	chain := &fakeChain{}
	harness := poolHarness{
		signKey:     signKey,
		payAddr:     payAddr,
		payScript:   pkScript,
		chainParams: dagParams,

		chain: chain,
		txPool: New(&Config{
			DAG: dag,
			Policy: Policy{
				DisableRelayPriority: true,
				FreeTxRelayLimit:     15.0,
				MaxOrphanTxs:         5,
				MaxOrphanTxSize:      1000,
				MaxSigOpsPerTx:       blockdag.MaxSigOpsPerBlock / 5,
				MinRelayTxFee:        1000, // 1 Satoshi per byte
				MaxTxVersion:         1,
			},
			DAGParams:        dagParams,
			BestHeight:       chain.BestHeight,
			MedianTimePast:   chain.MedianTimePast,
			CalcSequenceLock: calcSequenceLock,
			SigCache:         nil,
			AddrIndex:        nil,
		}),
	}

	// Create a single coinbase transaction and add it to the harness
	// chain's utxo set and set the harness chain height such that the
	// coinbase will mature in the next block.  This ensures the txpool
	// accepts transactions which spend immature coinbases that will become
	// mature in the next block.
	outpoints := make([]spendableOutpoint, 0, numOutputs)
	curHeight := harness.chain.BestHeight()
	coinbase, err := harness.CreateCoinbaseTx(curHeight+1, numOutputs)
	if err != nil {
		return nil, nil, nil, err
	}
	harness.txPool.mpUTXOSet.AddTx(coinbase.MsgTx(), curHeight+1)
	for i := uint32(0); i < numOutputs; i++ {
		outpoints = append(outpoints, txOutToSpendableOutpoint(coinbase, i))
	}
	if dagParams.BlockRewardMaturity != 0 {
		harness.chain.SetHeight(int32(dagParams.BlockRewardMaturity) + curHeight)
	} else {
		harness.chain.SetHeight(curHeight + 1)
	}
	harness.chain.SetMedianTimePast(time.Now())

	return &harness, outpoints, teardownFunc, nil
}

// testContext houses a test-related state that is useful to pass to helper
// functions as a single argument.
type testContext struct {
	t       *testing.T
	harness *poolHarness
}

// testPoolMembership tests the transaction pool associated with the provided
// test context to determine if the passed transaction matches the provided
// orphan pool and transaction pool status.  It also further determines if it
// should be reported as available by the HaveTransaction function based upon
// the two flags and tests that condition as well.
func testPoolMembership(tc *testContext, tx *util.Tx, inOrphanPool, inTxPool bool, isDepends bool) {
	txID := tx.ID()
	gotOrphanPool := tc.harness.txPool.IsOrphanInPool(txID)
	if inOrphanPool != gotOrphanPool {
		_, file, line, _ := runtime.Caller(1)
		tc.t.Fatalf("%s:%d -- IsOrphanInPool: want %v, got %v", file,
			line, inOrphanPool, gotOrphanPool)
	}

	gotTxPool := tc.harness.txPool.IsTransactionInPool(txID)
	if inTxPool != gotTxPool {
		_, file, line, _ := runtime.Caller(1)
		tc.t.Fatalf("%s:%d -- IsTransactionInPool: want %v, got %v",
			file, line, inTxPool, gotTxPool)
	}

	gotIsDepends := tc.harness.txPool.IsInDependPool(txID)
	if isDepends != gotIsDepends {
		_, file, line, _ := runtime.Caller(1)
		tc.t.Fatalf("%s:%d -- IsInDependPool: want %v, got %v",
			file, line, isDepends, gotIsDepends)
	}

	gotHaveTx := tc.harness.txPool.HaveTransaction(txID)
	wantHaveTx := inOrphanPool || inTxPool
	if wantHaveTx != gotHaveTx {
		_, file, line, _ := runtime.Caller(1)
		tc.t.Fatalf("%s:%d -- HaveTransaction: want %v, got %v", file,
			line, wantHaveTx, gotHaveTx)
	}

	count := tc.harness.txPool.Count()
	txIDs := tc.harness.txPool.TxIDs()
	txDescs := tc.harness.txPool.TxDescs()
	txMiningDescs := tc.harness.txPool.MiningDescs()
	if count != len(txIDs) || count != len(txDescs) || count != len(txMiningDescs) {
		tc.t.Error("mempool.TxIDs(), mempool.TxDescs() and mempool.MiningDescs() have different length")
	}
	if inTxPool && !isDepends {
		wasFound := false
		for _, txI := range txIDs {
			if *txID == *txI {
				wasFound = true
				break
			}
		}
		if !wasFound {
			tc.t.Error("Can not find transaction in mempool.TxIDs")
		}

		wasFound = false
		for _, txd := range txDescs {
			if *txID == *txd.Tx.ID() {
				wasFound = true
				break
			}
		}
		if !wasFound {
			tc.t.Error("Can not find transaction in mempool.TxDescs")
		}

		wasFound = false
		for _, txd := range txMiningDescs {
			if *txID == *txd.Tx.ID() {
				wasFound = true
				break
			}
		}
		if !wasFound {
			tc.t.Error("Can not find transaction in mempool.MiningDescs")
		}
	}
}

func (p *poolHarness) createTx(outpoint spendableOutpoint, fee uint64, numOutputs int64) (*util.Tx, error) {
	txIns := []*wire.TxIn{&wire.TxIn{
		PreviousOutPoint: outpoint.outPoint,
		SignatureScript:  nil,
		Sequence:         wire.MaxTxInSequenceNum,
	}}

	txOuts := []*wire.TxOut{}
	amountPerOutput := (uint64(outpoint.amount) - fee) / uint64(numOutputs)
	for i := int64(0); i < numOutputs; i++ {
		txOuts = append(txOuts, &wire.TxOut{
			PkScript: p.payScript,
			Value:    amountPerOutput,
		})
	}
	tx := wire.NewMsgTx(wire.TxVersion, txIns, txOuts, nil, 0, nil)

	// Sign the new transaction.
	sigScript, err := txscript.SignatureScript(tx, 0, p.payScript,
		txscript.SigHashAll, p.signKey, true)
	if err != nil {
		return nil, err
	}
	tx.TxIn[0].SignatureScript = sigScript
	return util.NewTx(tx), nil
}

func TestProcessTransaction(t *testing.T) {
	params := dagconfig.SimNetParams
	params.BlockRewardMaturity = 0
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&params, 6, "TestProcessTransaction")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	//Checks that a transaction cannot be added to the transaction pool if it's already there
	tx, err := harness.createTx(spendableOuts[0], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectDuplicate {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectDuplicate, code)
	}

	orphanedTx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{}, Index: 1},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}

	//Checks that an orphaned transaction cannot be
	//added to the orphan pool if MaxOrphanTxs is 0
	harness.txPool.cfg.Policy.MaxOrphanTxs = 0
	_, err = harness.txPool.ProcessTransaction(orphanedTx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}
	testPoolMembership(tc, orphanedTx, false, false, false)

	harness.txPool.cfg.Policy.MaxOrphanTxs = 5
	_, err = harness.txPool.ProcessTransaction(orphanedTx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}

	//Checks that an orphaned transaction cannot be
	//added to the orphan pool if it's already there
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectDuplicate {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectDuplicate, code)
	}

	//Checks that a coinbase transaction cannot be added to the mempool
	curHeight := harness.chain.BestHeight()
	coinbase, err := harness.CreateCoinbaseTx(curHeight+1, 1)
	if err != nil {
		t.Errorf("CreateCoinbaseTx: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(coinbase, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectInvalid {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectInvalid, code)
	}

	//Checks that non standard transactions are rejected from the mempool
	nonStdTx, err := harness.createTx(spendableOuts[0], 0, 1)
	nonStdTx.MsgTx().Version = wire.TxVersion + 1
	_, err = harness.txPool.ProcessTransaction(nonStdTx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectNonstandard {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectNonstandard, code)
	}

	//Checks that a transaction is rejected from the mempool if its
	//size is above 50KB, and its fee is below the minimum relay fee
	bigLowFeeTx, err := harness.createTx(spendableOuts[1], 0, 2000) //A transaction with 2000 outputs, in order to make it bigger than 50kb
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(bigLowFeeTx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectInsufficientFee {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectInsufficientFee, code)
	}

	//Checks that if a ps2h sigscript has more sigops then maxStandardP2SHSigOps, it gets rejected

	//maxStandardP2SHSigOps is 15, so 16 OpCheckSig will make it a non standard script
	nonStdSigScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		AddOp(txscript.OpCheckSig).
		Script()
	if err != nil {
		t.Fatalf("Script: error creating nonStdSigScript: %v", err)
	}

	p2shPKScript, err := txscript.NewScriptBuilder().
		AddOp(txscript.OpHash160).
		AddData(util.Hash160(nonStdSigScript)).
		AddOp(txscript.OpEqual).
		Script()

	if err != nil {
		t.Fatalf("Script: error creating p2shPKScript: %v", err)
	}

	wrappedP2SHNonStdSigScript, err := txscript.NewScriptBuilder().AddData(nonStdSigScript).Script()
	if err != nil {
		t.Fatalf("Script: error creating wrappedP2shNonSigScript: %v", err)
	}

	dummyPrevOutHash, err := daghash.NewTxIDFromStr("01")
	if err != nil {
		t.Fatalf("NewShaHashFromStr: unexpected error: %v", err)
	}
	dummyPrevOut := wire.OutPoint{TxID: *dummyPrevOutHash, Index: 1}
	dummySigScript := bytes.Repeat([]byte{0x00}, 65)

	addrHash := [20]byte{0x01}
	addr, err := util.NewAddressPubKeyHash(addrHash[:],
		util.Bech32PrefixDAGTest)
	if err != nil {
		t.Fatalf("NewAddressPubKeyHash: unexpected error: %v", err)
	}
	dummyPkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		t.Fatalf("PayToAddrScript: unexpected error: %v", err)
	}
	p2shTx := util.NewTx(wire.NewMsgTx(1, nil, []*wire.TxOut{{Value: 5000000000, PkScript: p2shPKScript}}, nil, 0, nil))
	harness.txPool.mpUTXOSet.AddTx(p2shTx.MsgTx(), curHeight+1)

	txIns := []*wire.TxIn{{
		PreviousOutPoint: wire.OutPoint{TxID: *p2shTx.ID(), Index: 0},
		SignatureScript:  wrappedP2SHNonStdSigScript,
		Sequence:         wire.MaxTxInSequenceNum,
	}}
	txOuts := []*wire.TxOut{{
		Value:    5000000000,
		PkScript: dummyPkScript,
	}}
	nonStdSigScriptTx := util.NewTx(wire.NewMsgTx(1, txIns, txOuts, nil, 0, nil))
	_, err = harness.txPool.ProcessTransaction(nonStdSigScriptTx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectNonstandard {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectNonstandard, code)
	}
	expectedErrStr := fmt.Sprintf("transaction %v has a non-standard input: "+
		"transaction input #%d has "+
		"%d signature operations which is more "+
		"than the allowed max amount of %d",
		nonStdSigScriptTx.ID(), 0, 16, 15)
	if expectedErrStr != err.Error() {
		t.Errorf("Unexpected error message. Expected \"%s\" but got \"%s\"", expectedErrStr, err.Error())
	}

	//Checks that even if we accept non standard transactions, we reject by the MaxSigOpsPerTx consensus rule
	harness.txPool.cfg.Policy.AcceptNonStd = true
	harness.txPool.cfg.Policy.MaxSigOpsPerTx = 15
	_, err = harness.txPool.ProcessTransaction(nonStdSigScriptTx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectNonstandard {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectNonstandard, code)
	}
	expectedErrStr = fmt.Sprintf("transaction %v sigop count is too high: %v > %v",
		nonStdSigScriptTx.ID(), 16, 15)
	if expectedErrStr != err.Error() {
		t.Errorf("Unexpected error message. Expected \"%s\" but got \"%s\"", expectedErrStr, err.Error())
	}
	harness.txPool.cfg.Policy.AcceptNonStd = false

	//Checks that a transaction with no outputs will not get rejected
	noOutsTx := util.NewTx(wire.NewMsgTx(1, []*wire.TxIn{{
		PreviousOutPoint: dummyPrevOut,
		SignatureScript:  dummySigScript,
		Sequence:         wire.MaxTxInSequenceNum,
	}},
		nil, nil, 0, nil))
	_, err = harness.txPool.ProcessTransaction(noOutsTx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: %v", err)
	}

	//Checks that transactions get rejected from mempool if sequence lock is not active
	harness.txPool.cfg.CalcSequenceLock = func(tx *util.Tx,
		view blockdag.UTXOSet) (*blockdag.SequenceLock, error) {

		return &blockdag.SequenceLock{
			Seconds:     math.MaxInt64,
			BlockHeight: math.MaxInt32,
		}, nil
	}
	tx, err = harness.createTx(spendableOuts[2], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectNonstandard {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectNonstandard, code)
	}
	expectedErrStr = "transaction's sequence locks on inputs not met"
	if err.Error() != expectedErrStr {
		t.Errorf("Unexpected error message. Expected \"%s\" but got \"%s\"", expectedErrStr, err.Error())
	}
	harness.txPool.cfg.CalcSequenceLock = calcSequenceLock

	// This is done in order to increase the input age, so the tx priority will be higher
	harness.chain.SetHeight(curHeight + 100)
	harness.txPool.cfg.Policy.DisableRelayPriority = false
	//Transaction should be accepted to mempool although it has low fee, because its priority is above mining.MinHighPriority
	tx, err = harness.createTx(spendableOuts[3], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}

	//Transaction should be rejected from mempool because it has low fee, and its priority is above mining.MinHighPriority
	tx, err = harness.createTx(spendableOuts[4], 0, 100)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectInsufficientFee {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectInsufficientFee, code)
	}
	harness.txPool.cfg.Policy.DisableRelayPriority = true

	txIns = []*wire.TxIn{{
		PreviousOutPoint: spendableOuts[5].outPoint,
		SignatureScript:  []byte{02, 01}, //Unparsable script
		Sequence:         wire.MaxTxInSequenceNum,
	}}
	txOuts = []*wire.TxOut{{
		Value:    1,
		PkScript: dummyPkScript,
	}}
	tx = util.NewTx(wire.NewMsgTx(1, txIns, txOuts, nil, 0, nil))
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	fmt.Println(err)
	if err == nil {
		t.Errorf("ProcessTransaction: expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectNonstandard {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectNonstandard, code)
	}
}

func TestAddrIndex(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 2, "TestAddrIndex")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	harness.txPool.cfg.AddrIndex = &indexers.AddrIndex{}
	enteredAddUnconfirmedTx := false
	guard := monkey.Patch((*indexers.AddrIndex).AddUnconfirmedTx, func(idx *indexers.AddrIndex, tx *util.Tx, utxoSet blockdag.UTXOSet) {
		enteredAddUnconfirmedTx = true
	})
	defer guard.Unpatch()
	enteredRemoveUnconfirmedTx := false
	guard = monkey.Patch((*indexers.AddrIndex).RemoveUnconfirmedTx, func(idx *indexers.AddrIndex, hash *daghash.TxID) {
		enteredRemoveUnconfirmedTx = true
	})
	defer guard.Unpatch()

	tx, err := harness.createTx(spendableOuts[0], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}

	if !enteredAddUnconfirmedTx {
		t.Errorf("TestAddrIndex: (*indexers.AddrIndex).AddUnconfirmedTx was not called")
	}

	err = harness.txPool.RemoveTransaction(tx, false, false)
	if err != nil {
		t.Errorf("TestAddrIndex: unexpected error: %v", err)
	}

	if !enteredRemoveUnconfirmedTx {
		t.Errorf("TestAddrIndex: (*indexers.AddrIndex).RemoveUnconfirmedTx was not called")
	}
}

func TestFeeEstimatorCfg(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 2, "TestFeeEstimatorCfg")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	harness.txPool.cfg.FeeEstimator = &FeeEstimator{}
	enteredObserveTransaction := false
	guard := monkey.Patch((*FeeEstimator).ObserveTransaction, func(ef *FeeEstimator, t *TxDesc) {
		enteredObserveTransaction = true
	})
	defer guard.Unpatch()

	tx, err := harness.createTx(spendableOuts[0], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}

	if !enteredObserveTransaction {
		t.Errorf("TestFeeEstimatorCfg: (*FeeEstimator).ObserveTransaction was not called")
	}
}

func TestDoubleSpends(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 2, "TestDoubleSpends")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	//Add two transactions to the mempool
	tx1, err := harness.createTx(spendableOuts[0], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.txPool.ProcessTransaction(tx1, true, false, 0)

	tx2, err := harness.createTx(spendableOuts[1], 1, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.txPool.ProcessTransaction(tx2, true, false, 0)
	testPoolMembership(tc, tx1, false, true, false)
	testPoolMembership(tc, tx2, false, true, false)

	//Spends the same outpoint as tx2
	tx3, err := harness.createTx(spendableOuts[0], 2, 1) //We put here different fee to create different transaction hash
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}

	//First we try to add it to the mempool and see it rejected
	_, err = harness.txPool.ProcessTransaction(tx3, true, false, 0)
	if err == nil {
		t.Errorf("ProcessTransaction expected an error, not nil")
	}
	if code, _ := extractRejectCode(err); code != wire.RejectDuplicate {
		t.Errorf("Unexpected error code. Expected %v but got %v", wire.RejectDuplicate, code)
	}
	testPoolMembership(tc, tx3, false, false, false)

	//Then we assume tx3 is already in the DAG, so we need to remove
	//transactions that spends the same outpoints from the mempool
	harness.txPool.RemoveDoubleSpends(tx3)
	//Ensures that only the transaction that double spends the same
	//funds as tx3 is removed, and the other one remains unaffected
	testPoolMembership(tc, tx1, false, false, false)
	testPoolMembership(tc, tx2, false, true, false)
}

//TestFetchTransaction checks that FetchTransaction
//returns only transaction from the main pool and not from the orphan pool
func TestFetchTransaction(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestFetchTransaction")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	orphanedTx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{1}, Index: 1},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(orphanedTx, true, false, 0)
	testPoolMembership(tc, orphanedTx, true, false, false)
	fetchedorphanedTx, err := harness.txPool.FetchTransaction(orphanedTx.ID())
	if fetchedorphanedTx != nil {
		t.Fatalf("FetchTransaction: expected fetchedorphanedTx to be nil")
	}
	if err == nil {
		t.Errorf("FetchTransaction: expected an error, not nil")
	}

	tx, err := harness.createTx(spendableOuts[0], 0, 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	harness.txPool.ProcessTransaction(tx, true, false, 0)
	testPoolMembership(tc, tx, false, true, false)
	fetchedTx, err := harness.txPool.FetchTransaction(tx.ID())
	if !reflect.DeepEqual(fetchedTx, tx) {
		t.Fatalf("FetchTransaction: returned a transaction, but not the right one")
	}
	if err != nil {
		t.Errorf("FetchTransaction: unexpected error: %v", err)
	}

}

// TestSimpleOrphanChain ensures that a simple chain of orphans is handled
// properly.  In particular, it generates a chain of single input, single output
// transactions and inserts them while skipping the first linking transaction so
// they are all orphans.  Finally, it adds the linking transaction and ensures
// the entire orphan chain is moved to the transaction pool.
func TestSimpleOrphanChain(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestSimpleOrphanChain")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness.
	maxOrphans := uint32(harness.txPool.cfg.Policy.MaxOrphanTxs)
	chainedTxns, err := harness.CreateTxChain(spendableOuts[0], maxOrphans+1)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Ensure the orphans are accepted (only up to the maximum allowed so
	// none are evicted).
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept valid "+
				"orphan %v", err)
		}

		// Ensure no transactions were reported as accepted.
		if len(acceptedTxns) != 0 {
			t.Fatalf("ProcessTransaction: reported %d accepted "+
				"transactions from what should be an orphan",
				len(acceptedTxns))
		}

		// Ensure the transaction is in the orphan pool, is not in the
		// transaction pool, and is reported as available.
		testPoolMembership(tc, tx, true, false, false)
	}

	// Add the transaction which completes the orphan chain and ensure they
	// all get accepted.  Notice the accept orphans flag is also false here
	// to ensure it has no bearing on whether or not already existing
	// orphans in the pool are linked.
	acceptedTxns, err := harness.txPool.ProcessTransaction(chainedTxns[0],
		false, false, 0)
	if err != nil {
		t.Fatalf("ProcessTransaction: failed to accept valid "+
			"orphan %v", err)
	}
	if len(acceptedTxns) != len(chainedTxns) {
		t.Fatalf("ProcessTransaction: reported accepted transactions "+
			"length does not match expected -- got %d, want %d",
			len(acceptedTxns), len(chainedTxns))
	}
	for _, txD := range acceptedTxns {
		// Ensure the transaction is no longer in the orphan pool, is
		// now in the transaction pool, and is reported as available.
		testPoolMembership(tc, txD.Tx, false, true, txD.Tx != chainedTxns[0])
	}
}

// TestOrphanReject ensures that orphans are properly rejected when the allow
// orphans flag is not set on ProcessTransaction.
func TestOrphanReject(t *testing.T) {
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestOrphanReject")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness.
	maxOrphans := uint32(harness.txPool.cfg.Policy.MaxOrphanTxs)
	chainedTxns, err := harness.CreateTxChain(outputs[0], maxOrphans+1)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Ensure orphans are rejected when the allow orphans flag is not set.
	for _, tx := range chainedTxns[1:] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, false,
			false, 0)
		if err == nil {
			t.Fatalf("ProcessTransaction: did not fail on orphan "+
				"%v when allow orphans flag is false", tx.ID())
		}
		expectedErr := RuleError{}
		if reflect.TypeOf(err) != reflect.TypeOf(expectedErr) {
			t.Fatalf("ProcessTransaction: wrong error got: <%T> %v, "+
				"want: <%T>", err, err, expectedErr)
		}
		code, extracted := extractRejectCode(err)
		if !extracted {
			t.Fatalf("ProcessTransaction: failed to extract reject "+
				"code from error %q", err)
		}
		if code != wire.RejectDuplicate {
			t.Fatalf("ProcessTransaction: unexpected reject code "+
				"-- got %v, want %v", code, wire.RejectDuplicate)
		}

		// Ensure no transactions were reported as accepted.
		if len(acceptedTxns) != 0 {
			t.Fatal("ProcessTransaction: reported %d accepted "+
				"transactions from failed orphan attempt",
				len(acceptedTxns))
		}

		// Ensure the transaction is not in the orphan pool, not in the
		// transaction pool, and not reported as available
		testPoolMembership(tc, tx, false, false, false)
	}
}

//TestOrphanExpiration checks every time we add an orphan transaction
// it will check if we are beyond nextExpireScan, and if so, it will remove
// all expired orphan transactions
func TestOrphanExpiration(t *testing.T) {
	harness, _, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestOrphanExpiration")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	expiredTx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{}, Index: 0},
	}}, 1)
	harness.txPool.ProcessTransaction(expiredTx, true,
		false, 0)
	harness.txPool.orphans[*expiredTx.ID()].expiration = time.Unix(0, 0)

	tx1, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{1}, Index: 0},
	}}, 1)
	harness.txPool.ProcessTransaction(tx1, true,
		false, 0)

	//First check that expired orphan transactions are not removed before nextExpireScan
	testPoolMembership(tc, tx1, true, false, false)
	testPoolMembership(tc, expiredTx, true, false, false)

	//Force nextExpireScan to be in the past
	harness.txPool.nextExpireScan = time.Unix(0, 0)
	fmt.Println(harness.txPool.nextExpireScan.Unix())

	tx2, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{2}, Index: 0},
	}}, 1)
	harness.txPool.ProcessTransaction(tx2, true,
		false, 0)
	//Check that only expired orphan transactions are removed
	testPoolMembership(tc, tx1, true, false, false)
	testPoolMembership(tc, tx2, true, false, false)
	testPoolMembership(tc, expiredTx, false, false, false)
}

//TestMaxOrphanTxSize ensures that a transaction that is
//bigger than MaxOrphanTxSize will get rejected
func TestMaxOrphanTxSize(t *testing.T) {
	harness, _, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestMaxOrphanTxSize")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}
	harness.txPool.cfg.Policy.MaxOrphanTxSize = 1

	tx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{}, Index: 0},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(tx, true,
		false, 0)

	testPoolMembership(tc, tx, false, false, false)

	harness.txPool.cfg.Policy.MaxOrphanTxSize = math.MaxInt32
	harness.txPool.ProcessTransaction(tx, true,
		false, 0)
	testPoolMembership(tc, tx, true, false, false)

}

func TestRemoveTransaction(t *testing.T) {
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestRemoveTransaction")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}
	chainedTxns, err := harness.CreateTxChain(outputs[0], 5)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	for i, tx := range chainedTxns {
		_, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: %v", err)
		}

		testPoolMembership(tc, tx, false, true, i != 0)
	}

	//Checks that when removeRedeemers is false, the specified transaction is the only transaction that gets removed
	harness.txPool.RemoveTransaction(chainedTxns[3], false, true)
	testPoolMembership(tc, chainedTxns[3], false, false, false)
	testPoolMembership(tc, chainedTxns[4], false, true, false)

	//Checks that when removeRedeemers is true, all of the transaction that are dependent on it get removed
	harness.txPool.RemoveTransaction(chainedTxns[1], true, true)
	testPoolMembership(tc, chainedTxns[0], false, true, false)
	testPoolMembership(tc, chainedTxns[1], false, false, false)
	testPoolMembership(tc, chainedTxns[2], false, false, false)

	fakeWithDiffErr := "error from WithDiff"
	guard := monkey.Patch((*blockdag.DiffUTXOSet).WithDiff, func(_ *blockdag.DiffUTXOSet, _ *blockdag.UTXODiff) (blockdag.UTXOSet, error) {
		return nil, errors.New(fakeWithDiffErr)
	})
	defer guard.Unpatch()
	err = harness.txPool.RemoveTransaction(chainedTxns[0], false, false)
	if err == nil || err.Error() != fakeWithDiffErr {
		t.Errorf("RemoveTransaction: expected error %v but got %v", fakeWithDiffErr, err)
	}
}

// TestOrphanEviction ensures that exceeding the maximum number of orphans
// evicts entries to make room for the new ones.
func TestOrphanEviction(t *testing.T) {
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestOrphanEviction")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness that is long enough to be able to force
	// several orphan evictions.
	maxOrphans := uint32(harness.txPool.cfg.Policy.MaxOrphanTxs)
	chainedTxns, err := harness.CreateTxChain(outputs[0], maxOrphans+5)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Add enough orphans to exceed the max allowed while ensuring they are
	// all accepted.  This will cause an eviction.
	for _, tx := range chainedTxns[1:] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept valid "+
				"orphan %v", err)
		}

		// Ensure no transactions were reported as accepted.
		if len(acceptedTxns) != 0 {
			t.Fatalf("ProcessTransaction: reported %d accepted "+
				"transactions from what should be an orphan",
				len(acceptedTxns))
		}

		// Ensure the transaction is in the orphan pool, is not in the
		// transaction pool, and is reported as available.
		testPoolMembership(tc, tx, true, false, false)
	}

	// Figure out which transactions were evicted and make sure the number
	// evicted matches the expected number.
	var evictedTxns []*util.Tx
	for _, tx := range chainedTxns[1:] {
		if !harness.txPool.IsOrphanInPool(tx.ID()) {
			evictedTxns = append(evictedTxns, tx)
		}
	}
	expectedEvictions := len(chainedTxns) - 1 - int(maxOrphans)
	if len(evictedTxns) != expectedEvictions {
		t.Fatalf("unexpected number of evictions -- got %d, want %d",
			len(evictedTxns), expectedEvictions)
	}

	// Ensure none of the evicted transactions ended up in the transaction
	// pool.
	for _, tx := range evictedTxns {
		testPoolMembership(tc, tx, false, false, false)
	}
}

// Attempt to remove orphans by tag,
// and ensure the state of all other orphans are unaffected.
func TestRemoveOrphansByTag(t *testing.T) {
	harness, _, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestRemoveOrphansByTag")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	orphanedTx1, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{1}, Index: 1},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(orphanedTx1, true,
		false, 1)
	orphanedTx2, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{2}, Index: 2},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(orphanedTx2, true,
		false, 1)
	orphanedTx3, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{3}, Index: 3},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(orphanedTx3, true,
		false, 1)

	orphanedTx4, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{4}, Index: 4},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	harness.txPool.ProcessTransaction(orphanedTx4, true,
		false, 2)

	harness.txPool.RemoveOrphansByTag(1)
	testPoolMembership(tc, orphanedTx1, false, false, false)
	testPoolMembership(tc, orphanedTx2, false, false, false)
	testPoolMembership(tc, orphanedTx3, false, false, false)
	testPoolMembership(tc, orphanedTx4, true, false, false)
}

// TestBasicOrphanRemoval ensure that orphan removal works as expected when an
// orphan that doesn't exist is removed  both when there is another orphan that
// redeems it and when there is not.
func TestBasicOrphanRemoval(t *testing.T) {
	const maxOrphans = 4
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestBasicOrphanRemoval")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	harness.txPool.cfg.Policy.MaxOrphanTxs = maxOrphans
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness.
	chainedTxns, err := harness.CreateTxChain(spendableOuts[0], maxOrphans+1)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Ensure the orphans are accepted (only up to the maximum allowed so
	// none are evicted).
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept valid "+
				"orphan %v", err)
		}

		// Ensure no transactions were reported as accepted.
		if len(acceptedTxns) != 0 {
			t.Fatalf("ProcessTransaction: reported %d accepted "+
				"transactions from what should be an orphan",
				len(acceptedTxns))
		}

		// Ensure the transaction is in the orphan pool, not in the
		// transaction pool, and reported as available.
		testPoolMembership(tc, tx, true, false, false)
	}

	// Attempt to remove an orphan that has no redeemers and is not present,
	// and ensure the state of all other orphans are unaffected.
	nonChainedOrphanTx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(5000000000),
		outPoint: wire.OutPoint{TxID: daghash.TxID{}, Index: 0},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}

	harness.txPool.RemoveOrphan(nonChainedOrphanTx)
	testPoolMembership(tc, nonChainedOrphanTx, false, false, false)
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		testPoolMembership(tc, tx, true, false, false)
	}

	// Attempt to remove an orphan that has a existing redeemer but itself
	// is not present and ensure the state of all other orphans (including
	// the one that redeems it) are unaffected.
	harness.txPool.RemoveOrphan(chainedTxns[0])
	testPoolMembership(tc, chainedTxns[0], false, false, false)
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		testPoolMembership(tc, tx, true, false, false)
	}

	// Remove each orphan one-by-one and ensure they are removed as
	// expected.
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		harness.txPool.RemoveOrphan(tx)
		testPoolMembership(tc, tx, false, false, false)
	}
}

// TestOrphanChainRemoval ensure that orphan chains (orphans that spend outputs
// from other orphans) are removed as expected.
func TestOrphanChainRemoval(t *testing.T) {
	const maxOrphans = 10
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestOrphanChainRemoval")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	harness.txPool.cfg.Policy.MaxOrphanTxs = maxOrphans
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness.
	chainedTxns, err := harness.CreateTxChain(spendableOuts[0], maxOrphans+1)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Ensure the orphans are accepted (only up to the maximum allowed so
	// none are evicted).
	for _, tx := range chainedTxns[1 : maxOrphans+1] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept valid "+
				"orphan %v", err)
		}

		// Ensure no transactions were reported as accepted.
		if len(acceptedTxns) != 0 {
			t.Fatalf("ProcessTransaction: reported %d accepted "+
				"transactions from what should be an orphan",
				len(acceptedTxns))
		}

		// Ensure the transaction is in the orphan pool, not in the
		// transaction pool, and reported as available.
		testPoolMembership(tc, tx, true, false, false)
	}

	// Remove the first orphan that starts the orphan chain without the
	// remove redeemer flag set and ensure that only the first orphan was
	// removed.
	harness.txPool.mtx.Lock()
	harness.txPool.removeOrphan(chainedTxns[1], false)
	harness.txPool.mtx.Unlock()
	testPoolMembership(tc, chainedTxns[1], false, false, false)
	for _, tx := range chainedTxns[2 : maxOrphans+1] {
		testPoolMembership(tc, tx, true, false, false)
	}

	// Remove the first remaining orphan that starts the orphan chain with
	// the remove redeemer flag set and ensure they are all removed.
	harness.txPool.mtx.Lock()
	harness.txPool.removeOrphan(chainedTxns[2], true)
	harness.txPool.mtx.Unlock()
	for _, tx := range chainedTxns[2 : maxOrphans+1] {
		testPoolMembership(tc, tx, false, false, false)
	}
}

// TestMultiInputOrphanDoubleSpend ensures that orphans that spend from an
// output that is spend by another transaction entering the pool are removed.
func TestMultiInputOrphanDoubleSpend(t *testing.T) {
	const maxOrphans = 4
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestMultiInputOrphanDoubleSpend")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	harness.txPool.cfg.Policy.MaxOrphanTxs = maxOrphans
	tc := &testContext{t, harness}

	// Create a chain of transactions rooted with the first spendable output
	// provided by the harness.
	chainedTxns, err := harness.CreateTxChain(outputs[0], maxOrphans+1)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}

	// Start by adding the orphan transactions from the generated chain
	// except the final one.
	for _, tx := range chainedTxns[1:maxOrphans] {
		acceptedTxns, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept valid "+
				"orphan %v", err)
		}
		if len(acceptedTxns) != 0 {
			t.Fatalf("ProcessTransaction: reported %d accepted transactions "+
				"from what should be an orphan", len(acceptedTxns))
		}
		testPoolMembership(tc, tx, true, false, false)
	}

	// Ensure a transaction that contains a double spend of the same output
	// as the second orphan that was just added as well as a valid spend
	// from that last orphan in the chain generated above (and is not in the
	// orphan pool) is accepted to the orphan pool.  This must be allowed
	// since it would otherwise be possible for a malicious actor to disrupt
	// tx chains.
	doubleSpendTx, err := harness.CreateSignedTx([]spendableOutpoint{
		txOutToSpendableOutpoint(chainedTxns[1], 0),
		txOutToSpendableOutpoint(chainedTxns[maxOrphans], 0),
	}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	acceptedTxns, err := harness.txPool.ProcessTransaction(doubleSpendTx,
		true, false, 0)
	if err != nil {
		t.Fatalf("ProcessTransaction: failed to accept valid orphan %v",
			err)
	}
	if len(acceptedTxns) != 0 {
		t.Fatalf("ProcessTransaction: reported %d accepted transactions "+
			"from what should be an orphan", len(acceptedTxns))
	}
	testPoolMembership(tc, doubleSpendTx, true, false, false)

	// Add the transaction which completes the orphan chain and ensure the
	// chain gets accepted.  Notice the accept orphans flag is also false
	// here to ensure it has no bearing on whether or not already existing
	// orphans in the pool are linked.
	//
	// This will cause the shared output to become a concrete spend which
	// will in turn must cause the double spending orphan to be removed.
	acceptedTxns, err = harness.txPool.ProcessTransaction(chainedTxns[0],
		false, false, 0)
	if err != nil {
		t.Fatalf("ProcessTransaction: failed to accept valid tx %v", err)
	}
	if len(acceptedTxns) != maxOrphans {
		t.Fatalf("ProcessTransaction: reported accepted transactions "+
			"length does not match expected -- got %d, want %d",
			len(acceptedTxns), maxOrphans)
	}
	for _, txD := range acceptedTxns {
		// Ensure the transaction is no longer in the orphan pool, is
		// in the transaction pool, and is reported as available.
		testPoolMembership(tc, txD.Tx, false, true, txD.Tx != chainedTxns[0])
	}

	// Ensure the double spending orphan is no longer in the orphan pool and
	// was not moved to the transaction pool.
	testPoolMembership(tc, doubleSpendTx, false, false, false)
}

// TestCheckSpend tests that CheckSpend returns the expected spends found in
// the mempool.
func TestCheckSpend(t *testing.T) {
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestCheckSpend")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()

	// The mempool is empty, so none of the spendable outputs should have a
	// spend there.
	for _, op := range outputs {
		spend := harness.txPool.CheckSpend(op.outPoint)
		if spend != nil {
			t.Fatalf("Unexpeced spend found in pool: %v", spend)
		}
	}

	// Create a chain of transactions rooted with the first spendable
	// output provided by the harness.
	const txChainLength = 5
	chainedTxns, err := harness.CreateTxChain(outputs[0], txChainLength)
	if err != nil {
		t.Fatalf("unable to create transaction chain: %v", err)
	}
	for _, tx := range chainedTxns {
		_, err := harness.txPool.ProcessTransaction(tx, true,
			false, 0)
		if err != nil {
			t.Fatalf("ProcessTransaction: failed to accept "+
				"tx: %v", err)
		}
	}

	// The first tx in the chain should be the spend of the spendable
	// output.
	op := outputs[0].outPoint
	spend := harness.txPool.CheckSpend(op)
	if spend != chainedTxns[0] {
		t.Fatalf("expected %v to be spent by %v, instead "+
			"got %v", op, chainedTxns[0], spend)
	}

	// Now all but the last tx should be spent by the next.
	for i := 0; i < len(chainedTxns)-1; i++ {
		op = wire.OutPoint{
			TxID:  *chainedTxns[i].ID(),
			Index: 0,
		}
		expSpend := chainedTxns[i+1]
		spend = harness.txPool.CheckSpend(op)
		if spend != expSpend {
			t.Fatalf("expected %v to be spent by %v, instead "+
				"got %v", op, expSpend, spend)
		}
	}

	// The last tx should have no spend.
	op = wire.OutPoint{
		TxID:  *chainedTxns[txChainLength-1].ID(),
		Index: 0,
	}
	spend = harness.txPool.CheckSpend(op)
	if spend != nil {
		t.Fatalf("Unexpeced spend found in pool: %v", spend)
	}
}

func TestCount(t *testing.T) {
	harness, outputs, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 1, "TestCount")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	if harness.txPool.Count() != 0 {
		t.Errorf("TestCount: txPool should be initialized with 0 transactions")
	}

	chainedTxns, err := harness.CreateTxChain(outputs[0], 3)
	if err != nil {
		t.Fatalf("harness.CreateTxChain: unexpected error: %v", err)
	}

	for i, tx := range chainedTxns {
		_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
		if err != nil {
			t.Errorf("ProcessTransaction: unexpected error: %v", err)
		}
		if harness.txPool.Count()+harness.txPool.DepCount() != i+1 {
			t.Errorf("TestCount: txPool expected to have %v transactions but got %v", i+1, harness.txPool.Count())
		}
	}

	err = harness.txPool.RemoveTransaction(chainedTxns[0], false, false)
	if err != nil {
		t.Fatalf("harness.CreateTxChain: unexpected error: %v", err)
	}
	if harness.txPool.Count()+harness.txPool.DepCount() != 2 {
		t.Errorf("TestCount: txPool expected to have 2 transactions but got %v", harness.txPool.Count())
	}
}

func TestExtractRejectCode(t *testing.T) {
	tests := []struct {
		blockdagRuleErrorCode blockdag.ErrorCode
		wireRejectCode        wire.RejectCode
	}{
		{
			blockdagRuleErrorCode: blockdag.ErrDuplicateBlock,
			wireRejectCode:        wire.RejectDuplicate,
		},
		{
			blockdagRuleErrorCode: blockdag.ErrBlockVersionTooOld,
			wireRejectCode:        wire.RejectObsolete,
		},
		{
			blockdagRuleErrorCode: blockdag.ErrCheckpointTimeTooOld,
			wireRejectCode:        wire.RejectCheckpoint,
		},
		{
			blockdagRuleErrorCode: blockdag.ErrDifficultyTooLow,
			wireRejectCode:        wire.RejectCheckpoint,
		},
		{
			blockdagRuleErrorCode: blockdag.ErrBadCheckpoint,
			wireRejectCode:        wire.RejectCheckpoint,
		},
		{
			blockdagRuleErrorCode: blockdag.ErrForkTooOld,
			wireRejectCode:        wire.RejectCheckpoint,
		},
		{
			blockdagRuleErrorCode: math.MaxUint32,
			wireRejectCode:        wire.RejectInvalid,
		},
	}

	for _, test := range tests {
		err := blockdag.RuleError{ErrorCode: test.blockdagRuleErrorCode}
		code, ok := extractRejectCode(err)
		if !ok {
			t.Errorf("TestExtractRejectCode: %v could not be extracted", test.blockdagRuleErrorCode)
		}
		if test.wireRejectCode != code {
			t.Errorf("TestExtractRejectCode: expected %v to extract %v but got %v", test.blockdagRuleErrorCode, test.wireRejectCode, code)
		}
	}

	txRuleError := TxRuleError{RejectCode: wire.RejectDust}
	txExtractedCode, ok := extractRejectCode(txRuleError)
	if !ok {
		t.Errorf("TestExtractRejectCode: %v could not be extracted", txRuleError)
	}
	if txExtractedCode != wire.RejectDust {
		t.Errorf("TestExtractRejectCode: expected %v to extract %v but got %v", wire.RejectDust, wire.RejectDust, txExtractedCode)
	}

	var nilErr error
	nilErrExtractedCode, ok := extractRejectCode(nilErr)
	if nilErrExtractedCode != wire.RejectInvalid {
		t.Errorf("TestExtractRejectCode: expected %v to extract %v but got %v", wire.RejectInvalid, wire.RejectInvalid, nilErrExtractedCode)
	}
	if ok {
		t.Errorf("TestExtractRejectCode: a nil error is expected to return false but got %v", ok)
	}

	nonRuleError := errors.New("nonRuleError")

	fErrExtractedCode, ok := extractRejectCode(nonRuleError)
	if fErrExtractedCode != wire.RejectInvalid {
		t.Errorf("TestExtractRejectCode: expected %v to extract %v but got %v", wire.RejectInvalid, wire.RejectInvalid, nilErrExtractedCode)
	}
	if ok {
		t.Errorf("TestExtractRejectCode: a nonRuleError is expected to return false but got %v", ok)
	}
}

// TestHandleNewBlock
func TestHandleNewBlock(t *testing.T) {
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&dagconfig.MainNetParams, 2, "TestHandleNewBlock")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	tc := &testContext{t, harness}

	// Create parent transaction for orphan transaction below
	blockTx1, err := harness.CreateSignedTx(spendableOuts[:1], 1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}

	// Create orphan transaction and add it to UTXO set
	txID := blockTx1.ID()
	orphanTx, err := harness.CreateSignedTx([]spendableOutpoint{{
		amount:   util.Amount(2500000000),
		outPoint: wire.OutPoint{TxID: *txID, Index: 0},
	}}, 1)
	if err != nil {
		t.Fatalf("unable to create signed tx: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(orphanTx, true, false, 0)
	if err != nil {
		t.Fatalf("ProcessTransaction: unexpected error: %v", err)
	}
	// ensure that transaction added to orphan pool
	testPoolMembership(tc, orphanTx, true, false, false)

	// Add one more transaction to block
	blockTx2, err := harness.CreateSignedTx(spendableOuts[1:], 1)
	if err != nil {
		t.Fatalf("unable to create transaction 1: %v", err)
	}
	dummyBlock.Transactions = append(dummyBlock.Transactions, blockTx1.MsgTx(), blockTx2.MsgTx())

	// Create block and add its transactions to UTXO set
	block := util.NewBlock(&dummyBlock)
	for i, tx := range block.Transactions() {
		if !harness.txPool.mpUTXOSet.AddTx(tx.MsgTx(), 1) {
			t.Fatalf("Failed to add transaction %v to UTXO set: %v", i, tx.ID())
		}
	}

	// Handle new block by pool
	ch := make(chan NewBlockMsg)
	go func() {
		err = harness.txPool.HandleNewBlock(block, ch)
		close(ch)
	}()

	// process messages pushed by HandleNewBlock
	blockTransnactions := make(map[daghash.TxID]int)
	for msg := range ch {
		blockTransnactions[*msg.Tx.ID()] = 1
		if *msg.Tx.ID() != *blockTx1.ID() {
			if len(msg.AcceptedTxs) != 0 {
				t.Fatalf("Expected amount of accepted transactions 0. Got: %v", len(msg.AcceptedTxs))
			}
		} else {
			if len(msg.AcceptedTxs) != 1 {
				t.Fatalf("Wrong accepted transactions length")
			}
			if *msg.AcceptedTxs[0].Tx.ID() != *orphanTx.ID() {
				t.Fatalf("Wrong accepted transaction ID")
			}
		}
	}
	// ensure that HandleNewBlock has not failed
	if err != nil {
		t.Fatalf("HandleNewBlock failed to handle block %v", err)
	}

	// Validate messages pushed by HandleNewBlock into the channel
	if len(blockTransnactions) != 2 {
		t.Fatalf("Wrong size of blockTransnactions after new block handling")
	}

	if _, ok := blockTransnactions[*blockTx1.ID()]; !ok {
		t.Fatalf("Transaction 1 of new block is not handled")
	}

	if _, ok := blockTransnactions[*blockTx2.ID()]; !ok {
		t.Fatalf("Transaction 2 of new block is not handled")
	}

	// ensure that orphan transaction moved to main pool
	testPoolMembership(tc, orphanTx, false, true, false)
}

// dummyBlock defines a block on the block DAG. It is used to test block operations.
var dummyBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version: 1,
		ParentHashes: []daghash.Hash{
			[32]byte{ // Make go vet happy.
				0x82, 0xdc, 0xbd, 0xe6, 0x88, 0x37, 0x74, 0x5b,
				0x78, 0x6b, 0x03, 0x1d, 0xa3, 0x48, 0x3c, 0x45,
				0x3f, 0xc3, 0x2e, 0xd4, 0x53, 0x5b, 0x6f, 0x26,
				0x26, 0xb0, 0x48, 0x4f, 0x09, 0x00, 0x00, 0x00,
			}, // MainNet genesis
			[32]byte{ // Make go vet happy.
				0xc1, 0x5b, 0x71, 0xfe, 0x20, 0x70, 0x0f, 0xd0,
				0x08, 0x49, 0x88, 0x1b, 0x32, 0xb5, 0xbd, 0x13,
				0x17, 0xbe, 0x75, 0xe7, 0x29, 0x46, 0xdd, 0x03,
				0x01, 0x92, 0x90, 0xf1, 0xca, 0x8a, 0x88, 0x11,
			}}, // SimNet genesis
		HashMerkleRoot: daghash.Hash([32]byte{ // Make go vet happy.
			0x66, 0x57, 0xa9, 0x25, 0x2a, 0xac, 0xd5, 0xc0,
			0xb2, 0x94, 0x09, 0x96, 0xec, 0xff, 0x95, 0x22,
			0x28, 0xc3, 0x06, 0x7c, 0xc3, 0x8d, 0x48, 0x85,
			0xef, 0xb5, 0xa4, 0xac, 0x42, 0x47, 0xe9, 0xf3,
		}), // f3e94742aca4b5ef85488dc37c06c3282295ffec960994b2c0d5ac2a25a95766
		Timestamp: time.Unix(1529483563, 0), // 2018-06-20 08:32:43 +0000 UTC
		Bits:      0x1e00ffff,               // 503382015
		Nonce:     0x000ae53f,               // 714047
	},
	Transactions: []*wire.MsgTx{
		{
			Version: 1,
			TxIn: []*wire.TxIn{
				{
					PreviousOutPoint: wire.OutPoint{
						TxID:  daghash.TxID{},
						Index: 0xffffffff,
					},
					SignatureScript: []byte{
						0x04, 0x4c, 0x86, 0x04, 0x1b, 0x02, 0x06, 0x02,
					},
					Sequence: math.MaxUint64,
				},
			},
			TxOut: []*wire.TxOut{
				{
					Value: 0x12a05f200, // 5000000000
					PkScript: []byte{
						0x41, // OP_DATA_65
						0x04, 0x1b, 0x0e, 0x8c, 0x25, 0x67, 0xc1, 0x25,
						0x36, 0xaa, 0x13, 0x35, 0x7b, 0x79, 0xa0, 0x73,
						0xdc, 0x44, 0x44, 0xac, 0xb8, 0x3c, 0x4e, 0xc7,
						0xa0, 0xe2, 0xf9, 0x9d, 0xd7, 0x45, 0x75, 0x16,
						0xc5, 0x81, 0x72, 0x42, 0xda, 0x79, 0x69, 0x24,
						0xca, 0x4e, 0x99, 0x94, 0x7d, 0x08, 0x7f, 0xed,
						0xf9, 0xce, 0x46, 0x7c, 0xb9, 0xf7, 0xc6, 0x28,
						0x70, 0x78, 0xf8, 0x01, 0xdf, 0x27, 0x6f, 0xdf,
						0x84, // 65-byte signature
						0xac, // OP_CHECKSIG
					},
				},
			},
			LockTime:     0,
			SubnetworkID: *subnetworkid.SubnetworkIDNative,
		},
	},
}

func TestTransactionGas(t *testing.T) {
	params := dagconfig.SimNetParams
	params.BlockRewardMaturity = 1
	harness, spendableOuts, teardownFunc, err := newPoolHarness(&params, 6, "TestTransactionGas")
	if err != nil {
		t.Fatalf("unable to create test pool: %v", err)
	}
	defer teardownFunc()
	//	tc := &testContext{t, harness}

	const gasLimit = 10000
	subnetworkID, err := testtools.RegisterSubnetworkForTest(harness.txPool.cfg.DAG, &params, gasLimit)
	if err != nil {
		t.Fatalf("unable to register network: %v", err)
	}

	// Create valid transaction
	tx, err := harness.CreateSignedTxForSubnetwork(spendableOuts[:1], 1, subnetworkID, gasLimit)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err != nil {
		t.Errorf("ProcessTransaction: unexpected error: %v", err)
	}

	// Create invalid transaction
	tx, err = harness.CreateSignedTxForSubnetwork(spendableOuts[1:], 1, subnetworkID, gasLimit+1)
	if err != nil {
		t.Fatalf("unable to create transaction: %v", err)
	}
	_, err = harness.txPool.ProcessTransaction(tx, true, false, 0)
	if err == nil {
		t.Error("ProcessTransaction did not return error, expecting ErrInvalidGas")
	}
}
