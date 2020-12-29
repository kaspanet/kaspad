package model

import (
	"reflect"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func TestReachabilityData_Equal(t *testing.T) {
	type dataToCompare struct {
		data           *ReachabilityData
		expectedResult bool
	}
	tests := []struct {
		baseData        *ReachabilityData
		dataToCompareTo []dataToCompare
	}{
		// Test nil data
		{
			baseData:        nil,
			dataToCompareTo: nil,
		},
		// Test empty data
		{
			baseData: &ReachabilityData{
				&ReachabilityTreeNode{
					[]*externalapi.DomainHash{},
					&externalapi.DomainHash{},
					&reachabilitydata.ReachabilityInterval{},
				},
				FutureCoveringTreeNodeSet{},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{},
							&reachabilitydata.ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: true,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
							}, // Changed
							&externalapi.DomainHash{},
							&reachabilitydata.ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}), // Changed
							&reachabilitydata.ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{},
							&reachabilitydata.ReachabilityInterval{100, 0}, // Changed start
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{},
							&reachabilitydata.ReachabilityInterval{0, 100}, // Changed end
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{},
							&reachabilitydata.ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})}, // Changed
					},
					expectedResult: false,
				},
			},
		},
		// Test filled data
		{
			baseData: &ReachabilityData{
				&ReachabilityTreeNode{
					[]*externalapi.DomainHash{
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
						externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					&reachabilitydata.ReachabilityInterval{100, 200},
				},
				FutureCoveringTreeNodeSet{
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: true,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{200, 200}, // Changed start
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							nil, //Changed
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{3})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{100, 100}, // Changed end
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{}, // Changed
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
							&externalapi.DomainHash{}, // Changed
							&reachabilitydata.ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{}, // Changed
						},
						FutureCoveringTreeNodeSet{
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
								externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
							externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
							&reachabilitydata.ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						nil,
						FutureCoveringTreeNodeSet{},
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

func TestReachabilityData_Clone(t *testing.T) {
	testData := []*ReachabilityData{
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{},
				&externalapi.DomainHash{},
				&reachabilitydata.ReachabilityInterval{},
			},
			FutureCoveringTreeNodeSet{},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				&reachabilitydata.ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				&reachabilitydata.ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
					externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				&reachabilitydata.ReachabilityInterval{},
			},
			FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{},
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				&reachabilitydata.ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{1}),
				externalapi.NewDomainHashFromByteArray(&[externalapi.DomainHashSize]byte{2})},
		},
	}

	for i, data := range testData {
		clone := data.Clone()
		if !clone.Equal(data) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}

		if !reflect.DeepEqual(data, clone) {
			t.Fatalf("Test #%d: clone should be equal to the original", i)
		}
	}
}
