package blockdag

import (
	"reflect"
	"strings"
	"testing"
)

func TestAddChild(t *testing.T) {
	// Scenario 1: test addChild in a chain
	//             root -> a -> b -> c...
	// Create the root node of a new reachability tree
	root := newReachabilityTreeNode(&blockNode{})
	root.setInterval(newReachabilityInterval(1, 100))

	// Add a chain of child nodes just before a reindex occurs (2^6=64 < 100)
	currentTip := root
	for i := 0; i < 6; i++ {
		node := newReachabilityTreeNode(&blockNode{})
		modifiedNodes, err := currentTip.addChild(node)
		if err != nil {
			t.Fatalf("TestAddChild: addChild failed: %s", err)
		}

		// Expect only the node and its parent to be affected
		expectedModifiedNodes := []*reachabilityTreeNode{currentTip, node}
		if !reflect.DeepEqual(modifiedNodes, expectedModifiedNodes) {
			t.Fatalf("TestAddChild: unexpected modifiedNodes. "+
				"want: %s, got: %s", expectedModifiedNodes, modifiedNodes)
		}

		currentTip = node
	}

	// Add another node to the tip of the chain to trigger a reindex (100 < 2^7=128)
	lastChild := newReachabilityTreeNode(&blockNode{})
	modifiedNodes, err := currentTip.addChild(lastChild)
	if err != nil {
		t.Fatalf("TestAddChild: addChild failed: %s", err)
	}

	// Expect more than just the node and its parent to be modified but not
	// all the nodes
	if len(modifiedNodes) <= 2 && len(modifiedNodes) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of modifiedNodes.")
	}

	// Expect the tip to have an interval of 1 and remaining interval of 0
	tipInterval := lastChild.interval.size()
	if tipInterval != 1 {
		t.Fatalf("TestAddChild: unexpected tip interval size: want: 1, got: %d", tipInterval)
	}
	tipRemainingInterval := lastChild.remainingInterval.size()
	if tipRemainingInterval != 0 {
		t.Fatalf("TestAddChild: unexpected tip interval size: want: 0, got: %d", tipRemainingInterval)
	}

	// Expect all nodes to be descendant nodes of root
	currentNode := currentTip
	for currentNode != nil {
		if !root.isAncestorOf(currentNode) {
			t.Fatalf("TestAddChild: currentNode is not a descendant of root")
		}
		currentNode = currentNode.parent
	}

	// Scenario 2: test addChild where all nodes are direct descendants of root
	//             root -> a, b, c...
	// Create the root node of a new reachability tree
	root = newReachabilityTreeNode(&blockNode{})
	root.setInterval(newReachabilityInterval(1, 100))

	// Add child nodes to root just before a reindex occurs (2^6=64 < 100)
	childNodes := make([]*reachabilityTreeNode, 6)
	for i := 0; i < len(childNodes); i++ {
		childNodes[i] = newReachabilityTreeNode(&blockNode{})
		modifiedNodes, err := root.addChild(childNodes[i])
		if err != nil {
			t.Fatalf("TestAddChild: addChild failed: %s", err)
		}

		// Expect only the node and the root to be affected
		expectedModifiedNodes := []*reachabilityTreeNode{root, childNodes[i]}
		if !reflect.DeepEqual(modifiedNodes, expectedModifiedNodes) {
			t.Fatalf("TestAddChild: unexpected modifiedNodes. "+
				"want: %s, got: %s", expectedModifiedNodes, modifiedNodes)
		}
	}

	// Add another node to the root to trigger a reindex (100 < 2^7=128)
	lastChild = newReachabilityTreeNode(&blockNode{})
	modifiedNodes, err = root.addChild(lastChild)
	if err != nil {
		t.Fatalf("TestAddChild: addChild failed: %s", err)
	}

	// Expect more than just the node and the root to be modified but not
	// all the nodes
	if len(modifiedNodes) <= 2 && len(modifiedNodes) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of modifiedNodes.")
	}

	// Expect the last-added child to have an interval of 1 and remaining interval of 0
	lastChildInterval := lastChild.interval.size()
	if lastChildInterval != 1 {
		t.Fatalf("TestAddChild: unexpected lastChild interval size: want: 1, got: %d", lastChildInterval)
	}
	lastChildRemainingInterval := lastChild.remainingInterval.size()
	if lastChildRemainingInterval != 0 {
		t.Fatalf("TestAddChild: unexpected lastChild interval size: want: 0, got: %d", lastChildRemainingInterval)
	}

	// Expect all nodes to be descendant nodes of root
	for _, childNode := range childNodes {
		if !root.isAncestorOf(childNode) {
			t.Fatalf("TestAddChild: childNode is not a descendant of root")
		}
	}
}

func TestSplitFraction(t *testing.T) {
	tests := []struct {
		interval      *reachabilityInterval
		fraction      float64
		expectedLeft  *reachabilityInterval
		expectedRight *reachabilityInterval
	}{
		{
			interval:      newReachabilityInterval(1, 100),
			fraction:      0.5,
			expectedLeft:  newReachabilityInterval(1, 50),
			expectedRight: newReachabilityInterval(51, 100),
		},
		{
			interval:      newReachabilityInterval(2, 100),
			fraction:      0.5,
			expectedLeft:  newReachabilityInterval(2, 51),
			expectedRight: newReachabilityInterval(52, 100),
		},
		{
			interval:      newReachabilityInterval(1, 99),
			fraction:      0.5,
			expectedLeft:  newReachabilityInterval(1, 50),
			expectedRight: newReachabilityInterval(51, 99),
		},
		{
			interval:      newReachabilityInterval(1, 100),
			fraction:      0.2,
			expectedLeft:  newReachabilityInterval(1, 20),
			expectedRight: newReachabilityInterval(21, 100),
		},
		{
			interval:      newReachabilityInterval(1, 100),
			fraction:      0,
			expectedLeft:  newReachabilityInterval(1, 0),
			expectedRight: newReachabilityInterval(1, 100),
		},
		{
			interval:      newReachabilityInterval(1, 100),
			fraction:      1,
			expectedLeft:  newReachabilityInterval(1, 100),
			expectedRight: newReachabilityInterval(101, 100),
		},
	}

	for i, test := range tests {
		left, right, err := test.interval.splitFraction(test.fraction)
		if err != nil {
			t.Fatalf("TestSplitFraction: splitFraction unexpectedly failed in test #%d: %s", i, err)
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

func TestSplitExact(t *testing.T) {
	tests := []struct {
		interval          *reachabilityInterval
		sizes             []uint64
		expectedIntervals []*reachabilityInterval
	}{
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{50, 50},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 20, 30, 40},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 30),
				newReachabilityInterval(31, 60),
				newReachabilityInterval(61, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{0, 100},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 0),
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100, 0},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 100),
				newReachabilityInterval(101, 100),
			},
		},
	}

	for i, test := range tests {
		intervals, err := test.interval.splitExact(test.sizes)
		if err != nil {
			t.Fatalf("TestSplitExact: splitExact unexpectedly failed in test #%d: %s", i, err)
		}
		if !reflect.DeepEqual(intervals, test.expectedIntervals) {
			t.Errorf("TestSplitExact: unexpected intervals in test #%d. "+
				"want: %s, got: %s", i, test.expectedIntervals, intervals)
		}
	}
}

func TestSplitWithExponentialBias(t *testing.T) {
	tests := []struct {
		interval          *reachabilityInterval
		sizes             []uint64
		expectedIntervals []*reachabilityInterval
	}{
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{50, 50},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 20, 30, 40},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 30),
				newReachabilityInterval(31, 60),
				newReachabilityInterval(61, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{25, 25},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{1, 1},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{33, 33, 33},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 33),
				newReachabilityInterval(34, 66),
				newReachabilityInterval(67, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 15, 25},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 25),
				newReachabilityInterval(26, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{25, 15, 10},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 75),
				newReachabilityInterval(76, 90),
				newReachabilityInterval(91, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 10_000),
			sizes:    []uint64{10, 10, 20},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 20),
				newReachabilityInterval(21, 40),
				newReachabilityInterval(41, 10_000),
			},
		},
		{
			interval: newReachabilityInterval(1, 100_000),
			sizes:    []uint64{31_000, 31_000, 30_001},
			expectedIntervals: []*reachabilityInterval{
				newReachabilityInterval(1, 35_000),
				newReachabilityInterval(35_001, 69_999),
				newReachabilityInterval(70_000, 100_000),
			},
		},
	}

	for i, test := range tests {
		intervals, err := test.interval.splitWithExponentialBias(test.sizes)
		if err != nil {
			t.Fatalf("TestSplitWithExponentialBias: splitWithExponentialBias unexpectedly failed in test #%d: %s", i, err)
		}
		if !reflect.DeepEqual(intervals, test.expectedIntervals) {
			t.Errorf("TestSplitWithExponentialBias: unexpected intervals in test #%d. "+
				"want: %s, got: %s", i, test.expectedIntervals, intervals)
		}
	}
}

func TestIsInFuture(t *testing.T) {
	blocks := futureCoveringBlockSet{
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(2, 3)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 77)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
	}

	tests := []struct {
		block          *futureCoveringBlock
		expectedResult bool
	}{
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 1)}},
			expectedResult: false,
		},
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(5, 7)}},
			expectedResult: true,
		},
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 76)}},
			expectedResult: true,
		},
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(78, 100)}},
			expectedResult: false,
		},
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1980, 2000)}},
			expectedResult: false,
		},
		{
			block:          &futureCoveringBlock{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1920)}},
			expectedResult: true,
		},
	}

	for i, test := range tests {
		result := blocks.isInFuture(test.block)
		if result != test.expectedResult {
			t.Errorf("TestIsInFuture: unexpected result in test #%d. Want: %t, got: %t",
				i, test.expectedResult, result)
		}
	}
}

func TestInsertBlock(t *testing.T) {
	blocks := futureCoveringBlockSet{
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 3)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 77)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
		{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
	}

	tests := []struct {
		toInsert       []*futureCoveringBlock
		expectedResult futureCoveringBlockSet
	}{
		{
			toInsert: []*futureCoveringBlock{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(5, 7)}},
			},
			expectedResult: futureCoveringBlockSet{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 3)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 77)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
			},
		},
		{
			toInsert: []*futureCoveringBlock{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(65, 78)}},
			},
			expectedResult: futureCoveringBlockSet{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 3)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(65, 78)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
			},
		},
		{
			toInsert: []*futureCoveringBlock{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(88, 97)}},
			},
			expectedResult: futureCoveringBlockSet{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 3)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 77)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(88, 97)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
			},
		},
		{
			toInsert: []*futureCoveringBlock{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(88, 97)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(3000, 3010)}},
			},
			expectedResult: futureCoveringBlockSet{
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1, 3)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(4, 67)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(67, 77)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(88, 97)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(657, 789)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)}},
				{treeNode: &reachabilityTreeNode{interval: newReachabilityInterval(3000, 3010)}},
			},
		},
	}

	for i, test := range tests {
		// Create a clone of blocks so that we have a clean start for every test
		blocksClone := make(futureCoveringBlockSet, len(blocks))
		for i, block := range blocks {
			blocksClone[i] = block
		}

		for _, block := range test.toInsert {
			blocksClone.insertBlock(block)
		}
		if !reflect.DeepEqual(blocksClone, test.expectedResult) {
			t.Errorf("TestInsertBlock: unexpected result in test #%d. Want: %s, got: %s",
				i, test.expectedResult, blocksClone)
		}
	}
}

func TestSplitFractionErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 200)

	// Negative fraction
	_, _, err := interval.splitFraction(-0.5)
	if err == nil {
		t.Errorf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error")
	}

	// Fraction > 1
	_, _, err = interval.splitFraction(1.5)
	if err == nil {
		t.Errorf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error")
	}

	// Splitting an empty interval
	emptyInterval := newReachabilityInterval(1, 0)
	_, _, err = emptyInterval.splitFraction(0.5)
	if err == nil {
		t.Errorf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error")
	}
}

func TestSplitExactErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 199)

	// Sum of sizes greater than the size of the interval
	sizes := []uint64{50, 51}
	_, err := interval.splitExact(sizes)
	if err == nil {
		t.Errorf("TestSplitExactErrors: splitExact unexpectedly " +
			"didn't return an error")
	}

	// Sum of sizes smaller than the size of the interval
	sizes = []uint64{50, 49}
	_, err = interval.splitExact(sizes)
	if err == nil {
		t.Errorf("TestSplitExactErrors: splitExact unexpectedly " +
			"didn't return an error")
	}
}

func TestSplitWithExponentialBiasErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 199)

	// Sum of sizes greater than the size of the interval
	sizes := []uint64{50, 51}
	_, err := interval.splitWithExponentialBias(sizes)
	if err == nil {
		t.Errorf("TestSplitWithExponentialBiasErrors: splitWithExponentialBias " +
			"unexpectedly didn't return an error")
	}
}

func TestReindexIntervalErrors(t *testing.T) {
	treeNode := newReachabilityTreeNode(&blockNode{})
	treeNode.setInterval(newReachabilityInterval(0, 99))

	var err error
	currentTreeNode := treeNode
	for i := 0; i < 100; i++ {
		childTreeNode := newReachabilityTreeNode(&blockNode{})
		_, err = currentTreeNode.addChild(childTreeNode)
		if err != nil {
			break
		}
		currentTreeNode = childTreeNode
	}

	if err == nil {
		t.Fatalf("TestReindexIntervalErrors: reindexIntervals " +
			"unexpectedly didn't return an error")
	}
	if !strings.Contains(err.Error(), "missing tree parent during reindexing") {
		t.Fatalf("TestReindexIntervalErrors: reindexIntervals "+
			"returned an expected error: %s", err)
	}
}

func TestFutureCoveringBlockSetString(t *testing.T) {
	treeNodeA := newReachabilityTreeNode(&blockNode{})
	treeNodeA.setInterval(newReachabilityInterval(123, 456))
	treeNodeB := newReachabilityTreeNode(&blockNode{})
	treeNodeB.setInterval(newReachabilityInterval(457, 789))
	futureCoveringSet := futureCoveringBlockSet{
		&futureCoveringBlock{treeNode: treeNodeA},
		&futureCoveringBlock{treeNode: treeNodeB},
	}

	str := futureCoveringSet.String()
	expectedStr := "[123,456][457,789]"
	if str != expectedStr {
		t.Fatalf("TestFutureCoveringBlockSetString: unexpected "+
			"string. Want: %s, got: %s", expectedStr, str)
	}
}

func TestReachabilityTreeNodeString(t *testing.T) {
	treeNodeA := newReachabilityTreeNode(&blockNode{})
	treeNodeA.setInterval(newReachabilityInterval(100, 199))
	treeNodeB1 := newReachabilityTreeNode(&blockNode{})
	treeNodeB1.setInterval(newReachabilityInterval(100, 150))
	treeNodeB2 := newReachabilityTreeNode(&blockNode{})
	treeNodeB2.setInterval(newReachabilityInterval(150, 199))
	treeNodeC := newReachabilityTreeNode(&blockNode{})
	treeNodeC.setInterval(newReachabilityInterval(100, 149))
	treeNodeA.children = []*reachabilityTreeNode{treeNodeB1, treeNodeB2}
	treeNodeB2.children = []*reachabilityTreeNode{treeNodeC}

	str := treeNodeA.String()
	expectedStr := "[100,149]\n[100,150][150,199]\n[100,199]"
	if str != expectedStr {
		t.Fatalf("TestReachabilityTreeNodeString: unexpected "+
			"string. Want: %s, got: %s", expectedStr, str)
	}
}
