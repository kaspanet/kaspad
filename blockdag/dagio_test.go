// Copyright (c) 2015-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/database"
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

// TestUtxoSerialization ensures serializing and deserializing unspent
// trasaction output entries works as expected.
func TestUtxoSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entry      *UTXOEntry
		serialized []byte
	}{
		// From tx in main blockchain:
		// b7c3332bc138e2c9429818f5fed500bcc1746544218772389054dc8047d7cd3f:0
		{
			name: "height 1, coinbase",
			entry: &UTXOEntry{
				amount:           5000000000,
				pkScript:         hexToBytes("410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac"),
				blockChainHeight: 1,
				packedFlags:      tfBlockReward,
			},
			serialized: hexToBytes("03320496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52"),
		},
		// From tx in main blockchain:
		// 8131ffb0a2c945ecaf9b9063e59558784f9c3a74741ce6ae2a18d0571dac15bb:1
		{
			name: "height 100001, not coinbase",
			entry: &UTXOEntry{
				amount:           1000000,
				pkScript:         hexToBytes("76a914ee8bd501094a7d5ca318da2506de35e1cb025ddc88ac"),
				blockChainHeight: 100001,
				packedFlags:      0,
			},
			serialized: hexToBytes("8b99420700ee8bd501094a7d5ca318da2506de35e1cb025ddc"),
		},
	}

	for i, test := range tests {
		// Ensure the utxo entry serializes to the expected value.
		gotBytes, err := serializeUTXOEntry(test.entry)
		if err != nil {
			t.Errorf("serializeUTXOEntry #%d (%s) unexpected "+
				"error: %v", i, test.name, err)
			continue
		}
		if !bytes.Equal(gotBytes, test.serialized) {
			t.Errorf("serializeUTXOEntry #%d (%s): mismatched "+
				"bytes - got %x, want %x", i, test.name,
				gotBytes, test.serialized)
			continue
		}

		// Deserialize to a utxo entry.
		utxoEntry, err := deserializeUTXOEntry(test.serialized)
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

		if !bytes.Equal(utxoEntry.PkScript(), test.entry.PkScript()) {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"scripts: got %x, want %x", i, test.name,
				utxoEntry.PkScript(), test.entry.PkScript())
			continue
		}
		if utxoEntry.BlockChainHeight() != test.entry.BlockChainHeight() {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"block height: got %d, want %d", i, test.name,
				utxoEntry.BlockChainHeight(), test.entry.BlockChainHeight())
			continue
		}
		if utxoEntry.IsBlockReward() != test.entry.IsBlockReward() {
			t.Errorf("deserializeUTXOEntry #%d (%s) mismatched "+
				"coinbase flag: got %v, want %v", i, test.name,
				utxoEntry.IsBlockReward(), test.entry.IsBlockReward())
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
		errType    error
	}{
		{
			name:       "no data after header code",
			serialized: hexToBytes("02"),
			errType:    errDeserialize(""),
		},
		{
			name:       "incomplete compressed txout",
			serialized: hexToBytes("0232"),
			errType:    errDeserialize(""),
		},
	}

	for _, test := range tests {
		// Ensure the expected error type is returned and the returned
		// entry is nil.
		entry, err := deserializeUTXOEntry(test.serialized)
		if reflect.TypeOf(err) != reflect.TypeOf(test.errType) {
			t.Errorf("deserializeUTXOEntry (%s): expected error "+
				"type does not match - got %T, want %T",
				test.name, err, test.errType)
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
			serialized: []byte("{\"TipHashes\":[[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0]],\"LastFinalityPoint\":[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0]}"),
		},
		{
			name: "block 1",
			state: &dagState{
				TipHashes:         []*daghash.Hash{newHashFromStr("00000000839a8e6886ab5951d76f411475428afc90947ee320161bbf18eb6048")},
				LastFinalityPoint: newHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f"),
			},
			serialized: []byte("{\"TipHashes\":[[72,96,235,24,191,27,22,32,227,126,148,144,252,138,66,117,20,65,111,215,81,89,171,134,104,142,154,131,0,0,0,0]],\"LastFinalityPoint\":[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,79,147,30,131,101,225,90,8,156,104,214,25,0,0,0,0,0]}"),
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

// TestDAGStateDeserializeErrors performs negative tests against
// deserializing the DAG state to ensure error paths work as expected.
func TestDAGStateDeserializeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		serialized []byte
		errType    error
	}{
		{
			name:       "nothing serialized",
			serialized: hexToBytes(""),
			errType:    database.Error{ErrorCode: database.ErrCorruption},
		},
		{
			name:       "corrupted data",
			serialized: []byte("[[111,226,140,10,182,241,179,114,193,166,162,70,174,99,247,7"),
			errType:    database.Error{ErrorCode: database.ErrCorruption},
		},
	}

	for _, test := range tests {
		// Ensure the expected error type and code is returned.
		_, err := deserializeDAGState(test.serialized)
		if reflect.TypeOf(err) != reflect.TypeOf(test.errType) {
			t.Errorf("deserializeDAGState (%s): expected "+
				"error type does not match - got %T, want %T",
				test.name, err, test.errType)
			continue
		}
		if derr, ok := err.(database.Error); ok {
			tderr := test.errType.(database.Error)
			if derr.ErrorCode != tderr.ErrorCode {
				t.Errorf("deserializeDAGState (%s): "+
					"wrong error code got: %v, want: %v",
					test.name, derr.ErrorCode,
					tderr.ErrorCode)
				continue
			}
		}
	}
}
