package model

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"reflect"
	"testing"
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
					&ReachabilityInterval{},
				},
				FutureCoveringTreeNodeSet{},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{},
							&ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: true,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}}, // Changed
							&externalapi.DomainHash{},
							&ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{},
							&externalapi.DomainHash{1}, // Changed
							&ReachabilityInterval{},
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
							&ReachabilityInterval{100, 0}, // Changed start
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
							&ReachabilityInterval{0, 100}, // Changed end
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
							&ReachabilityInterval{},
						},
						FutureCoveringTreeNodeSet{{1}, {2}}, // Changed
					},
					expectedResult: false,
				},
			},
		},
		// Test filled data
		{
			baseData: &ReachabilityData{
				&ReachabilityTreeNode{
					[]*externalapi.DomainHash{{1}, {2}, {3}},
					&externalapi.DomainHash{1},
					&ReachabilityInterval{100, 200},
				},
				FutureCoveringTreeNodeSet{{1}, {2}},
			},
			dataToCompareTo: []dataToCompare{
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}, {3}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: true,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}, {3}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{}, // Changed
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}, {3}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{200, 200}, // Changed start
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}, {3}},
							&externalapi.DomainHash{1},
							nil, //Changed
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}, {3}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{100, 100}, // Changed end
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{}, // Changed
							&externalapi.DomainHash{1},
							&ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}},
							&externalapi.DomainHash{}, // Changed
							&ReachabilityInterval{100, 200},
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{}, // Changed
						},
						FutureCoveringTreeNodeSet{{1}, {2}},
					},
					expectedResult: false,
				},
				{
					data: &ReachabilityData{
						&ReachabilityTreeNode{
							[]*externalapi.DomainHash{{1}, {2}},
							&externalapi.DomainHash{1},
							&ReachabilityInterval{100, 200},
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
				&ReachabilityInterval{},
			},
			FutureCoveringTreeNodeSet{},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{{1}, {2}},
				&externalapi.DomainHash{1},
				&ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{{1}, {2}},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{},
				&externalapi.DomainHash{1},
				&ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{{1}, {2}},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{{1}, {2}},
				&externalapi.DomainHash{1},
				&ReachabilityInterval{},
			},
			FutureCoveringTreeNodeSet{{1}, {2}},
		},
		{
			&ReachabilityTreeNode{
				[]*externalapi.DomainHash{},
				&externalapi.DomainHash{1},
				&ReachabilityInterval{100, 200},
			},
			FutureCoveringTreeNodeSet{{1}, {2}},
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
