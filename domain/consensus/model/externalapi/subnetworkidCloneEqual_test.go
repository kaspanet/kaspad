package externalapi

import (
	"reflect"
	"testing"
)

func InitTestDomainSubnetworkIDForClone() []DomainSubnetworkID {

	tests := []DomainSubnetworkID{{'a', 'b', 0xFF, 0}, {0, 1, 0xFF, 1},
		{0, 1, 0xFF, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}}

	return tests
}

type TestDomainSubnetworkIDToCompare struct {
	domainSubnetworkID DomainSubnetworkID
	expectedResult     bool
}

type TestDomainSubnetworkIDStruct struct {
	baseDomainSubnetworkID        DomainSubnetworkID
	domainSubnetworkIDToCompareTo []TestDomainSubnetworkIDToCompare
}

func InitTestDomainSubnetworkIDForEqual() []TestDomainSubnetworkIDStruct {
	tests := []TestDomainSubnetworkIDStruct{
		{
			baseDomainSubnetworkID: DomainSubnetworkID{0},
			domainSubnetworkIDToCompareTo: []TestDomainSubnetworkIDToCompare{
				{
					domainSubnetworkID: DomainSubnetworkID{'a', 'b', 0xFF, 0},
					expectedResult:     false,
				},
				{
					domainSubnetworkID: DomainSubnetworkID{0},
					expectedResult:     true,
				},
			},
		},
		{
			baseDomainSubnetworkID: DomainSubnetworkID{0, 1, 0xFF, 1, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
				0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			domainSubnetworkIDToCompareTo: []TestDomainSubnetworkIDToCompare{
				{
					domainSubnetworkID: DomainSubnetworkID{0, 1, 0xFF, 1, 1, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
						0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
					expectedResult: true,
				},
				{
					domainSubnetworkID: DomainSubnetworkID{'a', 'b', 0xFF, 0},
					expectedResult:     false,
				},
			},
		},
	}
	return tests
}

func TestDomainSubnetworkID_Equal(t *testing.T) {

	domainSubnetworkIDs := InitTestDomainSubnetworkIDForEqual()

	for i, test := range domainSubnetworkIDs {
		for j, subTest := range test.domainSubnetworkIDToCompareTo {
			result1 := test.baseDomainSubnetworkID.Equal(&subTest.domainSubnetworkID)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.domainSubnetworkID.Equal(&test.baseDomainSubnetworkID)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestDomainSubnetworkID_Clone(t *testing.T) {

	domainSubnetworkIDs := InitTestDomainSubnetworkIDForClone()

	for i, domainSubnetworkID := range domainSubnetworkIDs {
		clone := domainSubnetworkID.Clone()
		if !clone.Equal(&domainSubnetworkID) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(domainSubnetworkID, clone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
