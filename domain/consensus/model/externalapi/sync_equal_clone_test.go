package externalapi

import (
	"reflect"
	"testing"
)

func initTestSyncInfoForClone() []*SyncInfo {

	tests := []*SyncInfo{{
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
						0xF,
						0xF},
					expectedResult: false,
				}, {
					syncInfo:       nil,
					expectedResult: true,
				},
			}}, {
			baseSyncInfo: &SyncInfo{
				0xF,
				0xF},
			syncInfoToCompareTo: []testSyncInfoToCompare{
				{
					syncInfo: &SyncInfo{
						0xF,
						0xF},
					expectedResult: true,
				},
				{
					syncInfo: &SyncInfo{
						0xF1,
						0xF},
					expectedResult: false,
				}, {
					syncInfo:       nil,
					expectedResult: false,
				}, {
					syncInfo: &SyncInfo{
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
