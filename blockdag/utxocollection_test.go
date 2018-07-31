package blockdag

import (
	"testing"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/wire"
	"reflect"
)

func TestUTXOCollection(t *testing.T) {
	hash0, _ := daghash.NewHashFromStr("0000000000000000000000000000000000000000000000000000000000000000")
	hash1, _ := daghash.NewHashFromStr("1111111111111111111111111111111111111111111111111111111111111111")
	txOut0 := &wire.TxOut{PkScript: []byte{}, Value: 10}
	txOut1 := &wire.TxOut{PkScript: []byte{}, Value: 20}

	tests := []struct {
		name            string
		toAdd           []utxoIteratorOutput
		toRemove        []utxoIteratorOutput
		expectedMembers []utxoIteratorOutput
		expectedString  string
	}{
		{
			name:            "empty collection",
			toAdd:           []utxoIteratorOutput{},
			toRemove:        []utxoIteratorOutput{},
			expectedMembers: []utxoIteratorOutput{},
			expectedString:  "[  ]",
		},
		{
			name: "add one member",
			toAdd: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
			},
			toRemove:        []utxoIteratorOutput{},
			expectedMembers: []utxoIteratorOutput{{hash: *hash0, index: 0, txOut: txOut0}},
			expectedString:  "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10 ]",
		},
		{
			name:  "remove a member from an empty collection",
			toAdd: []utxoIteratorOutput{},
			toRemove: []utxoIteratorOutput{
				{hash: *hash0, index: 0},
			},
			expectedMembers: []utxoIteratorOutput{},
			expectedString:  "[  ]",
		},
		{
			name: "add one member and then remove it",
			toAdd: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
			},
			toRemove: []utxoIteratorOutput{
				{hash: *hash0, index: 0},
			},
			expectedMembers: []utxoIteratorOutput{},
			expectedString:  "[  ]",
		},
		{
			name: "add two members with the same hash",
			toAdd: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
				{hash: *hash0, index: 1, txOut: txOut1},
			},
			toRemove: []utxoIteratorOutput{},
			expectedMembers: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
				{hash: *hash0, index: 1, txOut: txOut1},
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, (0000000000000000000000000000000000000000000000000000000000000000, 1) => 20 ]",
		},
		{
			name: "add two members with the different hashes",
			toAdd: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
				{hash: *hash1, index: 0, txOut: txOut1},
			},
			toRemove: []utxoIteratorOutput{},
			expectedMembers: []utxoIteratorOutput{
				{hash: *hash0, index: 0, txOut: txOut0},
				{hash: *hash1, index: 0, txOut: txOut1},
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ]",
		},
	}

	for _, test := range tests {
		collection := make(utxoCollection)
		for _, member := range test.expectedMembers {
			if collection.contains(member.hash, 0) {
				t.Errorf("empty collection unexpectedly contains a member")
			}
			if txOut, ok := collection.get(member.hash, 0); ok || txOut != nil {
				t.Errorf("empty collection returned a member")
			}
		}

		for _, utxo := range test.toAdd {
			collection.add(utxo.hash, utxo.index, utxo.txOut)
		}
		for _, utxo := range test.toRemove {
			collection.remove(utxo.hash, utxo.index)
		}

		for _, member := range test.expectedMembers {
			if !collection.contains(member.hash, member.index) {
				t.Errorf("missing member in test \"%s\". "+
					"Missing: %v", test.name, member)
			}
			txOut, _ := collection.get(member.hash, member.index)
			if txOut != member.txOut {
				t.Errorf("unexpected member got in test \"%s\". "+
					"Expected: %v, got: %v.", test.name, member.txOut, txOut)
			}
		}

		expectedLength := len(test.expectedMembers)
		if collection.len() != expectedLength {
			t.Errorf("unexpected length in test \"%s\". "+
				"Expected: %d, got: %d.", test.name, expectedLength, collection.len())
		}
		collectionString := collection.String()
		if collectionString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, collectionString)
		}
		collectionClone := collection.clone()
		if !reflect.DeepEqual(collection, collectionClone) {
			t.Errorf("collection is not equal to its clone in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, collectionString, collectionClone.String())
		}
	}
}
