package externalapi

import (
	"reflect"
	"testing"
)

func InitTestBlockStatusForClone() []BlockStatus {

	tests := []BlockStatus{'a', 'b', 0xFF, 0}

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

func InitTestBlockStatusForEqual() []TestBlockStatusStruct {
	tests := []TestBlockStatusStruct{
		{
			baseBlockStatus: 0,
			blockStatusesToCompareTo: []TestBlockStatusToCompare{
				{
					blockStatus:    'a',
					expectedResult: false,
				},
				{
					blockStatus:    0,
					expectedResult: true,
				},
			},
		},
		{
			baseBlockStatus: 255,
			blockStatusesToCompareTo: []TestBlockStatusToCompare{
				{
					blockStatus:    'a',
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

	testBlockStatus := InitTestBlockStatusForEqual()

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

	testBlockStatus := InitTestBlockStatusForClone()

	for i, blockStatus := range testBlockStatus {
		clone := blockStatus.Clone()
		if !clone.Equal(blockStatus) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(blockStatus, clone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
