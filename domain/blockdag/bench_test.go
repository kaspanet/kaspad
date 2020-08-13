package blockdag

import (
	"strconv"
	"testing"

	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/network/domainmessage"
)

func generateNewUTXOEntry(index uint64) (domainmessage.Outpoint, *UTXOEntry) {
	txSuffix := strconv.FormatUint(index, 10)
	txStr := "0000000000000000000000000000000000000000000000000000000000000000"
	txID, _ := daghash.NewTxIDFromStr(txStr[0:len(txStr)-len(txSuffix)] + txSuffix)
	outpoint := *domainmessage.NewOutpoint(txID, 0)
	utxoEntry := NewUTXOEntry(&domainmessage.TxOut{ScriptPubKey: []byte{}, Value: index}, true, index)

	return outpoint, utxoEntry
}

func generateUtxoCollection(startIndex uint64, numItems uint64) utxoCollection {
	uc := make(utxoCollection)

	for i := uint64(0); i < numItems; i++ {
		outpoint, utxoEntry := generateNewUTXOEntry(startIndex + i)
		uc.add(outpoint, utxoEntry)
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
		this  *UTXODiff
		other *UTXODiff
	}{
		{
			this: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc2,
			},
			other: &UTXODiff{
				toAdd:    uc3,
				toRemove: uc4,
			},
		},
		{
			this: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc2,
			},
			other: &UTXODiff{
				toAdd:    uc3,
				toRemove: uc1,
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

// BenchmarkWithDiff performs a benchmark on how long it takes to apply provided diff to this diff
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
		this  *UTXODiff
		other *UTXODiff
	}{
		{
			this: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc2,
			},
			other: &UTXODiff{
				toAdd:    uc3,
				toRemove: uc4,
			},
		},
		{
			this: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc2,
			},
			other: &UTXODiff{
				toAdd:    uc3,
				toRemove: uc2,
			},
		},
		{
			this: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc2,
			},
			other: &UTXODiff{
				toAdd:    uc1,
				toRemove: uc3,
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
