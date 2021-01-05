package reachabilitydata

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func TestReachabilityData_Equal(t *testing.T) {
	type dataToCompare struct {
		data           *reachabilityData
		expectedResult bool
	}
	tests := []struct {
		baseData        *reachabilityData
		dataToCompareTo []dataToCompare
	}{
		// Test nil data
		{
			baseData: nil,
			dataToCompareTo: []dataToCompare{
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
			},
		},
		// Test empty data
		{
			baseData: &reachabilityData{
				[]*externalapi.DomainHash{},
				&externalapi.DomainHash{},
				model.FutureCoveringTreeNodeSet{},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: true,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						}, // Changed
						&externalapi.DomainHash{},
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}), // Changed
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						model.FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})}, // Changed
					},
					expectedResult: false,
				},
			},
		},
		// Test filled data
		{
			baseData: &reachabilityData{
				[]*externalapi.DomainHash{
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				model.FutureCoveringTreeNodeSet{
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						model.FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: true,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						model.FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{}, // Changed
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						model.FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
						&externalapi.DomainHash{}, // Changed
						model.FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						model.FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						nil,
						nil,
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data:           nil,
					expectedResult: false,
				},
			},
		},
	}

	for i, test := range tests {
		for j, subTest := range test.dataToCompareTo {
			result1 := test.baseData.Equal(subTest.data)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.data.Equal(test.baseData)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestReachabilityData_CloneWritable(t *testing.T) {
	testData := []*reachabilityData{
		{
			[]*externalapi.DomainHash{},
			&externalapi.DomainHash{},
			model.FutureCoveringTreeNodeSet{},
		},
		{
			[]*externalapi.DomainHash{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
	}

	for i, data := range testData {
		clone := data.CloneMutable()
		if !clone.Equal(data) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}

		if !reflect.DeepEqual(data, clone) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}
	}
}
