// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mining

import (
	"github.com/pkg/errors"
	"testing"

	"github.com/daglabs/kaspad/util/subnetworkid"

	"bou.ke/monkey"
	"github.com/daglabs/kaspad/blockdag"
	"github.com/daglabs/kaspad/dagconfig"
	"github.com/daglabs/kaspad/txscript"
	"github.com/daglabs/kaspad/util/daghash"
	"github.com/daglabs/kaspad/wire"

	"github.com/daglabs/kaspad/util"
)

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

	scriptPubKey, err := txscript.NewScriptBuilder().AddOp(txscript.OpTrue).Script()
	if err != nil {
		t.Fatalf("Failed to create scriptPubKey: %v", err)
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
		ScriptPubKey: scriptPubKey,
		Value:        1,
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
		ScriptPubKey: scriptPubKey,
		Value:        1,
	}
	nonFinalizedTx := wire.NewNativeMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut})
	nonFinalizedTx.LockTime = dag.ChainHeight() + 2

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
		ScriptPubKey: scriptPubKey,
		Value:        1,
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
		ScriptPubKey: scriptPubKey,
		Value:        1,
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
		ScriptPubKey: scriptPubKey,
		Value:        1,
	}
	subnetworkTx2 := wire.NewSubnetworkMsgTx(wire.TxVersion, []*wire.TxIn{txIn}, []*wire.TxOut{txOut}, existingSubnetwork,
		100, // Subnetwork gas limit is 90
		[]byte{})

	txSource.txDescs = []*TxDesc{
		{
			Tx:  util.NewTx(tx),
			Fee: 1,
		},
		{
			Tx:  util.NewTx(nonFinalizedTx),
			Fee: 1,
		},
		{
			Tx:  util.NewTx(subnetworkTx1),
			Fee: 1,
		},
		{
			Tx:  util.NewTx(subnetworkTx2),
			Fee: 1,
		},
		{
			Tx:  util.NewTx(nonExistingSubnetworkTx),
			Fee: 1,
		},
	}

	// Here we define nonExistingSubnetwork to be non-exist, and existingSubnetwork to have a gas limit of 90
	gasLimitPatch := monkey.Patch((*blockdag.SubnetworkStore).GasLimit, func(_ *blockdag.SubnetworkStore, subnetworkID *subnetworkid.SubnetworkID) (uint64, error) {
		if subnetworkID.IsEqual(nonExistingSubnetwork) {
			return 0, errors.New("not found")
		}
		return 90, nil
	})
	defer gasLimitPatch.Unpatch()

	template3, err := blockTemplateGenerator.NewBlockTemplate(OpTrueAddr)
	gasLimitPatch.Unpatch()

	if err != nil {
		t.Errorf("NewBlockTemplate: unexpected error: %v", err)
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
