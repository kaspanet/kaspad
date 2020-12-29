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
			baseData:        nil,
			dataToCompareTo: nil,
		},
		// Test empty data
		{
			baseData: &reachabilityData{
				[]*externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&model.ReachabilityInterval{},
				model.FutureCoveringTreeNodeSet{},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						&model.ReachabilityInterval{},
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
						&model.ReachabilityInterval{},
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}), // Changed
						&model.ReachabilityInterval{},
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						&model.ReachabilityInterval{100, 0}, // Changed start
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						&model.ReachabilityInterval{0, 100}, // Changed end
						model.FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{},
						&externalapi.DomainHash{},
						&model.ReachabilityInterval{},
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
				&model.ReachabilityInterval{100, 200},
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
						&model.ReachabilityInterval{100, 200},
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
						&model.ReachabilityInterval{100, 200},
						model.FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						&model.ReachabilityInterval{200, 200}, // Changed start
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
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						nil, //Changed
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
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						&model.ReachabilityInterval{100, 100}, // Changed end
						model.FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						[]*externalapi.DomainHash{}, // Changed
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						&model.ReachabilityInterval{100, 200},
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
						&model.ReachabilityInterval{100, 200},
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
						&model.ReachabilityInterval{}, // Changed
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
						&model.ReachabilityInterval{100, 200},
						model.FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &reachabilityData{
						nil,
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
			&model.ReachabilityInterval{},
			model.FutureCoveringTreeNodeSet{},
		},
		{
			[]*externalapi.DomainHash{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			&model.ReachabilityInterval{100, 200},
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			&model.ReachabilityInterval{100, 200},
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			&model.ReachabilityInterval{},
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			[]*externalapi.DomainHash{},
			externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
			&model.ReachabilityInterval{100, 200},
			model.FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
	}

	for i, data := range testData {
		clone := data.CloneWritable()
		if !clone.Equal(data) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}

		if !reflect.DeepEqual(data, clone) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}
	}
}
