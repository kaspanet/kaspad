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
	outPoint0 := *wire.NewOutPoint(hash0, 0)
	outPoint1 := *wire.NewOutPoint(hash1, 0)
	utxoEntry0 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 10})
	utxoEntry1 := newUTXOEntry(&wire.TxOut{PkScript: []byte{}, Value: 20})

	tests := []struct {
		name           string
		collection     utxoCollection
		expectedString string
	}{
		{
			name:           "empty collection",
			collection:     utxoCollection{},
			expectedString: "[  ]",
		},
		{
			name: "one member",
			collection: utxoCollection{
				outPoint0: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 20 ]",
		},
		{
			name: "two members",
			collection: utxoCollection{
				outPoint0: utxoEntry0,
				outPoint1: utxoEntry1,
			},
			expectedString: "[ (0000000000000000000000000000000000000000000000000000000000000000, 0) => 10, (1111111111111111111111111111111111111111111111111111111111111111, 0) => 20 ]",
		},
	}

	for _, test := range tests {
		collectionString := test.collection.String()
		if collectionString != test.expectedString {
			t.Errorf("unexpected string in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, test.expectedString, collectionString)
		}
		collectionClone := test.collection.clone()
		if !reflect.DeepEqual(test.collection, collectionClone) {
			t.Errorf("collection is not equal to its clone in test \"%s\". "+
				"Expected: \"%s\", got: \"%s\".", test.name, collectionString, collectionClone.String())
		}
	}
}
