package externalapi

import (
	"reflect"
	"testing"
)

func TestDomainBlockHeader_Equal(t *testing.T) {
	type headerToCompare struct {
		header         *DomainBlockHeader
		expectedResult bool
	}
	tests := []struct {
		baseHeader         *DomainBlockHeader
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
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{0}},
						DomainHash{1},
						DomainHash{2},
						DomainHash{3},
						4,
						5,
						6,
					},
					expectedResult: false,
				},
			},
		},
		{
			baseHeader: &DomainBlockHeader{
				0,
				[]*DomainHash{{1}},
				DomainHash{2},
				DomainHash{3},
				DomainHash{4},
				5,
				6,
				7,
			},
			headersToCompareTo: []headerToCompare{
				{
					header:         nil,
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: true,
				},
				{
					header: &DomainBlockHeader{
						100,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}, {2}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{100}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{100},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{100},
						DomainHash{4},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{100},
						5,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						100,
						6,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						100,
						7,
					},
					expectedResult: false,
				},
				{
					header: &DomainBlockHeader{
						0,
						[]*DomainHash{{1}},
						DomainHash{2},
						DomainHash{3},
						DomainHash{4},
						5,
						6,
						100,
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

func TestDomainBlockHeader_Clone(t *testing.T) {
	headers := []*DomainBlockHeader{
		{
			0,
			[]*DomainHash{{0}},
			DomainHash{1},
			DomainHash{2},
			DomainHash{3},
			4,
			5,
			6,
		},
		{
			0,
			[]*DomainHash{},
			DomainHash{1},
			DomainHash{2},
			DomainHash{3},
			4,
			5,
			6,
		},
	}

	for i, header := range headers {
		clone := header.Clone()
		if !clone.Equal(header) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}

		if !reflect.DeepEqual(header, clone) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}
	}
}
