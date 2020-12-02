package externalapi

import (
	"reflect"
	"testing"
)

//UTXOEntry
/*
Amount          uint64
	ScriptPublicKey []byte // The public key script for the output.
	BlockBlueScore  uint64 // Blue score of the block accepting the tx.
	IsCoinbase      bool
*/

func InitTestUTXOEntryForClone() []UTXOEntry {

	tests := []UTXOEntry{
		{0xFFFF,
			[]byte{'a', 'b', 'c'},
			0xFFFF,
			true},
		{0x0000,
			[]byte{0, 0, 0},
			0x0000,
			false}}

	return tests
}

type TestUTXOEntryToCompare struct {
	uTXOEntry      *UTXOEntry
	expectedResult bool
}

type TestUTXOEntryStruct struct {
	baseUTXOEntry        *UTXOEntry
	UTXOEntryToCompareTo []TestUTXOEntryToCompare
}

func InitTestUTXOEntryForEqual() []*TestUTXOEntryStruct {
	tests := []*TestUTXOEntryStruct{
		{
			baseUTXOEntry: nil,
			UTXOEntryToCompareTo: []TestUTXOEntryToCompare{
				{
					uTXOEntry: &UTXOEntry{0xFFFF,
						[]byte{'a', 'b', 'c'},
						0xFFFF,
						false},
					expectedResult: false,
				},
				{
					uTXOEntry:      nil,
					expectedResult: true,
				},
			},
		},
		{
			baseUTXOEntry: &UTXOEntry{0xFFFF,
				[]byte{'a', 'b', 'c'},
				0xFFFF,
				true},

			UTXOEntryToCompareTo: []TestUTXOEntryToCompare{
				{
					uTXOEntry: &UTXOEntry{0xFFFF,
						[]byte{'a', 'b', 'c'},
						0xFFFF,
						true},
					expectedResult: true,
				},
				{
					uTXOEntry:      nil,
					expectedResult: false,
				},
				{
					uTXOEntry: &UTXOEntry{0xFFFF,
						[]byte{'a', 0, 'c'},
						0xFFFF,
						true},
					expectedResult: false,
				},
			},
		},
	}
	return tests
}

func TestUTXOEntry_Equal(t *testing.T) {

	testSyncState := InitTestUTXOEntryForEqual()

	for i, test := range testSyncState {
		for j, subTest := range test.UTXOEntryToCompareTo {
			result1 := test.baseUTXOEntry.Equal(subTest.uTXOEntry)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.uTXOEntry.Equal(test.baseUTXOEntry)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestUTXOEntry_Clone(t *testing.T) {

	testUTXOEntry := InitTestUTXOEntryForClone()

	for i, uTXOEntry := range testUTXOEntry {
		clone := uTXOEntry.Clone()
		if !clone.Equal(&uTXOEntry) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(uTXOEntry, clone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
