package utxo

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"testing"
)

func TestUTXOEntry_Equal(t *testing.T) {
	type testUTXOEntryToCompare struct {
		utxoEntry      *utxoEntry
		expectedResult bool
	}

	tests := []struct {
		baseUTXOEntry        *utxoEntry
		UTXOEntryToCompareTo []testUTXOEntryToCompare
	}{
		{
			baseUTXOEntry: nil,
			UTXOEntryToCompareTo: []testUTXOEntryToCompare{
				{
					utxoEntry:      nil,
					expectedResult: true,
				},
				{
					utxoEntry: &utxoEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false,
					},
					expectedResult: false,
				},
			},
		}, {
			baseUTXOEntry: &utxoEntry{
				0xFFFF,
				[]byte{0xA1, 0xA2, 0xA3},
				0xFFFF,
				true,
			},
			UTXOEntryToCompareTo: []testUTXOEntryToCompare{
				{
					utxoEntry: &utxoEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						true,
					},
					expectedResult: true,
				},
				{
					utxoEntry:      nil,
					expectedResult: false,
				},
				{
					utxoEntry: &utxoEntry{
						0xFFFF,
						[]byte{0xA1, 0xA0, 0xA3}, // Changed
						0xFFFF,
						true,
					},
					expectedResult: false,
				},
				{
					utxoEntry: &utxoEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false, // Changed
					},
					expectedResult: false,
				},
				{
					utxoEntry: &utxoEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFF0, // Changed
						true,
					},
					expectedResult: false,
				},
				{
					utxoEntry:      nil,
					expectedResult: false,
				},
				{
					utxoEntry: &utxoEntry{
						0xFFF0, // Changed
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						true,
					},
					expectedResult: false,
				},
			},
		},
	}

	for i, test := range tests {
		for j, subTest := range test.UTXOEntryToCompareTo {
			var base externalapi.UTXOEntry = test.baseUTXOEntry
			result1 := base.Equal(subTest.utxoEntry)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.utxoEntry.Equal(base)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}
