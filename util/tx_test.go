// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package util_test

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util"
	"github.com/davecgh/go-spew/spew"
)

// TestTx tests the API for Tx.
func TestTx(t *testing.T) {
	firstTestTx := Block100000.Transactions[0]
	firstTx := util.NewTx(firstTestTx)
	secondTestTx := Block100000.Transactions[1]
	secondTx := util.NewTx(secondTestTx)

	// Ensure we get the same data back out.
	if msgTx := firstTx.MsgTx(); !reflect.DeepEqual(msgTx, firstTestTx) {
		t.Errorf("MsgTx: mismatched MsgTx - got %v, want %v",
			spew.Sdump(msgTx), spew.Sdump(firstTestTx))
	}

	// Ensure transaction index set and get work properly.
	wantIndex := 0
	firstTx.SetIndex(0)
	if gotIndex := firstTx.Index(); gotIndex != wantIndex {
		t.Errorf("Index: mismatched index - got %v, want %v",
			gotIndex, wantIndex)
	}

	// Hash for block 100,000 transaction 0.
	wantHashStr := "9bdfb2c83f82ab919e2e87cba505392bdfd9fa987864cbf59879c7dedf6da2cf"
	wantHash, err := daghash.NewHashFromStr(wantHashStr)
	if err != nil {
		t.Errorf("NewHashFromStr: %v", err)
	}

	// Request the hash multiple times to test generation and caching.
	for i := 0; i < 2; i++ {
		hash := firstTx.Hash()
		if !hash.IsEqual(wantHash) {
			t.Errorf("Hash #%d mismatched hash - got %v, want %v", i,
				hash, wantHash)
		}
	}

	// ID for block 100,000 transaction 1.
	wantIDStr := "011f7009d8e5a99c4cf2f7216a3eb6044a7017a98732a8fb6390fbbb668e84d8"
	wantID, err := daghash.NewTxIDFromStr(wantIDStr)
	// Request the ID multiple times to test generation and caching.
	for i := 0; i < 2; i++ {
		id := secondTx.ID()
		if !id.IsEqual(wantID) {
			t.Errorf("Hash #%d mismatched hash - got %v, want %v", i,
				id, wantID)
		}
	}
}

// TestNewTxFromBytes tests creation of a Tx from serialized bytes.
func TestNewTxFromBytes(t *testing.T) {
	// Serialize the test transaction.
	testTx := Block100000.Transactions[0]
	var testTxBuf bytes.Buffer
	err := testTx.Serialize(&testTxBuf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	testTxBytes := testTxBuf.Bytes()

	// Create a new transaction from the serialized bytes.
	tx, err := util.NewTxFromBytes(testTxBytes)
	if err != nil {
		t.Errorf("NewTxFromBytes: %v", err)
		return
	}

	// Ensure the generated MsgTx is correct.
	if msgTx := tx.MsgTx(); !reflect.DeepEqual(msgTx, testTx) {
		t.Errorf("MsgTx: mismatched MsgTx - got %v, want %v",
			spew.Sdump(msgTx), spew.Sdump(testTx))
	}
}

// TestTxErrors tests the error paths for the Tx API.
func TestTxErrors(t *testing.T) {
	// Serialize the test transaction.
	testTx := Block100000.Transactions[0]
	var testTxBuf bytes.Buffer
	err := testTx.Serialize(&testTxBuf)
	if err != nil {
		t.Errorf("Serialize: %v", err)
	}
	testTxBytes := testTxBuf.Bytes()

	// Truncate the transaction byte buffer to force errors.
	shortBytes := testTxBytes[:4]
	_, err = util.NewTxFromBytes(shortBytes)
	if err != io.EOF {
		t.Errorf("NewTxFromBytes: did not get expected error - "+
			"got %v, want %v", err, io.EOF)
	}
}
