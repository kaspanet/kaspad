package blockdag

import (
	"reflect"
	"testing"
)

func TestSplitFraction(t *testing.T) {
	tests := []struct {
		interval      *reachabilityInterval
		fraction      float64
		expectedLeft  *reachabilityInterval
		expectedRight *reachabilityInterval
	}{
		{
			interval:      &reachabilityInterval{start: 1, end: 100},
			fraction:      0.5,
			expectedLeft:  &reachabilityInterval{start: 1, end: 50},
			expectedRight: &reachabilityInterval{start: 51, end: 100},
		},
		{
			interval:      &reachabilityInterval{start: 2, end: 100},
			fraction:      0.5,
			expectedLeft:  &reachabilityInterval{start: 2, end: 51},
			expectedRight: &reachabilityInterval{start: 52, end: 100},
		},
		{
			interval:      &reachabilityInterval{start: 1, end: 99},
			fraction:      0.5,
			expectedLeft:  &reachabilityInterval{start: 1, end: 50},
			expectedRight: &reachabilityInterval{start: 51, end: 99},
		},
		{
			interval:      &reachabilityInterval{start: 1, end: 100},
			fraction:      0.2,
			expectedLeft:  &reachabilityInterval{start: 1, end: 20},
			expectedRight: &reachabilityInterval{start: 21, end: 100},
		},
		{
			interval:      &reachabilityInterval{start: 1, end: 100},
			fraction:      0,
			expectedLeft:  &reachabilityInterval{start: 1, end: 0},
			expectedRight: &reachabilityInterval{start: 1, end: 100},
		},
		{
			interval:      &reachabilityInterval{start: 1, end: 100},
			fraction:      1,
			expectedLeft:  &reachabilityInterval{start: 1, end: 100},
			expectedRight: &reachabilityInterval{start: 101, end: 100},
		},
	}

	for i, test := range tests {
		left, right, err := test.interval.splitFraction(test.fraction)
		if err != nil {
			t.Fatalf("TestSplitFraction: splitFraction unexpectedly failed: %s", err)
		}
		if !reflect.DeepEqual(left, test.expectedLeft) {
			t.Errorf("TestSplitFraction: unexpected left in test #%d. "+
				"want: %s, got: %s", i, test.expectedLeft, left)
		}
		if !reflect.DeepEqual(right, test.expectedRight) {
			t.Errorf("TestSplitFraction: unexpected right in test #%d. "+
				"want: %s, got: %s", i, test.expectedRight, right)
		}
	}
}

func TestIsFutureBlock(t *testing.T) {
	blocks := futureBlocks{
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 2, end: 3}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 77}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
	}

	tests := []struct {
		block          *blockNode
		expectedResult bool
	}{
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 1}}},
			expectedResult: false,
		},
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 5, end: 7}}},
			expectedResult: true,
		},
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 76}}},
			expectedResult: true,
		},
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 78, end: 100}}},
			expectedResult: false,
		},
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1980, end: 2000}}},
			expectedResult: false,
		},
		{
			block:          &blockNode{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1920}}},
			expectedResult: true,
		},
	}

	for i, test := range tests {
		result := blocks.isFutureBlock(test.block)
		if result != test.expectedResult {
			t.Errorf("TestIsFutureBlock: unexpected result in test #%d. Want: %t, got: %t",
				i, test.expectedResult, result)
		}
	}
}

func TestInsertFutureBlock(t *testing.T) {
	blocks := futureBlocks{
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 3}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 77}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
		{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
	}

	tests := []struct {
		toInsert       []*blockNode
		expectedResult futureBlocks
	}{
		{
			toInsert: []*blockNode{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 5, end: 7}}},
			},
			expectedResult: futureBlocks{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 3}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 77}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
			},
		},
		{
			toInsert: []*blockNode{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 65, end: 78}}},
			},
			expectedResult: futureBlocks{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 3}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 65, end: 78}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
			},
		},
		{
			toInsert: []*blockNode{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 88, end: 97}}},
			},
			expectedResult: futureBlocks{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 3}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 77}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 88, end: 97}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
			},
		},
		{
			toInsert: []*blockNode{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 88, end: 97}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 3000, end: 3010}}},
			},
			expectedResult: futureBlocks{
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1, end: 3}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 4, end: 67}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 67, end: 77}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 88, end: 97}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 657, end: 789}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1000, end: 1000}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 1920, end: 1921}}},
				{reachabilityTreeNode: reachabilityTreeNode{interval: reachabilityInterval{start: 3000, end: 3010}}},
			},
		},
	}

	for i, test := range tests {
		// Create a clone of blocks so that we have a clean start for every test
		blocksClone := make(futureBlocks, len(blocks))
		for i, block := range blocks {
			blocksClone[i] = block
		}

		for _, block := range test.toInsert {
			blocksClone.insertFutureBlock(block)
		}
		if !reflect.DeepEqual(blocksClone, test.expectedResult) {
			t.Errorf("TestInsertFutureBlock: unexpected result in test #%d. Want: %s, got: %s",
				i, test.expectedResult, blocksClone)
		}
	}
}
