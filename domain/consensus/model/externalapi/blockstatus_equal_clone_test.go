package externalapi

import (
	"reflect"
	"testing"
)

func initTestBlockStatusForClone() []BlockStatus {

	tests := []BlockStatus{1, 2, 0xFF, 0}

	return tests
}

type TestBlockStatusToCompare struct {
	blockStatus    BlockStatus
	expectedResult bool
}

type TestBlockStatusStruct struct {
	baseBlockStatus          BlockStatus
	blockStatusesToCompareTo []TestBlockStatusToCompare
}

func initTestBlockStatusForEqual() []TestBlockStatusStruct {
	tests := []TestBlockStatusStruct{
		{
			baseBlockStatus: 0,
			blockStatusesToCompareTo: []TestBlockStatusToCompare{
				{
					blockStatus:    1,
					expectedResult: false,
				},
				{
					blockStatus:    0,
					expectedResult: true,
				},
			},
		}, {
			baseBlockStatus: 255,
			blockStatusesToCompareTo: []TestBlockStatusToCompare{
				{
					blockStatus:    1,
					expectedResult: false,
				},
				{
					blockStatus:    255,
					expectedResult: true,
				},
			},
		},
	}
	return tests
}

func TestBlockStatus_Equal(t *testing.T) {

	testBlockStatus := initTestBlockStatusForEqual()

	for i, test := range testBlockStatus {
		for j, subTest := range test.blockStatusesToCompareTo {
			result1 := test.baseBlockStatus.Equal(subTest.blockStatus)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.blockStatus.Equal(test.baseBlockStatus)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestBlockStatus_Clone(t *testing.T) {

	testBlockStatus := initTestBlockStatusForClone()
	for i, blockStatus := range testBlockStatus {
		blockStatusClone := blockStatus.Clone()
		if !blockStatusClone.Equal(blockStatus) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(blockStatus, blockStatusClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
