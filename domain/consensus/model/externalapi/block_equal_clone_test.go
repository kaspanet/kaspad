package externalapi

import (
	"reflect"
	"testing"
)

type blockToCompare struct {
	block          *DomainBlock
	expectedResult bool
}

type TestBlockStruct struct {
	baseBlock         *DomainBlock
	blocksToCompareTo []blockToCompare
}

func initTestBaseTransactions() []*DomainTransaction {

	testTx := []*DomainTransaction{{
		Version:      1,
		Inputs:       []*DomainTransactionInput{},
		Outputs:      []*DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: DomainSubnetworkID{0x01},
		Gas:          1,
		PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Payload: []byte{0x01},
		Fee:     0,
		Mass:    1,
		ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
	}}
	return testTx
}

func initTestAnotherTransactions() []*DomainTransaction {

	testTx := []*DomainTransaction{{
		Version:      1,
		Inputs:       []*DomainTransactionInput{},
		Outputs:      []*DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: DomainSubnetworkID{0x01},
		Gas:          1,
		PayloadHash: DomainHash{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		Payload: []byte{0x01},
		Fee:     0,
		Mass:    1,
		ID: &DomainTransactionID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
	}}
	return testTx
}

func initTestBlockStructsForClone() []*DomainBlock {

	tests := []*DomainBlock{
		{
			&DomainBlockHeader{

				0,
				[]*DomainHash{{0}},
				DomainHash{1},
				DomainHash{2},
				DomainHash{3},
				4,
				5,
				6,
			},
			initTestBaseTransactions(),
		}, {
			&DomainBlockHeader{

				0,
				[]*DomainHash{},
				DomainHash{1},
				DomainHash{2},
				DomainHash{3},
				4,
				5,
				6,
			},
			initTestBaseTransactions(),
		},
	}

	return tests
}

func initTestBlockStructsForEqual() *[]TestBlockStruct {

	tests := []TestBlockStruct{
		{
			baseBlock: nil,
			blocksToCompareTo: []blockToCompare{
				{
					block:          nil,
					expectedResult: true,
				},
				{
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{0}},
							DomainHash{1},
							DomainHash{2},
							DomainHash{3},
							4,
							5,
							6,
						},
						initTestBaseTransactions()},
					expectedResult: false,
				},
			},
		}, {
			baseBlock: &DomainBlock{
				&DomainBlockHeader{
					0,
					[]*DomainHash{{1}},
					DomainHash{2},
					DomainHash{3},
					DomainHash{4},
					5,
					6,
					7,
				},
				initTestBaseTransactions(),
			},
			blocksToCompareTo: []blockToCompare{
				{
					block:          nil,
					expectedResult: false,
				},
				{
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestAnotherTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: true,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}, {2}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{100}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{100},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{100},
							DomainHash{4},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{100},
							5,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							100,
							6,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							100,
							7,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &DomainBlock{
						&DomainBlockHeader{
							0,
							[]*DomainHash{{1}},
							DomainHash{2},
							DomainHash{3},
							DomainHash{4},
							5,
							6,
							100,
						},
						initTestBaseTransactions(),
					},
					expectedResult: false,
				},
			},
		},
	}

	return &tests
}

func TestDomainBlock_Equal(t *testing.T) {

	blockTests := initTestBlockStructsForEqual()
	for i, test := range *blockTests {
		for j, subTest := range test.blocksToCompareTo {
			result1 := test.baseBlock.Equal(subTest.block)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.block.Equal(test.baseBlock)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}

}

func TestDomainBlock_Clone(t *testing.T) {

	blocks := initTestBlockStructsForClone()
	for i, block := range blocks {
		blockClone := block.Clone()
		if !blockClone.Equal(block) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(block, blockClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
