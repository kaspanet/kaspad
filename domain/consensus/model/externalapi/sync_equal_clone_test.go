package externalapi

import (
	"reflect"
	"testing"
)

func initTestSyncStateForClone() []SyncState {

	tests := []SyncState{1, 2, 0xFF, 0}
	return tests
}

type TestSyncStateToCompare struct {
	syncStatus     SyncState
	expectedResult bool
}

type TestSyncStatusStruct struct {
	baseSyncState         SyncState
	syncStatesToCompareTo []TestSyncStateToCompare
}

func initTestSyncStateForEqual() []TestSyncStatusStruct {
	tests := []TestSyncStatusStruct{
		{
			baseSyncState: 0,
			syncStatesToCompareTo: []TestSyncStateToCompare{
				{
					syncStatus:     1,
					expectedResult: false,
				}, {
					syncStatus:     0,
					expectedResult: true,
				},
			},
		}, {
			baseSyncState: 255,
			syncStatesToCompareTo: []TestSyncStateToCompare{
				{
					syncStatus:     1,
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

	testSyncState := initTestSyncStateForEqual()
	for i, test := range testSyncState {
		for j, subTest := range test.syncStatesToCompareTo {
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

	testSyncState := initTestSyncStateForClone()
	for i, syncState := range testSyncState {
		syncStateClone := syncState.Clone()
		if !syncStateClone.Equal(syncState) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(syncState, syncStateClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}

func initTestSyncInfoForClone() []*SyncInfo {

	tests := []*SyncInfo{{
		SyncState(1),
		&DomainHash{1, 2},
		0xF,
		0xF}}
	return tests
}

type testSyncInfoToCompare struct {
	syncInfo       *SyncInfo
	expectedResult bool
}

type testSyncInfoStruct struct {
	baseSyncInfo        *SyncInfo
	syncInfoToCompareTo []testSyncInfoToCompare
}

func initTestSyncInfoForEqual() []*testSyncInfoStruct {
	tests := []*testSyncInfoStruct{
		{
			baseSyncInfo: nil,
			syncInfoToCompareTo: []testSyncInfoToCompare{
				{
					syncInfo: &SyncInfo{
						SyncState(1),
						&DomainHash{1, 2},
						0xF,
						0xF},
					expectedResult: false,
				}, {
					syncInfo:       nil,
					expectedResult: true,
				},
			}}, {
			baseSyncInfo: &SyncInfo{
				SyncState(1),
				&DomainHash{1, 2},
				0xF,
				0xF},
			syncInfoToCompareTo: []testSyncInfoToCompare{
				{
					syncInfo: &SyncInfo{
						SyncState(1),
						&DomainHash{1, 2},
						0xF,
						0xF},
					expectedResult: true,
				}, {
					syncInfo: &SyncInfo{
						SyncState(2),
						&DomainHash{1, 2},
						0xF,
						0xF},
					expectedResult: false,
				},
				{
					syncInfo: &SyncInfo{
						SyncState(1),
						&DomainHash{1, 3},
						0xF,
						0xF},
					expectedResult: false,
				},
				{
					syncInfo: &SyncInfo{
						SyncState(1),
						&DomainHash{1, 2},
						0xF1,
						0xF},
					expectedResult: false,
				}, {
					syncInfo:       nil,
					expectedResult: false,
				}, {
					syncInfo: &SyncInfo{
						SyncState(1),
						&DomainHash{1, 2},
						0xF,
						0xF1},
					expectedResult: false},
			},
		},
	}
	return tests
}

func TestSyncInfo_Equal(t *testing.T) {

	testSyncState := initTestSyncInfoForEqual()
	for i, test := range testSyncState {
		for j, subTest := range test.syncInfoToCompareTo {
			result1 := test.baseSyncInfo.Equal(subTest.syncInfo)
			if result1 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result1)
			}
			result2 := subTest.syncInfo.Equal(test.baseSyncInfo)
			if result2 != subTest.expectedResult {
				t.Fatalf("Test #%d:%d: Expected %t but got %t", i, j, subTest.expectedResult, result2)
			}
		}
	}
}

func TestSyncInfo_Clone(t *testing.T) {

	testSyncInfo := initTestSyncInfoForClone()
	for i, syncInfo := range testSyncInfo {
		syncStateClone := syncInfo.Clone()
		if !syncStateClone.Equal(syncInfo) {
			t.Fatalf("Test #%d:[Equal] clone should be equal to the original", i)
		}
		if !reflect.DeepEqual(syncInfo, syncStateClone) {
			t.Fatalf("Test #%d:[DeepEqual] clone should be equal to the original", i)
		}
	}
}
