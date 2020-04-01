// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"bytes"
	"encoding/hex"
	"github.com/pkg/errors"
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/util/daghash"
)

// TestErrNotInDAG ensures the functions related to errNotInDAG work
// as expected.
func TestErrNotInDAG(t *testing.T) {
	errStr := "no block at height 1 exists"
	err := error(errNotInDAG(errStr))

	// Ensure the stringized output for the error is as expected.
	if err.Error() != errStr {
		t.Fatalf("errNotInDAG retuned unexpected error string - "+
			"got %q, want %q", err.Error(), errStr)
	}

	// Ensure error is detected as the correct type.
	if !isNotInDAGErr(err) {
		t.Fatalf("isNotInDAGErr did not detect as expected type")
	}
	err = errors.New("something else")
	if isNotInDAGErr(err) {
		t.Fatalf("isNotInDAGErr detected incorrect type")
	}
}

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error. This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// TestUTXOSerialization ensures serializing and deserializing unspent
// trasaction output entries works as expected.
func TestUTXOSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entry      *UTXOEntry
		serialized []byte
	}{
		{
			name: "blue score 1, coinbase",
			entry: &UTXOEntry{
				amount:         5000000000,
				scriptPubKey:   hexToBytes("410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac"),
				blockBlueScore: 1,
				packedFlags:    tfCoinbase,
			},
			serialized: hexToBytes("030000000000000000f2052a0100000043410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac"),
		},
		{
			name: "blue score 100001, not coinbase",
			entry: &UTXOEntry{
				amount:         1000000,
				scriptPubKey:   hexToBytes("76a914ee8bd501094a7d5ca318da2506de35e1cb025ddc88ac"),
				blockBlueScore: 100001,
				packedFlags:    0,
			},
			serialized: hexToBytes("420d03000000000040420f00000000001976a914ee8bd501094a7d5ca318da2506de35e1cb025ddc88ac"),
		},
	}

	for i, test := range tests {
		// Ensure the utxo entry serializes to the expected value.
		w := &bytes.Buffer{}
		err := serializeUTXOEntry(w, test.entry)
		if err != nil {
			t.Errorf("serializeUTXOEntry #%d (%s) unexpected "+
				"error: %v", i, test.name, err)
			continue
		}

		gotBytes := w.Bytes()
		if !bytes.Equal(gotBytes, test.serialized) {
			t.Errorf("serializeUTXOEntry #%d (%s): mismatched "+
				"bytes - got %x, want %x", i, test.name,
				gotBytes, test.serialized)
			continue
		}

		// Deserialize to a utxo entry.gotBytes
		utxoEntry, err := deserializeUTXOEntry(bytes.NewReader(test.serialized))
		if err != nil {
			t.Errorf("deserializeUTXOEntry #%d (%s) unexpected "+
				"error: %v", i, test.name, err)
			continue
		}

		// Ensure the deserialized entry has the same properties as the
		// ones in the test entry.
		if utxoEntry.Amount() != test.entry.Amount() {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"amounts: got %d, want %d", i, test.name,
				utxoEntry.Amount(), test.entry.Amount())
			continue
		}

		if !bytes.Equal(utxoEntry.ScriptPubKey(), test.entry.ScriptPubKey()) {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"scripts: got %x, want %x", i, test.name,
				utxoEntry.ScriptPubKey(), test.entry.ScriptPubKey())
			continue
		}
		if utxoEntry.BlockBlueScore() != test.entry.BlockBlueScore() {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"block blue score: got %d, want %d", i, test.name,
				utxoEntry.BlockBlueScore(), test.entry.BlockBlueScore())
			continue
		}
		if utxoEntry.IsCoinbase() != test.entry.IsCoinbase() {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"coinbase flag: got %v, want %v", i, test.name,
				utxoEntry.IsCoinbase(), test.entry.IsCoinbase())
			continue
		}
	}
}

// TestUtxoEntryDeserializeErrors performs negative tests against deserializing
// unspent transaction outputs to ensure error paths work as expected.
func TestUtxoEntryDeserializeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serialized []byte
	}{
		{
			name:       "no data after header code",
			serialized: hexToBytes("02"),
		},
		{
			name:       "incomplete compressed txout",
			serialized: hexToBytes("0232"),
		},
	}

	for _, test := range tests {
		// Ensure the expected error type is returned and the returned
		// entry is nil.
		entry, err := deserializeUTXOEntry(bytes.NewReader(test.serialized))
		if err == nil {
			t.Errorf("deserializeUTXOEntry (%s): didn't return an error",
				test.name)
			continue
		}
		if entry != nil {
			t.Errorf("deserializeUTXOEntry (%s): returned entry "+
				"is not nil", test.name)
			continue
		}
	}
}

// TestDAGStateSerialization ensures serializing and deserializing the
// DAG state works as expected.
func TestDAGStateSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		state      *dagState
		serialized []byte
	}{
		{
			name: "genesis",
			state: &dagState{
				TipHashes:         []*daghash.Hash{newHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")},
				LastFinalityPoint: newHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},
			serialized: []byte("{\"TipHashes\":[[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0]],\"LastFinalityPoint\":[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0],\"LocalSubnetworkID\":null}"),
		},
		{
			name: "block 1",
			state: &dagState{
				TipHashes:         []*daghash.Hash{newHashFromStr("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")},
				LastFinalityPoint: newHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},
			serialized: []byte("{\"TipHashes\":[[72,96,235,24,191,27,22,32,227,126,148,144,252,138,66,117,20,65,111,215,81,89,171,134,104,142,154,131,0,0,0,0]],\"LastFinalityPoint\":[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0],\"LocalSubnetworkID\":null}"),
		},
	}

	for i, test := range tests {
		gotBytes, err := serializeDAGState(test.state)
		if err != nil {
			t.Errorf("serializeDAGState #%d (%s) "+
				"unexpected error: %v", i, test.name, err)
			continue
		}

		// Ensure the dagState serializes to the expected value.
		if !bytes.Equal(gotBytes, test.serialized) {
			t.Errorf("serializeDAGState #%d (%s): mismatched "+
				"bytes - got %s, want %s", i, test.name,
				string(gotBytes), string(test.serialized))
			continue
		}

		// Ensure the serialized bytes are decoded back to the expected
		// dagState.
		state, err := deserializeDAGState(test.serialized)
		if err != nil {
			t.Errorf("deserializeDAGState #%d (%s) "+
				"unexpected error: %v", i, test.name, err)
			continue
		}
		if !reflect.DeepEqual(state, test.state) {
			t.Errorf("deserializeDAGState #%d (%s) "+
				"mismatched state - got %v, want %v", i,
				test.name, state, test.state)
			continue
		}
	}
}

// newHashFromStr converts the passed big-endian hex string into a
// daghash.Hash. It only differs from the one available in daghash in that
// it panics in case of an error since it will only (and must only) be
// called with hard-coded, and therefore known good, hashes.
func newHashFromStr(hexStr string) *daghash.Hash {
	hash, err := daghash.NewHashFromStr(hexStr)
	if err != nil {
		panic(err)
	}
	return hash
}
