package utxo

import (
	"strconv"
	"testing"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/daghash"
)

func generateNewUTXOEntry(index uint64) (appmessage.Outpoint, *Entry) {
	txSuffix := strconv.FormatUint(index, 10)
	txStr := "0000000000000000000000000000000000000000000000000000000000000000"
	txID, _ := daghash.NewTxIDFromStr(txStr[0:len(txStr)-len(txSuffix)] + txSuffix)
	outpoint := *appmessage.NewOutpoint(txID, 0)
	utxoEntry := NewEntry(&appmessage.TxOut{ScriptPubKey: []byte{}, Value: index}, true, index)

	return outpoint, utxoEntry
}

func generateUtxoCollection(startIndex uint64, numItems uint64) utxoCollection {
	uc := make(utxoCollection)

	for i := uint64(0); i < numItems; i++ {
		outpoint, utxoEntry := generateNewUTXOEntry(startIndex + i)
		uc.Add(outpoint, utxoEntry)
	}

	return uc
}

// BenchmarkDiffFrom performs a benchmark on how long it takes to calculate
// the difference between this utxoDiff and another one
func BenchmarkDiffFrom(b *testing.B) {
	var numOfEntries uint64 = 100
	var startIndex uint64 = 0
	uc1 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc2 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc3 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc4 := generateUtxoCollection(startIndex, numOfEntries)

	tests := []struct {
		this  *Diff
		other *Diff
	}{
		{
			this: &Diff{
				ToAdd:    uc1,
				ToRemove: uc2,
			},
			other: &Diff{
				ToAdd:    uc3,
				ToRemove: uc4,
			},
		},
		{
			this: &Diff{
				ToAdd:    uc1,
				ToRemove: uc2,
			},
			other: &Diff{
				ToAdd:    uc3,
				ToRemove: uc1,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			test.this.diffFrom(test.other)
		}

	}
}

// BenchmarkWithDiff performs a benchmark on how long it takes to apply provided Diff to this Diff
func BenchmarkWithDiff(b *testing.B) {
	var numOfEntries uint64 = 100
	var startIndex uint64 = 0
	uc1 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc2 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc3 := generateUtxoCollection(startIndex, numOfEntries)
	startIndex = startIndex + numOfEntries
	uc4 := generateUtxoCollection(startIndex, numOfEntries)

	tests := []struct {
		this  *Diff
		other *Diff
	}{
		{
			this: &Diff{
				ToAdd:    uc1,
				ToRemove: uc2,
			},
			other: &Diff{
				ToAdd:    uc3,
				ToRemove: uc4,
			},
		},
		{
			this: &Diff{
				ToAdd:    uc1,
				ToRemove: uc2,
			},
			other: &Diff{
				ToAdd:    uc3,
				ToRemove: uc2,
			},
		},
		{
			this: &Diff{
				ToAdd:    uc1,
				ToRemove: uc2,
			},
			other: &Diff{
				ToAdd:    uc1,
				ToRemove: uc3,
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			test.this.WithDiff(test.other)
		}

	}
}
