package blockdag

import (
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/util/daghash"
	"reflect"
	"strings"
	"testing"
)

func TestAddChild(t *testing.T) {
	// Scenario 1: test addChild in a chain
	//             root -> a -> b -> c...
	// Create the root node of a new reachability tree
	root := newReachabilityTreeNode(&blockNode{})
	root.interval = newReachabilityInterval(1, 100)

	// Add a chain of child nodes just before a reindex occurs (2^6=64 < 100)
	currentTip := root
	for i := 0; i < 6; i++ {
		node := newReachabilityTreeNode(&blockNode{})
		modifiedNodes := newModifiedTreeNodes()
		err := currentTip.addChild(node, root, modifiedNodes)
		if err != nil {
			t.Fatalf("TestAddChild: addChild failed: %s", err)
		}

		// Expect only the node and its parent to be affected
		expectedModifiedNodes := newModifiedTreeNodes(currentTip, node)
		if !reflect.DeepEqual(modifiedNodes, expectedModifiedNodes) {
			t.Fatalf("TestAddChild: unexpected modifiedNodes. "+
				"want: %s, got: %s", expectedModifiedNodes, modifiedNodes)
		}

		currentTip = node
	}

	// Add another node to the tip of the chain to trigger a reindex (100 < 2^7=128)
	lastChild := newReachabilityTreeNode(&blockNode{})
	modifiedNodes := newModifiedTreeNodes()
	err := currentTip.addChild(lastChild, root, modifiedNodes)
	if err != nil {
		t.Fatalf("TestAddChild: addChild failed: %s", err)
	}

	// Expect more than just the node and its parent to be modified but not
	// all the nodes
	if len(modifiedNodes) <= 2 && len(modifiedNodes) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of modifiedNodes.")
	}

	// Expect the tip to have an interval of 1 and remaining interval of 0 both before and after
	tipInterval := lastChild.interval.size()
	if tipInterval != 1 {
		t.Fatalf("TestAddChild: unexpected tip interval size: want: 1, got: %d", tipInterval)
	}
	tipRemainingIntervalBefore := lastChild.remainingIntervalBefore().size()
	if tipRemainingIntervalBefore != 0 {
		t.Fatalf("TestAddChild: unexpected tip interval before size: want: 0, got: %d", tipRemainingIntervalBefore)
	}
	tipRemainingIntervalAfter := lastChild.remainingIntervalAfter().size()
	if tipRemainingIntervalAfter != 0 {
		t.Fatalf("TestAddChild: unexpected tip interval after size: want: 0, got: %d", tipRemainingIntervalAfter)
	}

	// Expect all nodes to be descendant nodes of root
	currentNode := currentTip
	for currentNode != root {
		if !root.isAncestorOf(currentNode) {
			t.Fatalf("TestAddChild: currentNode is not a descendant of root")
		}
		currentNode = currentNode.parent
	}

	// Scenario 2: test addChild where all nodes are direct descendants of root
	//             root -> a, b, c...
	// Create the root node of a new reachability tree
	root = newReachabilityTreeNode(&blockNode{})
	root.interval = newReachabilityInterval(1, 100)

	// Add child nodes to root just before a reindex occurs (2^6=64 < 100)
	childNodes := make([]*reachabilityTreeNode, 6)
	for i := 0; i < len(childNodes); i++ {
		childNodes[i] = newReachabilityTreeNode(&blockNode{})
		modifiedNodes := newModifiedTreeNodes()
		err := root.addChild(childNodes[i], root, modifiedNodes)
		if err != nil {
			t.Fatalf("TestAddChild: addChild failed: %s", err)
		}

		// Expect only the node and the root to be affected
		expectedModifiedNodes := newModifiedTreeNodes(root, childNodes[i])
		if !reflect.DeepEqual(modifiedNodes, expectedModifiedNodes) {
			t.Fatalf("TestAddChild: unexpected modifiedNodes. "+
				"want: %s, got: %s", expectedModifiedNodes, modifiedNodes)
		}
	}

	// Add another node to the root to trigger a reindex (100 < 2^7=128)
	lastChild = newReachabilityTreeNode(&blockNode{})
	modifiedNodes = newModifiedTreeNodes()
	err = root.addChild(lastChild, root, modifiedNodes)
	if err != nil {
		t.Fatalf("TestAddChild: addChild failed: %s", err)
	}

	// Expect more than just the node and the root to be modified but not
	// all the nodes
	if len(modifiedNodes) <= 2 && len(modifiedNodes) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of modifiedNodes.")
	}

	// Expect the last-added child to have an interval of 1 and remaining interval of 0 both before and after
	lastChildInterval := lastChild.interval.size()
	if lastChildInterval != 1 {
		t.Fatalf("TestAddChild: unexpected lastChild interval size: want: 1, got: %d", lastChildInterval)
	}
	lastChildRemainingIntervalBefore := lastChild.remainingIntervalBefore().size()
	if lastChildRemainingIntervalBefore != 0 {
		t.Fatalf("TestAddChild: unexpected lastChild interval before size: want: 0, got: %d", lastChildRemainingIntervalBefore)
	}
	lastChildRemainingIntervalAfter := lastChild.remainingIntervalAfter().size()
	if lastChildRemainingIntervalAfter != 0 {
		t.Fatalf("TestAddChild: unexpected lastChild interval after size: want: 0, got: %d", lastChildRemainingIntervalAfter)
	}

	// Expect all nodes to be descendant nodes of root
	for _, childNode := range childNodes {
		if !root.isAncestorOf(childNode) {
			t.Fatalf("TestAddChild: childNode is not a descendant of root")
		}
	}
}

func TestReachabilityTreeNodeIsAncestorOf(t *testing.T) {
	root := newReachabilityTreeNode(&blockNode{})
	currentTip := root
	const numberOfDescendants = 6
	descendants := make([]*reachabilityTreeNode, numberOfDescendants)
	for i := 0; i < numberOfDescendants; i++ {
		node := newReachabilityTreeNode(&blockNode{})
		err := currentTip.addChild(node, root, newModifiedTreeNodes())
		if err != nil {
			t.Fatalf("TestReachabilityTreeNodeIsAncestorOf: addChild failed: %s", err)
		}
		descendants[i] = node
		currentTip = node
	}

	// Expect all descendants to be in the future of root
	for _, node := range descendants {
		if !root.isAncestorOf(node) {
			t.Fatalf("TestReachabilityTreeNodeIsAncestorOf: node is not a descendant of root")
		}
	}

	if !root.isAncestorOf(root) {
		t.Fatalf("TestReachabilityTreeNodeIsAncestorOf: root is expected to be an ancestor of root")
	}
}

func TestIntervalContains(t *testing.T) {
	tests := []struct {
		name              string
		this, other       *reachabilityInterval
		thisContainsOther bool
	}{
		{
			name:              "this == other",
			this:              newReachabilityInterval(10, 100),
			other:             newReachabilityInterval(10, 100),
			thisContainsOther: true,
		},
		{
			name:              "this.start == other.start && this.end < other.end",
			this:              newReachabilityInterval(10, 90),
			other:             newReachabilityInterval(10, 100),
			thisContainsOther: false,
		},
		{
			name:              "this.start == other.start && this.end > other.end",
			this:              newReachabilityInterval(10, 100),
			other:             newReachabilityInterval(10, 90),
			thisContainsOther: true,
		},
		{
			name:              "this.start > other.start && this.end == other.end",
			this:              newReachabilityInterval(20, 100),
			other:             newReachabilityInterval(10, 100),
			thisContainsOther: false,
		},
		{
			name:              "this.start < other.start && this.end == other.end",
			this:              newReachabilityInterval(10, 100),
			other:             newReachabilityInterval(20, 100),
			thisContainsOther: true,
		},
		{
			name:              "this.start > other.start && this.end < other.end",
			this:              newReachabilityInterval(20, 90),
			other:             newReachabilityInterval(10, 100),
			thisContainsOther: false,
		},
		{
			name:              "this.start < other.start && this.end > other.end",
			this:              newReachabilityInterval(10, 100),
			other:             newReachabilityInterval(20, 90),
			thisContainsOther: true,
		},
	}

	for _, test := range tests {
		if thisContainsOther := test.this.contains(test.other); thisContainsOther != test.thisContainsOther {
			t.Errorf("test.this.contains(test.other) is expected to be %t but got %t",
				test.thisContainsOther, thisContainsOther)
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

func TestHasAncestorOf(t *testing.T) {
	treeNodes := futureCoveringTreeNodeSet{
		&reachabilityTreeNode{interval: newReachabilityInterval(2, 3)},
		&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
		&reachabilityTreeNode{interval: newReachabilityInterval(67, 77)},
		&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
		&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
		&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
	}

	tests := []struct {
		treeNode       *reachabilityTreeNode
		expectedResult bool
	}{
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(1, 1)},
			expectedResult: false,
		},
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(5, 7)},
			expectedResult: true,
		},
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(67, 76)},
			expectedResult: true,
		},
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(78, 100)},
			expectedResult: false,
		},
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(1980, 2000)},
			expectedResult: false,
		},
		{
			treeNode:       &reachabilityTreeNode{interval: newReachabilityInterval(1920, 1920)},
			expectedResult: true,
		},
	}

	for i, test := range tests {
		result := treeNodes.hasAncestorOf(test.treeNode)
		if result != test.expectedResult {
			t.Errorf("TestHasAncestorOf: unexpected result in test #%d. Want: %t, got: %t",
				i, test.expectedResult, result)
		}
	}
}

func TestInsertNode(t *testing.T) {
	treeNodes := futureCoveringTreeNodeSet{
		&reachabilityTreeNode{interval: newReachabilityInterval(1, 3)},
		&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
		&reachabilityTreeNode{interval: newReachabilityInterval(67, 77)},
		&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
		&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
		&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
	}

	tests := []struct {
		toInsert       []*reachabilityTreeNode
		expectedResult futureCoveringTreeNodeSet
	}{
		{
			toInsert: []*reachabilityTreeNode{
				{interval: newReachabilityInterval(5, 7)},
			},
			expectedResult: futureCoveringTreeNodeSet{
				&reachabilityTreeNode{interval: newReachabilityInterval(1, 3)},
				&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
				&reachabilityTreeNode{interval: newReachabilityInterval(67, 77)},
				&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
			},
		},
		{
			toInsert: []*reachabilityTreeNode{
				{interval: newReachabilityInterval(65, 78)},
			},
			expectedResult: futureCoveringTreeNodeSet{
				&reachabilityTreeNode{interval: newReachabilityInterval(1, 3)},
				&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
				&reachabilityTreeNode{interval: newReachabilityInterval(65, 78)},
				&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
			},
		},
		{
			toInsert: []*reachabilityTreeNode{
				{interval: newReachabilityInterval(88, 97)},
			},
			expectedResult: futureCoveringTreeNodeSet{
				&reachabilityTreeNode{interval: newReachabilityInterval(1, 3)},
				&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
				&reachabilityTreeNode{interval: newReachabilityInterval(67, 77)},
				&reachabilityTreeNode{interval: newReachabilityInterval(88, 97)},
				&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
			},
		},
		{
			toInsert: []*reachabilityTreeNode{
				{interval: newReachabilityInterval(88, 97)},
				{interval: newReachabilityInterval(3000, 3010)},
			},
			expectedResult: futureCoveringTreeNodeSet{
				&reachabilityTreeNode{interval: newReachabilityInterval(1, 3)},
				&reachabilityTreeNode{interval: newReachabilityInterval(4, 67)},
				&reachabilityTreeNode{interval: newReachabilityInterval(67, 77)},
				&reachabilityTreeNode{interval: newReachabilityInterval(88, 97)},
				&reachabilityTreeNode{interval: newReachabilityInterval(657, 789)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1000, 1000)},
				&reachabilityTreeNode{interval: newReachabilityInterval(1920, 1921)},
				&reachabilityTreeNode{interval: newReachabilityInterval(3000, 3010)},
			},
		},
	}

	for i, test := range tests {
		// Create a clone of treeNodes so that we have a clean start for every test
		treeNodesClone := make(futureCoveringTreeNodeSet, len(treeNodes))
		for i, treeNode := range treeNodes {
			treeNodesClone[i] = treeNode
		}

		for _, treeNode := range test.toInsert {
			treeNodesClone.insertNode(treeNode)
		}
		if !reflect.DeepEqual(treeNodesClone, test.expectedResult) {
			t.Errorf("TestInsertNode: unexpected result in test #%d. Want: %s, got: %s",
				i, test.expectedResult, treeNodesClone)
		}
	}
}

func TestSplitFractionErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 200)

	// Negative fraction
	_, _, err := interval.splitFraction(-0.5)
	if err == nil {
		t.Fatalf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error for a negative fraction")
	}
	expectedErrSubstring := "fraction must be between 0 and 1"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitFractionErrors: splitFraction returned wrong error "+
			"for a negative fraction. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}

	// Fraction > 1
	_, _, err = interval.splitFraction(1.5)
	if err == nil {
		t.Fatalf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error for a fraction greater than 1")
	}
	expectedErrSubstring = "fraction must be between 0 and 1"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitFractionErrors: splitFraction returned wrong error "+
			"for a fraction greater than 1. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}

	// Splitting an empty interval
	emptyInterval := newReachabilityInterval(1, 0)
	_, _, err = emptyInterval.splitFraction(0.5)
	if err == nil {
		t.Fatalf("TestSplitFractionErrors: splitFraction unexpectedly " +
			"didn't return an error for an empty interval")
	}
	expectedErrSubstring = "cannot split an empty interval"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitFractionErrors: splitFraction returned wrong error "+
			"for an empty interval. "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}
}

func TestSplitExactErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 199)

	// Sum of sizes greater than the size of the interval
	sizes := []uint64{50, 51}
	_, err := interval.splitExact(sizes)
	if err == nil {
		t.Fatalf("TestSplitExactErrors: splitExact unexpectedly " +
			"didn't return an error for (sum of sizes) > (size of interval)")
	}
	expectedErrSubstring := "sum of sizes must be equal to the interval's size"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitExactErrors: splitExact returned wrong error "+
			"for (sum of sizes) > (size of interval). "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}

	// Sum of sizes smaller than the size of the interval
	sizes = []uint64{50, 49}
	_, err = interval.splitExact(sizes)
	if err == nil {
		t.Fatalf("TestSplitExactErrors: splitExact unexpectedly " +
			"didn't return an error for (sum of sizes) < (size of interval)")
	}
	expectedErrSubstring = "sum of sizes must be equal to the interval's size"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitExactErrors: splitExact returned wrong error "+
			"for (sum of sizes) < (size of interval). "+
			"Want: %s, got: %s", expectedErrSubstring, err)
	}
}

func TestSplitWithExponentialBiasErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 199)

	// Sum of sizes greater than the size of the interval
	sizes := []uint64{50, 51}
	_, err := interval.splitWithExponentialBias(sizes)
	if err == nil {
		t.Fatalf("TestSplitWithExponentialBiasErrors: splitWithExponentialBias " +
			"unexpectedly didn't return an error")
	}
	expectedErrSubstring := "sum of sizes must be less than or equal to the interval's size"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Fatalf("TestSplitWithExponentialBiasErrors: splitWithExponentialBias "+
			"returned wrong error. Want: %s, got: %s", expectedErrSubstring, err)
	}
}

func TestReindexIntervalErrors(t *testing.T) {
	// Create a treeNode and give it size = 100
	treeNode := newReachabilityTreeNode(&blockNode{})
	treeNode.interval = newReachabilityInterval(0, 99)

	// Add a chain of 100 child treeNodes to treeNode
	var err error
	currentTreeNode := treeNode
	for i := 0; i < 100; i++ {
		childTreeNode := newReachabilityTreeNode(&blockNode{})
		err = currentTreeNode.addChild(childTreeNode, treeNode, newModifiedTreeNodes())
		if err != nil {
			break
		}
		currentTreeNode = childTreeNode
	}

	// At the 100th addChild we expect a reindex. This reindex should
	// fail because our initial treeNode only has size = 100, and the
	// reindex requires size > 100.
	// This simulates the case when (somehow) there's more than 2^64
	// blocks in the DAG, since the genesis block has size = 2^64.
	if err == nil {
		t.Fatalf("TestReindexIntervalErrors: reindexIntervals " +
			"unexpectedly didn't return an error")
	}
	if !strings.Contains(err.Error(), "missing tree parent during reindexing") {
		t.Fatalf("TestReindexIntervalErrors: reindexIntervals "+
			"returned an expected error: %s", err)
	}
}

func BenchmarkReindexInterval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		root := newReachabilityTreeNode(&blockNode{})

		const subTreeSize = 70000
		// We set the interval of the root to subTreeSize*2 because
		// its first child gets half of the interval, so a reindex
		// from the root should happen after adding subTreeSize
		// nodes.
		root.interval = newReachabilityInterval(0, subTreeSize*2)

		currentTreeNode := root
		for i := 0; i < subTreeSize; i++ {
			childTreeNode := newReachabilityTreeNode(&blockNode{})
			err := currentTreeNode.addChild(childTreeNode, root, newModifiedTreeNodes())
			if err != nil {
				b.Fatalf("addChild: %s", err)
			}

			currentTreeNode = childTreeNode
		}

		originalRemainingInterval := *root.remainingIntervalAfter()
		// After we added subTreeSize nodes, adding the next
		// node should lead to a reindex from root.
		fullReindexTriggeringNode := newReachabilityTreeNode(&blockNode{})
		b.StartTimer()
		err := currentTreeNode.addChild(fullReindexTriggeringNode, root, newModifiedTreeNodes())
		b.StopTimer()
		if err != nil {
			b.Fatalf("addChild: %s", err)
		}

		if *root.remainingIntervalAfter() == originalRemainingInterval {
			b.Fatal("Expected a reindex from root, but it didn't happen")
		}
	}
}

func TestFutureCoveringTreeNodeSetString(t *testing.T) {
	treeNodeA := newReachabilityTreeNode(&blockNode{})
	treeNodeA.interval = newReachabilityInterval(123, 456)
	treeNodeB := newReachabilityTreeNode(&blockNode{})
	treeNodeB.interval = newReachabilityInterval(457, 789)
	futureCoveringSet := futureCoveringTreeNodeSet{treeNodeA, treeNodeB}

	str := futureCoveringSet.String()
	expectedStr := "[123,456][457,789]"
	if str != expectedStr {
		t.Fatalf("TestFutureCoveringTreeNodeSetString: unexpected "+
			"string. Want: %s, got: %s", expectedStr, str)
	}
}

func TestReachabilityTreeNodeString(t *testing.T) {
	treeNodeA := newReachabilityTreeNode(&blockNode{})
	treeNodeA.interval = newReachabilityInterval(100, 199)
	treeNodeB1 := newReachabilityTreeNode(&blockNode{})
	treeNodeB1.interval = newReachabilityInterval(100, 150)
	treeNodeB2 := newReachabilityTreeNode(&blockNode{})
	treeNodeB2.interval = newReachabilityInterval(150, 199)
	treeNodeC := newReachabilityTreeNode(&blockNode{})
	treeNodeC.interval = newReachabilityInterval(100, 149)
	treeNodeA.children = []*reachabilityTreeNode{treeNodeB1, treeNodeB2}
	treeNodeB2.children = []*reachabilityTreeNode{treeNodeC}

	str := treeNodeA.String()
	expectedStr := "[100,149]\n[100,150][150,199]\n[100,199]"
	if str != expectedStr {
		t.Fatalf("TestReachabilityTreeNodeString: unexpected "+
			"string. Want: %s, got: %s", expectedStr, str)
	}
}

func TestIsInPast(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestIsInPast", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("TestIsInPast: Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Add a chain of two blocks above the genesis. This will be the
	// selected parent chain.
	blockA := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	blockB := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockA.BlockHash()}, nil)

	// Add another block above the genesis
	blockC := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	nodeC, ok := dag.index.LookupNode(blockC.BlockHash())
	if !ok {
		t.Fatalf("TestIsInPast: block C is not in the block index")
	}

	// Add a block whose parents are the two tips
	blockD := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{blockB.BlockHash(), blockC.BlockHash()}, nil)
	nodeD, ok := dag.index.LookupNode(blockD.BlockHash())
	if !ok {
		t.Fatalf("TestIsInPast: block C is not in the block index")
	}

	// Make sure that node C is in the past of node D
	isInFuture, err := dag.reachabilityTree.isInPast(nodeC, nodeD)
	if err != nil {
		t.Fatalf("TestIsInPast: isInPast unexpectedly failed: %s", err)
	}
	if !isInFuture {
		t.Fatalf("TestIsInPast: node C is unexpectedly not the past of node D")
	}
}

func TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestAddChildThatPointsDirectlyToTheSelectedParentChainBelowReindexRoot",
		true, Config{DAGParams: &dagconfig.SimnetParams})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Set the reindex window to a low number to make this test run fast
	originalReachabilityReindexWindow := reachabilityReindexWindow
	reachabilityReindexWindow = 10
	defer func() {
		reachabilityReindexWindow = originalReachabilityReindexWindow
	}()

	// Add a block on top of the genesis block
	chainRootBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)

	// Add chain of reachabilityReindexWindow blocks above chainRootBlock.
	// This should move the reindex root
	chainRootBlockTipHash := chainRootBlock.BlockHash()
	for i := uint64(0); i < reachabilityReindexWindow; i++ {
		chainBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{chainRootBlockTipHash}, nil)
		chainRootBlockTipHash = chainBlock.BlockHash()
	}

	// Add another block over genesis
	PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
}

func TestUpdateReindexRoot(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestUpdateReindexRoot", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Set the reindex window to a low number to make this test run fast
	originalReachabilityReindexWindow := reachabilityReindexWindow
	reachabilityReindexWindow = 10
	defer func() {
		reachabilityReindexWindow = originalReachabilityReindexWindow
	}()

	// Add two blocks on top of the genesis block
	chain1RootBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	chain2RootBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)

	// Add chain of reachabilityReindexWindow - 1 blocks above chain1RootBlock and
	// chain2RootBlock, respectively. This should not move the reindex root
	chain1RootBlockTipHash := chain1RootBlock.BlockHash()
	chain2RootBlockTipHash := chain2RootBlock.BlockHash()
	genesisTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(dag.genesis.hash)
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
		chain1Block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{chain1RootBlockTipHash}, nil)
		chain1RootBlockTipHash = chain1Block.BlockHash()

		chain2Block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{chain2RootBlockTipHash}, nil)
		chain2RootBlockTipHash = chain2Block.BlockHash()

		if dag.reachabilityTree.reindexRoot != genesisTreeNode {
			t.Fatalf("reindex root unexpectedly moved")
		}
	}

	// Add another block over chain1. This will move the reindex root to chain1RootBlock
	PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{chain1RootBlockTipHash}, nil)

	// Make sure that chain1RootBlock is now the reindex root
	chain1RootTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(chain1RootBlock.BlockHash())
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	if dag.reachabilityTree.reindexRoot != chain1RootTreeNode {
		t.Fatalf("chain1RootBlock is not the reindex root after reindex")
	}

	// Make sure that tight intervals have been applied to chain2. Since
	// we added reachabilityReindexWindow-1 blocks to chain2, the size
	// of the interval at its root should be equal to reachabilityReindexWindow
	chain2RootTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(chain2RootBlock.BlockHash())
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	if chain2RootTreeNode.interval.size() != reachabilityReindexWindow {
		t.Fatalf("got unexpected chain2RootNode interval. Want: %d, got: %d",
			chain2RootTreeNode.interval.size(), reachabilityReindexWindow)
	}

	// Make sure that the rest of the interval has been allocated to
	// chain1RootNode, minus slack from both sides
	expectedChain1RootIntervalSize := genesisTreeNode.interval.size() - 1 -
		chain2RootTreeNode.interval.size() - 2*reachabilityReindexSlack
	if chain1RootTreeNode.interval.size() != expectedChain1RootIntervalSize {
		t.Fatalf("got unexpected chain1RootNode interval. Want: %d, got: %d",
			chain1RootTreeNode.interval.size(), expectedChain1RootIntervalSize)
	}
}

func TestReindexIntervalsEarlierThanReindexRoot(t *testing.T) {
	// Create a new database and DAG instance to run tests against.
	dag, teardownFunc, err := DAGSetup("TestReindexIntervalsEarlierThanReindexRoot", true, Config{
		DAGParams: &dagconfig.SimnetParams,
	})
	if err != nil {
		t.Fatalf("Failed to setup DAG instance: %v", err)
	}
	defer teardownFunc()

	// Set the reindex window and slack to low numbers to make this test
	// run fast
	originalReachabilityReindexWindow := reachabilityReindexWindow
	originalReachabilityReindexSlack := reachabilityReindexSlack
	reachabilityReindexWindow = 10
	reachabilityReindexSlack = 5
	defer func() {
		reachabilityReindexWindow = originalReachabilityReindexWindow
		reachabilityReindexSlack = originalReachabilityReindexSlack
	}()

	// Add three children to the genesis: leftBlock, centerBlock, rightBlock
	leftBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	centerBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)
	rightBlock := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{dag.genesis.hash}, nil)

	// Add a chain of reachabilityReindexWindow blocks above centerBlock.
	// This will move the reindex root to centerBlock
	centerTipHash := centerBlock.BlockHash()
	for i := uint64(0); i < reachabilityReindexWindow; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{centerTipHash}, nil)
		centerTipHash = block.BlockHash()
	}

	// Make sure that centerBlock is now the reindex root
	centerTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(centerBlock.BlockHash())
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	if dag.reachabilityTree.reindexRoot != centerTreeNode {
		t.Fatalf("centerBlock is not the reindex root after reindex")
	}

	// Get the current interval for leftBlock. The reindex should have
	// resulted in a tight interval there
	leftTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(leftBlock.BlockHash())
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	if leftTreeNode.interval.size() != 1 {
		t.Fatalf("leftBlock interval not tight after reindex")
	}

	// Get the current interval for rightBlock. The reindex should have
	// resulted in a tight interval there
	rightTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(rightBlock.BlockHash())
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	if rightTreeNode.interval.size() != 1 {
		t.Fatalf("rightBlock interval not tight after reindex")
	}

	// Get the current interval for centerBlock. Its interval should be:
	// genesisInterval - 1 - leftInterval - leftSlack - rightInterval - rightSlack
	genesisTreeNode, err := dag.reachabilityTree.store.treeNodeByBlockHash(dag.genesis.hash)
	if err != nil {
		t.Fatalf("failed to get tree node: %s", err)
	}
	expectedCenterInterval := genesisTreeNode.interval.size() - 1 -
		leftTreeNode.interval.size() - reachabilityReindexSlack -
		rightTreeNode.interval.size() - reachabilityReindexSlack
	if centerTreeNode.interval.size() != expectedCenterInterval {
		t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
			expectedCenterInterval, centerTreeNode.interval.size())
	}

	// Add a chain of reachabilityReindexWindow - 1 blocks above leftBlock.
	// Each addition will trigger a low-than-reindex-root reindex. We
	// expect the centerInterval to shrink by 1 each time, but its child
	// to remain unaffected
	treeChildOfCenterBlock := centerTreeNode.children[0]
	treeChildOfCenterBlockOriginalIntervalSize := treeChildOfCenterBlock.interval.size()
	leftTipHash := leftBlock.BlockHash()
	for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{leftTipHash}, nil)
		leftTipHash = block.BlockHash()

		expectedCenterInterval--
		if centerTreeNode.interval.size() != expectedCenterInterval {
			t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
				expectedCenterInterval, centerTreeNode.interval.size())
		}

		if treeChildOfCenterBlock.interval.size() != treeChildOfCenterBlockOriginalIntervalSize {
			t.Fatalf("the interval of centerBlock's child unexpectedly changed")
		}
	}

	// Add a chain of reachabilityReindexWindow - 1 blocks above rightBlock.
	// Each addition will trigger a low-than-reindex-root reindex. We
	// expect the centerInterval to shrink by 1 each time, but its child
	// to remain unaffected
	rightTipHash := rightBlock.BlockHash()
	for i := uint64(0); i < reachabilityReindexWindow-1; i++ {
		block := PrepareAndProcessBlockForTest(t, dag, []*daghash.Hash{rightTipHash}, nil)
		rightTipHash = block.BlockHash()

		expectedCenterInterval--
		if centerTreeNode.interval.size() != expectedCenterInterval {
			t.Fatalf("unexpected centerBlock interval. Want: %d, got: %d",
				expectedCenterInterval, centerTreeNode.interval.size())
		}

		if treeChildOfCenterBlock.interval.size() != treeChildOfCenterBlockOriginalIntervalSize {
			t.Fatalf("the interval of centerBlock's child unexpectedly changed")
		}
	}
}
