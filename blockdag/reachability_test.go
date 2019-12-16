package blockdag

import "testing"

func TestIsFutureBlock(t *testing.T) {
	blocks := futureBlocks{
		{interval: reachabilityInterval{start: 1, end: 3}},
		{interval: reachabilityInterval{start: 4, end: 67}},
		{interval: reachabilityInterval{start: 67, end: 77}},
		{interval: reachabilityInterval{start: 657, end: 789}},
		{interval: reachabilityInterval{start: 1000, end: 1000}},
		{interval: reachabilityInterval{start: 1920, end: 1921}},
	}

	tests := []struct {
		block          *blockNode
		expectedResult bool
	}{
		{
			block:          &blockNode{interval: reachabilityInterval{start: 0, end: 0}},
			expectedResult: false,
		},
		{
			block:          &blockNode{interval: reachabilityInterval{start: 5, end: 7}},
			expectedResult: true,
		},
		{
			block:          &blockNode{interval: reachabilityInterval{start: 67, end: 76}},
			expectedResult: true,
		},
		{
			block:          &blockNode{interval: reachabilityInterval{start: 78, end: 100}},
			expectedResult: false,
		},
		{
			block:          &blockNode{interval: reachabilityInterval{start: 1980, end: 2000}},
			expectedResult: false,
		},
		{
			block:          &blockNode{interval: reachabilityInterval{start: 1920, end: 1920}},
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
