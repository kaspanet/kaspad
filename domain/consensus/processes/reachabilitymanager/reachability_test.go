package reachabilitymanager

import (
	"encoding/binary"
	"reflect"
	"strings"
	"testing"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type reachabilityDataStoreMock struct {
	reachabilityDataStaging        map[externalapi.DomainHash]model.ReachabilityData
	recorder                       map[externalapi.DomainHash]struct{}
	reachabilityReindexRootStaging *externalapi.DomainHash
}

func (r *reachabilityDataStoreMock) Discard() {
	panic("implement me")
}

func (r *reachabilityDataStoreMock) Commit(_ model.DBTransaction) error {
	panic("implement me")
}

func (r *reachabilityDataStoreMock) StageReachabilityData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, reachabilityData model.ReachabilityData) {

	r.reachabilityDataStaging[*blockHash] = reachabilityData
	r.recorder[*blockHash] = struct{}{}
}

func (r *reachabilityDataStoreMock) StageReachabilityReindexRoot(stagingArea *model.StagingArea, reachabilityReindexRoot *externalapi.DomainHash) {
	r.reachabilityReindexRootStaging = reachabilityReindexRoot
}

func (r *reachabilityDataStoreMock) IsAnythingStaged(*model.StagingArea) bool {
	panic("implement me")
}

func (r *reachabilityDataStoreMock) ReachabilityData(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (model.ReachabilityData, error) {

	return r.reachabilityDataStaging[*blockHash], nil
}

func (r *reachabilityDataStoreMock) HasReachabilityData(dbContext model.DBReader, stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (bool, error) {
	_, ok := r.reachabilityDataStaging[*blockHash]
	return ok, nil
}

func (r *reachabilityDataStoreMock) ReachabilityReindexRoot(dbContext model.DBReader, stagingArea *model.StagingArea) (*externalapi.DomainHash, error) {
	return r.reachabilityReindexRootStaging, nil
}

func (r *reachabilityDataStoreMock) isRecorderContainsOnly(nodes ...*externalapi.DomainHash) bool {
	if len(r.recorder) != len(nodes) {
		return false
	}

	for _, node := range nodes {
		if _, ok := r.recorder[*node]; !ok {
			return false
		}
	}

	return true
}

func (r *reachabilityDataStoreMock) resetRecorder() {
	r.recorder = make(map[externalapi.DomainHash]struct{})
}

func newReachabilityDataStoreMock() *reachabilityDataStoreMock {
	return &reachabilityDataStoreMock{
		reachabilityDataStaging:        make(map[externalapi.DomainHash]model.ReachabilityData),
		recorder:                       make(map[externalapi.DomainHash]struct{}),
		reachabilityReindexRootStaging: nil,
	}
}

type fatalfer interface {
	Fatalf(format string, args ...interface{})
}

type testHelper struct {
	*reachabilityManager
	t           fatalfer
	dataStore   *reachabilityDataStoreMock
	hashCounter uint64
}

func (th *testHelper) generateHash() *externalapi.DomainHash {
	var hashArray [externalapi.DomainHashSize]byte
	binary.LittleEndian.PutUint64(hashArray[:], th.hashCounter)
	th.hashCounter++
	return externalapi.NewDomainHashFromByteArray(&hashArray)
}

func (th *testHelper) newNode() *externalapi.DomainHash {
	node := th.generateHash()
	th.stageData(node, newReachabilityTreeData())
	return node
}

func (th *testHelper) newNodeWithInterval(interval *model.ReachabilityInterval) *externalapi.DomainHash {
	node := th.newNode()
	err := th.stageInterval(node, interval)
	if err != nil {
		th.t.Fatalf("stageInteval: %s", err)
	}

	return node
}

func (th *testHelper) getInterval(node *externalapi.DomainHash) *model.ReachabilityInterval {
	interval, err := th.interval(node)
	if err != nil {
		th.t.Fatalf("interval: %s", err)
	}

	return interval
}

func (th *testHelper) getIntervalSize(node *externalapi.DomainHash) uint64 {
	return intervalSize(th.getInterval(node))
}

func (th *testHelper) remainingIntervalBefore(node *externalapi.DomainHash) *model.ReachabilityInterval {
	interval, err := th.reachabilityManager.remainingIntervalBefore(node)
	if err != nil {
		th.t.Fatalf("remainingIntervalBefore: %s", err)
	}

	return interval
}

func (th *testHelper) remainingIntervalAfter(node *externalapi.DomainHash) *model.ReachabilityInterval {
	interval, err := th.reachabilityManager.remainingIntervalAfter(node)
	if err != nil {
		th.t.Fatalf("remainingIntervalAfter: %s", err)
	}

	return interval
}

func (th *testHelper) addChild(node, child, reindexRoot *externalapi.DomainHash) {
	err := th.reachabilityManager.addChild(node, child, reindexRoot)
	if err != nil {
		th.t.Fatalf("addChild: %s", err)
	}
}

func (th *testHelper) isReachabilityTreeAncestorOf(node, other *externalapi.DomainHash) bool {
	isReachabilityTreeAncestorOf, err := th.reachabilityManager.IsReachabilityTreeAncestorOf(node, other)
	if err != nil {
		th.t.Fatalf("IsReachabilityTreeAncestorOf: %s", err)
	}

	return isReachabilityTreeAncestorOf
}

func (th *testHelper) checkIsRecorderContainsOnly(nodes ...*externalapi.DomainHash) {
	if !th.dataStore.isRecorderContainsOnly(nodes...) {
		th.t.Fatalf("unexpected nodes on recorder. Want: %v, got: %v", nodes, th.dataStore.recorder)
	}
}

func (th *testHelper) resetRecorder() {
	th.dataStore.resetRecorder()
}

func newTestHelper(manager *reachabilityManager, t fatalfer, dataStore *reachabilityDataStoreMock) *testHelper {
	return &testHelper{reachabilityManager: manager, t: t, dataStore: dataStore}
}

func TestAddChild(t *testing.T) {
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	// Scenario 1: test addChild in a chain
	//             root -> a -> b -> c...
	// Create the root node of a new reachability tree
	root := helper.newNode()
	err := helper.stageInterval(root, newReachabilityInterval(1, 100))
	if err != nil {
		t.Fatalf("stageInterval: %s", err)
	}

	// Add a chain of child nodes just before a reindex occurs (2^6=64 < 100)
	currentTip := root
	for i := 0; i < 6; i++ {
		node := helper.newNode()
		helper.resetRecorder()
		helper.addChild(currentTip, node, root)

		// Expect only the node and its parent to be affected
		helper.checkIsRecorderContainsOnly(currentTip, node)
		currentTip = node
	}

	// Add another node to the tip of the chain to trigger a reindex (100 < 2^7=128)
	lastChild := helper.newNode()
	helper.resetRecorder()
	helper.addChild(currentTip, lastChild, root)

	// Expect more than just the node and its parent to be modified but not
	// all the nodes
	if len(helper.dataStore.recorder) <= 2 && len(helper.dataStore.recorder) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of staged nodes")
	}

	// Expect the tip to have an interval of 1 and remaining interval of 0 both before and after
	tipIntervalSize := helper.getIntervalSize(lastChild)
	if tipIntervalSize != 1 {
		t.Fatalf("TestAddChild: unexpected tip interval size: want: 1, got: %d", tipIntervalSize)
	}

	tipRemainingIntervalBefore := helper.remainingIntervalBefore(lastChild)

	if intervalSize(tipRemainingIntervalBefore) != 0 {
		t.Fatalf("TestAddChild: unexpected tip interval before size: want: 0, got: %d", intervalSize(tipRemainingIntervalBefore))
	}

	tipRemainingIntervalAfter := helper.remainingIntervalAfter(lastChild)
	if intervalSize(tipRemainingIntervalAfter) != 0 {
		t.Fatalf("TestAddChild: unexpected tip interval after size: want: 0, got: %d", intervalSize(tipRemainingIntervalAfter))
	}

	// Expect all nodes to be descendant nodes of root
	currentNode := currentTip
	for currentNode != root {
		isReachabilityTreeAncestorOf, err := helper.IsReachabilityTreeAncestorOf(root, currentNode)
		if err != nil {
			t.Fatalf("IsReachabilityTreeAncestorOf: %s", err)
		}
		if !isReachabilityTreeAncestorOf {
			t.Fatalf("TestAddChild: currentNode is not a descendant of root")
		}

		currentNode, err = helper.parent(currentNode)
		if err != nil {
			t.Fatalf("parent: %s", err)
		}
	}

	err = manager.validateIntervals(root)
	if err != nil {
		t.Fatal(err)
	}

	// Scenario 2: test addChild where all nodes are direct descendants of root
	//             root -> a, b, c...
	// Create the root node of a new reachability tree
	root = helper.newNode()
	err = helper.stageInterval(root, newReachabilityInterval(1, 100))
	if err != nil {
		t.Fatalf("stageInterval: %s", err)
	}

	// Add child nodes to root just before a reindex occurs (2^6=64 < 100)
	childNodes := make([]*externalapi.DomainHash, 6)
	for i := 0; i < len(childNodes); i++ {
		childNodes[i] = helper.newNode()
		helper.resetRecorder()
		helper.addChild(root, childNodes[i], root)

		// Expect only the node and the root to be affected
		helper.checkIsRecorderContainsOnly(root, childNodes[i])
	}

	// Add another node to the root to trigger a reindex (100 < 2^7=128)
	lastChild = helper.newNode()
	helper.resetRecorder()
	helper.addChild(root, lastChild, root)

	// Expect more than just the node and the root to be modified but not
	// all the nodes
	if len(helper.dataStore.recorder) <= 2 && len(helper.dataStore.recorder) >= 7 {
		t.Fatalf("TestAddChild: unexpected amount of modifiedNodes.")
	}

	// Expect the last-added child to have an interval of 1 and remaining interval of 0 both before and after
	lastChildInterval, err := helper.interval(lastChild)
	if err != nil {
		t.Fatalf("interval: %s", err)
	}

	if intervalSize(lastChildInterval) != 1 {
		t.Fatalf("TestAddChild: unexpected lastChild interval size: want: 1, got: %d", intervalSize(lastChildInterval))
	}
	lastChildRemainingIntervalBeforeSize := intervalSize(helper.remainingIntervalBefore(lastChild))
	if lastChildRemainingIntervalBeforeSize != 0 {
		t.Fatalf("TestAddChild: unexpected lastChild interval before size: want: 0, got: %d", lastChildRemainingIntervalBeforeSize)
	}
	lastChildRemainingIntervalAfterSize := intervalSize(helper.remainingIntervalAfter(lastChild))
	if lastChildRemainingIntervalAfterSize != 0 {
		t.Fatalf("TestAddChild: unexpected lastChild interval after size: want: 0, got: %d", lastChildRemainingIntervalAfterSize)
	}

	// Expect all nodes to be descendant nodes of root
	for _, childNode := range childNodes {
		isReachabilityTreeAncestorOf, err := helper.IsReachabilityTreeAncestorOf(root, childNode)
		if err != nil {
			t.Fatalf("IsReachabilityTreeAncestorOf: %s", err)
		}

		if !isReachabilityTreeAncestorOf {
			t.Fatalf("TestAddChild: childNode is not a descendant of root")
		}
	}

	err = manager.validateIntervals(root)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReachabilityTreeNodeIsAncestorOf(t *testing.T) {
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	root := helper.newNode()
	currentTip := root
	const numberOfDescendants = 6
	descendants := make([]*externalapi.DomainHash, numberOfDescendants)
	for i := 0; i < numberOfDescendants; i++ {
		node := helper.newNode()
		helper.addChild(currentTip, node, root)
		descendants[i] = node
		currentTip = node
	}

	// Expect all descendants to be in the future of root
	for _, node := range descendants {
		if !helper.isReachabilityTreeAncestorOf(root, node) {
			t.Fatalf("TestReachabilityTreeNodeIsAncestorOf: node is not a descendant of root")
		}
	}

	if !helper.isReachabilityTreeAncestorOf(root, root) {
		t.Fatalf("TestReachabilityTreeNodeIsAncestorOf: root is expected to be an ancestor of root")
	}

	err := manager.validateIntervals(root)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntervalContains(t *testing.T) {
	tests := []struct {
		name              string
		this, other       *model.ReachabilityInterval
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
		if thisContainsOther := intervalContains(test.this, test.other); thisContainsOther != test.thisContainsOther {
			t.Errorf("test.this.contains(test.other) is expected to be %t but got %t",
				test.thisContainsOther, thisContainsOther)
		}
	}
}

func TestSplitFraction(t *testing.T) {
	tests := []struct {
		interval      *model.ReachabilityInterval
		fraction      float64
		expectedLeft  *model.ReachabilityInterval
		expectedRight *model.ReachabilityInterval
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
		left, right, err := intervalSplitFraction(test.interval, test.fraction)
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
		interval          *model.ReachabilityInterval
		sizes             []uint64
		expectedIntervals []*model.ReachabilityInterval
	}{
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{50, 50},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 20, 30, 40},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 30),
				newReachabilityInterval(31, 60),
				newReachabilityInterval(61, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{0, 100},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 0),
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100, 0},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 100),
				newReachabilityInterval(101, 100),
			},
		},
	}

	for i, test := range tests {
		intervals, err := intervalSplitExact(test.interval, test.sizes)
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
		interval          *model.ReachabilityInterval
		sizes             []uint64
		expectedIntervals []*model.ReachabilityInterval
	}{
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{100},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{50, 50},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 20, 30, 40},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 30),
				newReachabilityInterval(31, 60),
				newReachabilityInterval(61, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{25, 25},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{1, 1},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 50),
				newReachabilityInterval(51, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{33, 33, 33},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 33),
				newReachabilityInterval(34, 66),
				newReachabilityInterval(67, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{10, 15, 25},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 10),
				newReachabilityInterval(11, 25),
				newReachabilityInterval(26, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 100),
			sizes:    []uint64{25, 15, 10},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 75),
				newReachabilityInterval(76, 90),
				newReachabilityInterval(91, 100),
			},
		},
		{
			interval: newReachabilityInterval(1, 10_000),
			sizes:    []uint64{10, 10, 20},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 20),
				newReachabilityInterval(21, 40),
				newReachabilityInterval(41, 10_000),
			},
		},
		{
			interval: newReachabilityInterval(1, 100_000),
			sizes:    []uint64{31_000, 31_000, 30_001},
			expectedIntervals: []*model.ReachabilityInterval{
				newReachabilityInterval(1, 35_000),
				newReachabilityInterval(35_001, 69_999),
				newReachabilityInterval(70_000, 100_000),
			},
		},
	}

	for i, test := range tests {
		intervals, err := intervalSplitWithExponentialBias(test.interval, test.sizes)
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
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	futureCoveringTreeNodeSet := model.FutureCoveringTreeNodeSet{
		helper.newNodeWithInterval(newReachabilityInterval(2, 3)),
		helper.newNodeWithInterval(newReachabilityInterval(4, 67)),
		helper.newNodeWithInterval(newReachabilityInterval(67, 77)),
		helper.newNodeWithInterval(newReachabilityInterval(657, 789)),
		helper.newNodeWithInterval(newReachabilityInterval(1000, 1000)),
		helper.newNodeWithInterval(newReachabilityInterval(1920, 1921)),
	}

	nodeWithFutureCoveringTreeNodeSet := helper.newNode()
	err := helper.stageFutureCoveringSet(nodeWithFutureCoveringTreeNodeSet, futureCoveringTreeNodeSet)
	if err != nil {
		t.Fatalf("stageFutureCoveringSet: %s", err)
	}

	tests := []struct {
		treeNode       *externalapi.DomainHash
		expectedResult bool
	}{
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(1, 1)),
			expectedResult: false,
		},
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(5, 7)),
			expectedResult: true,
		},
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(67, 76)),
			expectedResult: true,
		},
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(78, 100)),
			expectedResult: false,
		},
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(1980, 2000)),
			expectedResult: false,
		},
		{
			treeNode:       helper.newNodeWithInterval(newReachabilityInterval(1920, 1920)),
			expectedResult: true,
		},
	}

	for i, test := range tests {
		result, err := helper.futureCoveringSetHasAncestorOf(nodeWithFutureCoveringTreeNodeSet, test.treeNode)
		if err != nil {
			t.Fatalf("futureCoveringSetHasAncestorOf: %s", err)
		}

		if result != test.expectedResult {
			t.Errorf("TestHasAncestorOf: unexpected result in test #%d. Want: %t, got: %t",
				i, test.expectedResult, result)
		}
	}
}

func TestInsertToFutureCoveringSet(t *testing.T) {
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	nodeByIntervalMap := make(map[model.ReachabilityInterval]*externalapi.DomainHash)
	nodeByInterval := func(interval *model.ReachabilityInterval) *externalapi.DomainHash {
		if node, ok := nodeByIntervalMap[*interval]; ok {
			return node
		}

		nodeByIntervalMap[*interval] = helper.newNodeWithInterval(interval)
		return nodeByIntervalMap[*interval]
	}

	futureCoveringTreeNodeSet := model.FutureCoveringTreeNodeSet{
		nodeByInterval(newReachabilityInterval(1, 3)),
		nodeByInterval(newReachabilityInterval(4, 67)),
		nodeByInterval(newReachabilityInterval(67, 77)),
		nodeByInterval(newReachabilityInterval(657, 789)),
		nodeByInterval(newReachabilityInterval(1000, 1000)),
		nodeByInterval(newReachabilityInterval(1920, 1921)),
	}

	tests := []struct {
		toInsert       []*externalapi.DomainHash
		expectedResult model.FutureCoveringTreeNodeSet
	}{
		{
			toInsert: []*externalapi.DomainHash{
				nodeByInterval(newReachabilityInterval(5, 7)),
			},
			expectedResult: model.FutureCoveringTreeNodeSet{
				nodeByInterval(newReachabilityInterval(1, 3)),
				nodeByInterval(newReachabilityInterval(4, 67)),
				nodeByInterval(newReachabilityInterval(67, 77)),
				nodeByInterval(newReachabilityInterval(657, 789)),
				nodeByInterval(newReachabilityInterval(1000, 1000)),
				nodeByInterval(newReachabilityInterval(1920, 1921)),
			},
		},
		{
			toInsert: []*externalapi.DomainHash{
				nodeByInterval(newReachabilityInterval(65, 78)),
			},
			expectedResult: model.FutureCoveringTreeNodeSet{
				nodeByInterval(newReachabilityInterval(1, 3)),
				nodeByInterval(newReachabilityInterval(4, 67)),
				nodeByInterval(newReachabilityInterval(65, 78)),
				nodeByInterval(newReachabilityInterval(657, 789)),
				nodeByInterval(newReachabilityInterval(1000, 1000)),
				nodeByInterval(newReachabilityInterval(1920, 1921)),
			},
		},
		{
			toInsert: []*externalapi.DomainHash{
				nodeByInterval(newReachabilityInterval(88, 97)),
			},
			expectedResult: model.FutureCoveringTreeNodeSet{
				nodeByInterval(newReachabilityInterval(1, 3)),
				nodeByInterval(newReachabilityInterval(4, 67)),
				nodeByInterval(newReachabilityInterval(67, 77)),
				nodeByInterval(newReachabilityInterval(88, 97)),
				nodeByInterval(newReachabilityInterval(657, 789)),
				nodeByInterval(newReachabilityInterval(1000, 1000)),
				nodeByInterval(newReachabilityInterval(1920, 1921)),
			},
		},
		{
			toInsert: []*externalapi.DomainHash{
				nodeByInterval(newReachabilityInterval(88, 97)),
				nodeByInterval(newReachabilityInterval(3000, 3010)),
			},
			expectedResult: model.FutureCoveringTreeNodeSet{
				nodeByInterval(newReachabilityInterval(1, 3)),
				nodeByInterval(newReachabilityInterval(4, 67)),
				nodeByInterval(newReachabilityInterval(67, 77)),
				nodeByInterval(newReachabilityInterval(88, 97)),
				nodeByInterval(newReachabilityInterval(657, 789)),
				nodeByInterval(newReachabilityInterval(1000, 1000)),
				nodeByInterval(newReachabilityInterval(1920, 1921)),
				nodeByInterval(newReachabilityInterval(3000, 3010)),
			},
		},
	}

	for i, test := range tests {
		// Create a clone of treeNodes so that we have a clean start for every test
		futureCoveringTreeNodeSetClone := make(model.FutureCoveringTreeNodeSet, len(futureCoveringTreeNodeSet))
		copy(futureCoveringTreeNodeSetClone, futureCoveringTreeNodeSet)

		node := helper.newNode()
		err := helper.stageFutureCoveringSet(node, futureCoveringTreeNodeSetClone)
		if err != nil {
			t.Fatalf("stageFutureCoveringSet: %s", err)
		}

		for _, treeNode := range test.toInsert {
			err := helper.insertToFutureCoveringSet(node, treeNode)
			if err != nil {
				t.Fatalf("insertToFutureCoveringSet: %s", err)
			}
		}

		resultFutureCoveringTreeNodeSet, err := helper.futureCoveringSet(node)
		if err != nil {
			t.Fatalf("futureCoveringSet: %s", err)
		}
		if !reflect.DeepEqual(resultFutureCoveringTreeNodeSet, test.expectedResult) {
			t.Errorf("TestInsertToFutureCoveringSet: unexpected result in test #%d. Want: %s, got: %s",
				i, test.expectedResult, resultFutureCoveringTreeNodeSet)
		}
	}
}

func TestSplitFractionErrors(t *testing.T) {
	interval := newReachabilityInterval(100, 200)

	// Negative fraction
	_, _, err := intervalSplitFraction(interval, -0.5)
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
	_, _, err = intervalSplitFraction(interval, 1.5)
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
	_, _, err = intervalSplitFraction(emptyInterval, 0.5)
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
	_, err := intervalSplitExact(interval, sizes)
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
	_, err = intervalSplitExact(interval, sizes)
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
	_, err := intervalSplitWithExponentialBias(interval, sizes)
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
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	// Create a treeNode and give it size = 100
	treeNode := helper.newNodeWithInterval(newReachabilityInterval(0, 99))

	// Add a chain of 100 child treeNodes to treeNode
	var err error
	currentTreeNode := treeNode
	for i := 0; i < 100; i++ {
		childTreeNode := helper.newNode()
		err = helper.reachabilityManager.addChild(currentTreeNode, childTreeNode, treeNode)
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
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, b, reachabilityDataStore)

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		const subTreeSize = 70000
		// We set the interval of the root to subTreeSize*2 because
		// its first child gets half of the interval, so a reindex
		// from the root should happen after adding subTreeSize
		// nodes.
		root := helper.newNodeWithInterval(newReachabilityInterval(0, subTreeSize*2))

		currentTreeNode := root
		for i := 0; i < subTreeSize; i++ {
			childTreeNode := helper.newNode()
			helper.addChild(currentTreeNode, childTreeNode, root)

			currentTreeNode = childTreeNode
		}

		originalRemainingInterval := helper.remainingIntervalAfter(root).Clone()
		// After we added subTreeSize nodes, adding the next
		// node should lead to a reindex from root.
		fullReindexTriggeringNode := helper.newNode()
		b.StartTimer()
		err := helper.reachabilityManager.addChild(currentTreeNode, fullReindexTriggeringNode, root)
		b.StopTimer()
		if err != nil {
			b.Fatalf("addChild: %s", err)
		}

		if helper.remainingIntervalAfter(root).Equal(originalRemainingInterval) {
			b.Fatal("Expected a reindex from root, but it didn't happen")
		}
	}
}

func TestReachabilityTreeNodeString(t *testing.T) {
	reachabilityDataStore := newReachabilityDataStoreMock()
	manager := New(nil, nil, reachabilityDataStore).(*reachabilityManager)
	helper := newTestHelper(manager, t, reachabilityDataStore)

	treeNodeA := helper.newNodeWithInterval(newReachabilityInterval(100, 199))
	treeNodeB1 := helper.newNodeWithInterval(newReachabilityInterval(100, 150))
	treeNodeB2 := helper.newNodeWithInterval(newReachabilityInterval(150, 199))
	treeNodeC := helper.newNodeWithInterval(newReachabilityInterval(100, 149))

	err := helper.stageAddChild(treeNodeA, treeNodeB1)
	if err != nil {
		t.Fatalf("stageAddChild: %s", err)
	}

	err = helper.stageAddChild(treeNodeA, treeNodeB2)
	if err != nil {
		t.Fatalf("stageAddChild: %s", err)
	}

	err = helper.stageAddChild(treeNodeB2, treeNodeC)
	if err != nil {
		t.Fatalf("stageAddChild: %s", err)
	}

	str, err := manager.String(treeNodeA)
	if err != nil {
		t.Fatalf("String: %s", err)
	}
	expectedStr := "[100,149]\n[100,150][150,199]\n[100,199]"
	if str != expectedStr {
		t.Fatalf("TestReachabilityTreeNodeString: unexpected "+
			"string. Want: %s, got: %s", expectedStr, str)
	}
}
