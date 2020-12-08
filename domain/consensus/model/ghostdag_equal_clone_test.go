package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"
)

func initTestBlockGHOSTDAGDataForClone() []*BlockGHOSTDAGData {

	var testMap = map[externalapi.DomainHash]KType{}
	testMap[externalapi.DomainHash{1}] = 0x01
	tests := []*BlockGHOSTDAGData{
		{
			1,
			&externalapi.DomainHash{1},
			[]*externalapi.DomainHash{{1}, {2}},
			[]*externalapi.DomainHash{{1}, {2}},
			testMap,
		},
	}
	return tests
}

type TestBlockGHOSTDAGDataToCompare struct {
	blockGHOSTDAGData *BlockGHOSTDAGData
	expectedResult    bool
}

type TestBlockGHOSTDAGDataStruct struct {
	baseBlockGHOSTDAGData        *BlockGHOSTDAGData
	blockGHOSTDAGDataToCompareTo []TestBlockGHOSTDAGDataToCompare
}

func initTestBlockGHOSTDAGDataForEqual() []TestBlockGHOSTDAGDataStruct {

	var testMapBase = map[externalapi.DomainHash]KType{}
	testMapBase[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGDataBase = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		testMapBase,
	}
	var test1Map = map[externalapi.DomainHash]KType{}
	test1Map[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGData1 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test1Map,
	}
	var test2Map = map[externalapi.DomainHash]KType{}
	test2Map[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGData2 = BlockGHOSTDAGData{
		2, // changed
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test2Map,
	}
	var test3Map = map[externalapi.DomainHash]KType{}
	test3Map[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGData3 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{2}, // changed
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test3Map,
	}

	var test4Map = map[externalapi.DomainHash]KType{}
	test4Map[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGData4 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {3}}, //the first (count from zero) MergeSetBlues changed
		[]*externalapi.DomainHash{{1}, {2}},
		test4Map,
	}

	var test5Map = map[externalapi.DomainHash]KType{}
	test5Map[externalapi.DomainHash{1}] = 0x01
	var testBlockGHOSTDAGData5 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {3}}, //the first (count from zero) MergeSetReds changed
		test5Map,
	}

	var test6Map = map[externalapi.DomainHash]KType{} //map_size == 2
	test6Map[externalapi.DomainHash{1}] = 0x01
	test6Map[externalapi.DomainHash{2}] = 0x02
	var testBlockGHOSTDAGData6 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test6Map,
	}

	var test7Map = map[externalapi.DomainHash]KType{}
	test7Map[externalapi.DomainHash{2}] = 0x01 //DomainHash
	var testBlockGHOSTDAGData7 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test7Map,
	}

	var test8Map = map[externalapi.DomainHash]KType{}
	test8Map[externalapi.DomainHash{1}] = 0x02 //KType
	var testBlockGHOSTDAGData8 = BlockGHOSTDAGData{
		1,
		&externalapi.DomainHash{1},
		[]*externalapi.DomainHash{{1}, {2}},
		[]*externalapi.DomainHash{{1}, {2}},
		test8Map,
	}

	tests := []TestBlockGHOSTDAGDataStruct{
		{
			baseBlockGHOSTDAGData: &testBlockGHOSTDAGDataBase,
			blockGHOSTDAGDataToCompareTo: []TestBlockGHOSTDAGDataToCompare{
				{
					blockGHOSTDAGData: &testBlockGHOSTDAGData1,
					expectedResult:    true,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData2,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData3,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData4,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData5,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData6,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData7,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: &testBlockGHOSTDAGData8,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: nil,
					expectedResult:    false,
				},
			},
		}, {
			baseBlockGHOSTDAGData: nil,
			blockGHOSTDAGDataToCompareTo: []TestBlockGHOSTDAGDataToCompare{
				{
					blockGHOSTDAGData: &testBlockGHOSTDAGData1,
					expectedResult:    false,
				}, {
					blockGHOSTDAGData: nil,
					expectedResult:    true,
				},
			},
		},
	}
	return tests
}

func TestBlockGHOSTDAGData_Equal(t *testing.T) {

	blockGHOSTDAGDatas := initTestBlockGHOSTDAGDataForEqual()
	for i, test := range blockGHOSTDAGDatas {
		for j, subTest := range test.blockGHOSTDAGDataToCompareTo {
			result1 := test.baseBlockGHOSTDAGData.Equal(subTest.blockGHOSTDAGData)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.blockGHOSTDAGData.Equal(test.baseBlockGHOSTDAGData)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestBlockGHOSTDAGData_Clone(t *testing.T) {

	testBlockGHOSTDAGData := initTestBlockGHOSTDAGDataForClone()
	for i, blockGHOSTDAGData := range testBlockGHOSTDAGData {
		blockGHOSTDAGDataClone := blockGHOSTDAGData.Clone()
		if !blockGHOSTDAGDataClone.Equal(blockGHOSTDAGData) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(blockGHOSTDAGData, blockGHOSTDAGDataClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
