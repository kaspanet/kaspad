package externalapi

import (
	"reflect"
	"testing"
)

func initTestUTXOEntryForClone() []*UTXOEntry {

	tests := []*UTXOEntry{
		{0xFFFF,
			[]byte{0xA1, 0xA2, 0xA3},
			0xFFFF,
			true},
		{0x0000,
			[]byte{0, 0, 0},
			0x0000,
			false}}
	return tests
}

type TestUTXOEntryToCompare struct {
	utxoEntry      *UTXOEntry
	expectedResult bool
}

type TestUTXOEntryStruct struct {
	baseUTXOEntry        *UTXOEntry
	UTXOEntryToCompareTo []TestUTXOEntryToCompare
}

func initTestUTXOEntryForEqual() []*TestUTXOEntryStruct {
	tests := []*TestUTXOEntryStruct{
		{
			baseUTXOEntry: nil,
			UTXOEntryToCompareTo: []TestUTXOEntryToCompare{
				{
					utxoEntry: &UTXOEntry{0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false},
					expectedResult: false,
				}, {
					utxoEntry:      nil,
					expectedResult: true,
				},
			},
		}, {
			baseUTXOEntry: &UTXOEntry{0xFFFF,
				[]byte{0xA1, 0xA2, 0xA3},
				0xFFFF,
				true},
			UTXOEntryToCompareTo: []TestUTXOEntryToCompare{
				{
					utxoEntry: &UTXOEntry{0xFFFF,
						[]byte{0xA1, 0xA0, 0xA3},
						0xFFFF,
						true},
					expectedResult: false,
				}, {
					utxoEntry: &UTXOEntry{0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFFF,
						false},
					expectedResult: false,
				}, {
					utxoEntry: &UTXOEntry{0xFFFF,
						[]byte{0xA1, 0xA2, 0xA3},
						0xFFF0,
						true},
					expectedResult: false,
				}, {
					utxoEntry:      nil,
					expectedResult: false,
				}, {
					utxoEntry: &UTXOEntry{0xFFF0,
						[]byte{0xA1, 0xA2, 0xA3},
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

	testSyncState := initTestUTXOEntryForEqual()
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

	testUTXOEntry := initTestUTXOEntryForClone()
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
