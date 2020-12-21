package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"
)

func initTestBlockRelationsForClone() []*BlockRelations {

	tests := []*BlockRelations{
		{
			[]*externalapi.DomainHash{{1}, {2}},
			[]*externalapi.DomainHash{{3}, {4}},
		},
	}
	return tests
}

type testBlockRelationsToCompare struct {
	blockRelations *BlockRelations
	expectedResult bool
}

type testBlockRelationsStruct struct {
	baseBlockRelations        *BlockRelations
	blockRelationsToCompareTo []testBlockRelationsToCompare
}

func initTestBlockRelationsForEqual() []testBlockRelationsStruct {

	var testBlockRelationsBase = BlockRelations{
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{3}, {4}},
	}
	//First test: structs are equal
	var testBlockRelations1 = BlockRelations{
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{3}, {4}},
	}
	//Second test: children changed
	var testBlockRelations2 = BlockRelations{
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{3}, {5}},
	}
	//Third test: parents changed
	var testBlockRelations3 = BlockRelations{
		[]*externalapi.DomainHash{{6}, {2}},
		[]*externalapi.DomainHash{{3}, {4}},
	}

	tests := []testBlockRelationsStruct{
		{
			baseBlockRelations: &testBlockRelationsBase,
			blockRelationsToCompareTo: []testBlockRelationsToCompare{
				{
					blockRelations: &testBlockRelations1,
					expectedResult: true,
				}, {
					blockRelations: &testBlockRelations2,
					expectedResult: false,
				}, {
					blockRelations: &testBlockRelations3,
					expectedResult: false,
				}, {
					blockRelations: nil,
					expectedResult: false,
				},
			},
		}, {
			baseBlockRelations: nil,
			blockRelationsToCompareTo: []testBlockRelationsToCompare{
				{
					blockRelations: &testBlockRelations1,
					expectedResult: false,
				}, {
					blockRelations: nil,
					expectedResult: true,
				},
			},
		},
	}
	return tests
}

func TestBlockRelationsData_Equal(t *testing.T) {

	blockRelationss := initTestBlockRelationsForEqual()
	for i, test := range blockRelationss {
		for j, subTest := range test.blockRelationsToCompareTo {
			result1 := test.baseBlockRelations.Equal(subTest.blockRelations)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.blockRelations.Equal(test.baseBlockRelations)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestBlockRelations_Clone(t *testing.T) {

	testBlockRelations := initTestBlockRelationsForClone()
	for i, blockRelations := range testBlockRelations {
		blockRelationsClone := blockRelations.Clone()
		if !blockRelationsClone.Equal(blockRelations) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(blockRelations, blockRelationsClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
