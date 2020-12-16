package externalapi

import (
	"reflect"
	"testing"
)

func TestUTXOEntry_Equal(t *testing.T) {
	type testUTXOEntryToCompare struct {
		utxoEntry      *UTXOEntry
		expectedResult bool
	}

	testSyncState := []struct {
		baseUTXOEntry        *UTXOEntry
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
					utxoEntry: &UTXOEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false,
					},
					expectedResult: false,
				},
			},
		}, {
			baseUTXOEntry: &UTXOEntry{
				0xFFFF,
				[]byte{0xA1, 0xA2, 0xA3},
				0xFFFF,
				true,
			},
			UTXOEntryToCompareTo: []testUTXOEntryToCompare{
				{
					utxoEntry: &UTXOEntry{
						0xFFFF,
						[]byte{0xA1, 0xA0, 0xA3}, // Changed
						0xFFFF,
						true,
					},
					expectedResult: false,
				},
				{
					utxoEntry: &UTXOEntry{
						0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false, // Changed
					},
					expectedResult: false,
				},
				{
					utxoEntry: &UTXOEntry{
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
					utxoEntry: &UTXOEntry{
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

	for i, test := range testSyncState {
		for j, subTest := range test.UTXOEntryToCompareTo {
			result1 := test.baseUTXOEntry.Equal(subTest.utxoEntry)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.utxoEntry.Equal(test.baseUTXOEntry)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestUTXOEntry_Clone(t *testing.T) {
	testUTXOEntry := []*UTXOEntry{
		{
			0xFFFF,
			[]byte{0xA1, 0xA2, 0xA3},
			0xFFFF,
			true,
		},
		{
			0x0000,
			[]byte{0, 0, 0},
			0x0000,
			false,
		},
	}

	for i, utxoEntry := range testUTXOEntry {
		utxoEntryClone := utxoEntry.Clone()
		if !utxoEntryClone.Equal(utxoEntry) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(utxoEntry, utxoEntryClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
