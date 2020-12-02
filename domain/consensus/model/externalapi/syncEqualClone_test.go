package externalapi

import (
	"reflect"
	"testing"
)

func InitTestSyncStateForClone() []SyncState {

	tests := []SyncState{'a', 'b', 0xFF, 0}

	return tests
}

type TestSyncStateToCompare struct {
	syncStatus     SyncState
	expectedResult bool
}

type TestSyncStatusStruct struct {
	baseSyncState            SyncState
	blockStatusesToCompareTo []TestSyncStateToCompare
}

func InitTestSyncStateForEqual() []TestSyncStatusStruct {
	tests := []TestSyncStatusStruct{
		{
			baseSyncState: 0,
			blockStatusesToCompareTo: []TestSyncStateToCompare{
				{
					syncStatus:     'a',
					expectedResult: false,
				},
				{
					syncStatus:     0,
					expectedResult: true,
				},
			},
		},
		{
			baseSyncState: 255,
			blockStatusesToCompareTo: []TestSyncStateToCompare{
				{
					syncStatus:     'a',
					expectedResult: false,
				},
				{
					syncStatus:     255,
					expectedResult: true,
				},
			},
		},
	}
	return tests
}

func TestSyncState_Equal(t *testing.T) {

	testSyncState := InitTestSyncStateForEqual()

	for i, test := range testSyncState {
		for j, subTest := range test.blockStatusesToCompareTo {
			result1 := test.baseSyncState.Equal(subTest.syncStatus)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}

			result2 := subTest.syncStatus.Equal(test.baseSyncState)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestSyncState_Clone(t *testing.T) {

	testSyncState := InitTestSyncStateForClone()

	for i, syncState := range testSyncState {
		clone := syncState.Clone()
		if !clone.Equal(syncState) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(syncState, clone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
