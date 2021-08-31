package blockheader

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
	"testing"
)

func TestDomainBlockHeader_Equal(t *testing.T) {
	type headerToCompare struct {
		header         *blockHeader
		expectedResult bool
	}
	tests := []struct {
		baseHeader         *blockHeader
		headersToCompareTo []headerToCompare
	}{
		{
			baseHeader: nil,
			headersToCompareTo: []headerToCompare{
				{
					header:         nil,
					expectedResult: true,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{0})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						4,
						5,
						6,
						7,
						8,
						big.NewInt(9),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{10}),
					},
					expectedResult: false,
				},
			},
		},
		{
			baseHeader: &blockHeader{
				0,
				[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
				5,
				6,
				7,
				8,
				9,
				big.NewInt(10),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
			},
			headersToCompareTo: []headerToCompare{
				{
					header:         nil,
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: true,
				},
				{
					header: &blockHeader{
						100,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						100,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						100,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						100,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						100,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						100,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(100),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{11}),
					},
					expectedResult: false,
				},
				{
					header: &blockHeader{
						0,
						[]externalapi.BlockLevelParents{[]*externalapi.DomainHash{externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1})}},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{4}),
						5,
						6,
						7,
						8,
						9,
						big.NewInt(10),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{100}),
					},
					expectedResult: false,
				},
			},
		},
	}

	for i, test := range tests {
		for j, subTest := range test.headersToCompareTo {
			result1 := test.baseHeader.Equal(subTest.header)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.header.Equal(test.baseHeader)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}
