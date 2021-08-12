package externalapi_test

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/blockheader"
	"math/big"
	"reflect"
	"testing"
)

type blockToCompare struct {
	block          *externalapi.DomainBlock
	expectedResult bool
}

type TestBlockStruct struct {
	baseBlock         *externalapi.DomainBlock
	blocksToCompareTo []blockToCompare
}

func initTestBaseTransactions() []*externalapi.DomainTransaction {

	testTx := []*externalapi.DomainTransaction{{
		Version:      1,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: externalapi.DomainSubnetworkID{0x01},
		Gas:          1,
		Payload:      []byte{0x01},
		Fee:          0,
		Mass:         1,
		ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}),
	}}
	return testTx
}

func initTestAnotherTransactions() []*externalapi.DomainTransaction {

	testTx := []*externalapi.DomainTransaction{{
		Version:      1,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: externalapi.DomainSubnetworkID{0x01},
		Gas:          1,
		Payload:      []byte{0x02},
		Fee:          0,
		Mass:         1,
		ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}),
	}}
	return testTx
}

func initTestTwoTransactions() []*externalapi.DomainTransaction {

	testTx := []*externalapi.DomainTransaction{{
		Version:      1,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: externalapi.DomainSubnetworkID{0x01},
		Gas:          1,
		Payload:      []byte{0x01},
		Fee:          0,
		Mass:         1,
		ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}),
	}, {
		Version:      1,
		Inputs:       []*externalapi.DomainTransactionInput{},
		Outputs:      []*externalapi.DomainTransactionOutput{},
		LockTime:     1,
		SubnetworkID: externalapi.DomainSubnetworkID{0x01},
		Gas:          1,
		Payload:      []byte{0x01},
		Fee:          0,
		Mass:         1,
		ID: externalapi.NewDomainTransactionIDFromByteArray(&[externalapi.DomainHashSize]byte{
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}),
	}}
	return testTx
}

func initTestBlockStructsForClone() []*externalapi.DomainBlock {

	tests := []*externalapi.DomainBlock{
		{
			blockheader.NewImmutableBlockHeader(

				0,
				[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0})},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
				4,
				5,
				6,
				7,
				big.NewInt(8),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{9}),
			),
			initTestBaseTransactions(),
		}, {
			blockheader.NewImmutableBlockHeader(
				0,
				[]*externalapi.DomainHash{},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
				4,
				5,
				6,
				7,
				big.NewInt(8),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{9}),
			),
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
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							4,
							5,
							6,
							7,
							big.NewInt(8),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{9}),
						),
						initTestBaseTransactions()},
					expectedResult: false,
				},
			},
		}, {
			baseBlock: &externalapi.DomainBlock{
				blockheader.NewImmutableBlockHeader(
					0,
					[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
					5,
					6,
					7,
					8,
					big.NewInt(9),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
				),
				initTestBaseTransactions(),
			},
			blocksToCompareTo: []blockToCompare{
				{
					block:          nil,
					expectedResult: false,
				},
				{
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestAnotherTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: true,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100})}, // Changed
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestTwoTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}), // Changed
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}), // Changed
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}), // Changed
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							100, // Changed
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							100, // Changed
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							100, // Changed
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							100, // Changed
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(100), // Changed
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
						),
						initTestBaseTransactions(),
					},
					expectedResult: false,
				}, {
					block: &externalapi.DomainBlock{
						blockheader.NewImmutableBlockHeader(
							0,
							[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
							5,
							6,
							7,
							8,
							big.NewInt(9),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}), // Changed
						),
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
