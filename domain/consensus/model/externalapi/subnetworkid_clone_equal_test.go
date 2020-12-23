package externalapi

import (
	"reflect"
	"testing"
)

func initTestDomainSubnetworkIDForClone() []*DomainSubnetworkID {

	tests := []*DomainSubnetworkID{{1, 0, 0xFF, 0}, {0, 1, 0xFF, 1},
		{0, 1, 0xFF, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}
	return tests
}

type testDomainSubnetworkIDToCompare struct {
	domainSubnetworkID *DomainSubnetworkID
	expectedResult     bool
}

type testDomainSubnetworkIDStruct struct {
	baseDomainSubnetworkID        *DomainSubnetworkID
	domainSubnetworkIDToCompareTo []testDomainSubnetworkIDToCompare
}

func initTestDomainSubnetworkIDForEqual() []testDomainSubnetworkIDStruct {
	tests := []testDomainSubnetworkIDStruct{
		{
			baseDomainSubnetworkID: nil,
			domainSubnetworkIDToCompareTo: []testDomainSubnetworkIDToCompare{
				{
					domainSubnetworkID: &DomainSubnetworkID{255, 255, 0xFF, 0},
					expectedResult:     false,
				},
				{
					domainSubnetworkID: nil,
					expectedResult:     true,
				},
			},
		}, {
			baseDomainSubnetworkID: &DomainSubnetworkID{0},
			domainSubnetworkIDToCompareTo: []testDomainSubnetworkIDToCompare{
				{
					domainSubnetworkID: &DomainSubnetworkID{255, 254, 0xFF, 0},
					expectedResult:     false,
				},
				{
					domainSubnetworkID: &DomainSubnetworkID{0},
					expectedResult:     true,
				},
			},
		}, {
			baseDomainSubnetworkID: &DomainSubnetworkID{0, 1, 0xFF, 1, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			domainSubnetworkIDToCompareTo: []testDomainSubnetworkIDToCompare{
				{
					domainSubnetworkID: &DomainSubnetworkID{0, 1, 0xFF, 1, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
						0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
					expectedResult: true,
				},
				{
					domainSubnetworkID: &DomainSubnetworkID{0, 10, 0xFF, 0},
					expectedResult:     false,
				},
			},
		},
	}
	return tests
}

func TestDomainSubnetworkID_Equal(t *testing.T) {

	domainSubnetworkIDs := initTestDomainSubnetworkIDForEqual()
	for i, test := range domainSubnetworkIDs {
		for j, subTest := range test.domainSubnetworkIDToCompareTo {
			result1 := test.baseDomainSubnetworkID.Equal(subTest.domainSubnetworkID)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.domainSubnetworkID.Equal(test.baseDomainSubnetworkID)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestDomainSubnetworkID_Clone(t *testing.T) {

	domainSubnetworkIDs := initTestDomainSubnetworkIDForClone()
	for i, domainSubnetworkID := range domainSubnetworkIDs {
		domainSubnetworkIDClone := domainSubnetworkID.Clone()
		if !domainSubnetworkIDClone.Equal(domainSubnetworkID) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(domainSubnetworkID, domainSubnetworkIDClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
