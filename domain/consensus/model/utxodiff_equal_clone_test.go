package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"
)

func initTestUTXOCollectionForClone() []UTXOCollection {

	var testUTXOCollection1 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{}
	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	tests := []UTXOCollection{
		testUTXOCollection1,
	}
	return tests
}

type TestUTXOCollectionToCompare struct {
	utxoCollection UTXOCollection
	expectedResult bool
}

type TestUTXOCollectionStruct struct {
	baseUTXOCollection        UTXOCollection
	utxoCollectionToCompareTo []TestUTXOCollectionToCompare
}

func initTestUTXOCollectionForEqual() []TestUTXOCollectionStruct {

	var testUTXOCollection1 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{}
	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	var testUTXOCollection2 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{}

	testUTXOCollection2[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection2[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	var testUTXOCollection3 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{} // map.size()==3
	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x03}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA5},
		0xFFFF,
		true}

	var testUTXOCollection4 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{} // second elem key
	testUTXOCollection4[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection4[externalapi.DomainOutpoint{ //DomainTransactionID is diff to the base
		externalapi.DomainTransactionID{0x04}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	var testUTXOCollection5 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{} //second element
	testUTXOCollection5[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection5[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA5}, //second element (count from zero) is different to the base
		0xFFFF,
		true}

	tests := []TestUTXOCollectionStruct{
		{
			baseUTXOCollection: testUTXOCollection1,
			utxoCollectionToCompareTo: []TestUTXOCollectionToCompare{
				{
					utxoCollection: testUTXOCollection2,
					expectedResult: true,
				}, {
					utxoCollection: testUTXOCollection3,
					expectedResult: false,
				}, {
					utxoCollection: testUTXOCollection4,
					expectedResult: false,
				}, {
					utxoCollection: testUTXOCollection5,
					expectedResult: false,
				},
			},
		},
	}
	return tests
}

func TestUTXOCollection_Equal(t *testing.T) {

	utxoCollections := initTestUTXOCollectionForEqual()
	for i, test := range utxoCollections {
		for j, subTest := range test.utxoCollectionToCompareTo {
			result1 := test.baseUTXOCollection.Equal(subTest.utxoCollection)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.utxoCollection.Equal(test.baseUTXOCollection)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestUTXOCollection_Clone(t *testing.T) {

	testUTXOCollection := initTestUTXOCollectionForClone()
	for i, utxoCollection := range testUTXOCollection {
		utxoCollectionClone := utxoCollection.Clone()
		if !utxoCollectionClone.Equal(utxoCollection) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(utxoCollection, utxoCollectionClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}

func initTestUTXODiffForClone() []UTXODiff {

	var testUTXOCollection1 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{}
	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection1[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	var testUTXOCollection3 UTXOCollection = map[externalapi.DomainOutpoint]*externalapi.UTXOEntry{} // map.size()==3
	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x01}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA3},
		0xFFFF,
		true}

	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x02}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA4},
		0xFFFF,
		true}

	testUTXOCollection3[externalapi.DomainOutpoint{
		externalapi.DomainTransactionID{0x03}, 0xFFFF}] = &externalapi.UTXOEntry{0xFFFF,
		[]byte{0xA1, 0xA2, 0xA5},
		0xFFFF,
		true}

	tests := []UTXODiff{
		{
			testUTXOCollection1,
			testUTXOCollection3},
	}
	return tests
}

func TestUTXODiff_Clone(t *testing.T) {

	testUTXODiff := initTestUTXODiffForClone()
	for i, utxoDiff := range testUTXODiff {
		utxoDiffClone := utxoDiff.Clone()
		if !reflect.DeepEqual(utxoDiffClone, &utxoDiff) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
