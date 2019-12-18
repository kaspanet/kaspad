package blockdag

import (
	"reflect"
	"testing"
)

func TestIsFutureBlock(t *testing.T) {
	blocks := futureBlocks{
		{treeInterval: reachabilityInterval{start: 1, end: 3}},
		{treeInterval: reachabilityInterval{start: 4, end: 67}},
		{treeInterval: reachabilityInterval{start: 67, end: 77}},
		{treeInterval: reachabilityInterval{start: 657, end: 789}},
		{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
		{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
	}

	tests := []struct {
		block          *blockNode
		expectedResult bool
	}{
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 0, end: 0}},
			expectedResult: false,
		},
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 5, end: 7}},
			expectedResult: true,
		},
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 67, end: 76}},
			expectedResult: true,
		},
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 78, end: 100}},
			expectedResult: false,
		},
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 1980, end: 2000}},
			expectedResult: false,
		},
		{
			block:          &blockNode{treeInterval: reachabilityInterval{start: 1920, end: 1920}},
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
		{treeInterval: reachabilityInterval{start: 1, end: 3}},
		{treeInterval: reachabilityInterval{start: 4, end: 67}},
		{treeInterval: reachabilityInterval{start: 67, end: 77}},
		{treeInterval: reachabilityInterval{start: 657, end: 789}},
		{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
		{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
	}

	tests := []struct {
		toInsert       []*blockNode
		expectedResult futureBlocks
	}{
		{
			toInsert: []*blockNode{
				{treeInterval: reachabilityInterval{start: 5, end: 7}},
			},
			expectedResult: futureBlocks{
				{treeInterval: reachabilityInterval{start: 1, end: 3}},
				{treeInterval: reachabilityInterval{start: 4, end: 67}},
				{treeInterval: reachabilityInterval{start: 67, end: 77}},
				{treeInterval: reachabilityInterval{start: 657, end: 789}},
				{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
				{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
			},
		},
		{
			toInsert: []*blockNode{
				{treeInterval: reachabilityInterval{start: 65, end: 78}},
			},
			expectedResult: futureBlocks{
				{treeInterval: reachabilityInterval{start: 1, end: 3}},
				{treeInterval: reachabilityInterval{start: 4, end: 67}},
				{treeInterval: reachabilityInterval{start: 65, end: 78}},
				{treeInterval: reachabilityInterval{start: 657, end: 789}},
				{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
				{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
			},
		},
		{
			toInsert: []*blockNode{
				{treeInterval: reachabilityInterval{start: 88, end: 97}},
			},
			expectedResult: futureBlocks{
				{treeInterval: reachabilityInterval{start: 1, end: 3}},
				{treeInterval: reachabilityInterval{start: 4, end: 67}},
				{treeInterval: reachabilityInterval{start: 67, end: 77}},
				{treeInterval: reachabilityInterval{start: 88, end: 97}},
				{treeInterval: reachabilityInterval{start: 657, end: 789}},
				{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
				{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
			},
		},
		{
			toInsert: []*blockNode{
				{treeInterval: reachabilityInterval{start: 88, end: 97}},
				{treeInterval: reachabilityInterval{start: 3000, end: 3010}},
			},
			expectedResult: futureBlocks{
				{treeInterval: reachabilityInterval{start: 1, end: 3}},
				{treeInterval: reachabilityInterval{start: 4, end: 67}},
				{treeInterval: reachabilityInterval{start: 67, end: 77}},
				{treeInterval: reachabilityInterval{start: 88, end: 97}},
				{treeInterval: reachabilityInterval{start: 657, end: 789}},
				{treeInterval: reachabilityInterval{start: 1000, end: 1000}},
				{treeInterval: reachabilityInterval{start: 1920, end: 1921}},
				{treeInterval: reachabilityInterval{start: 3000, end: 3010}},
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
