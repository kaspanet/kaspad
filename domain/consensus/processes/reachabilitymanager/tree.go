package reachabilitymanager

import (
	"math"
	"strings"
	"time"

	"github.com/kaspanet/kaspad/domain/consensus/utils/reachabilitydata"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

	"github.com/pkg/errors"
)

var (
	// defaultReindexWindow is the default target window size for reachability
	// reindexes. Note that this is not a constant for testing purposes.
	defaultReindexWindow uint64 = 200

	// defaultReindexSlack is default the slack interval given to reachability
	// tree nodes not in the selected parent chain. Note that this is not
	// a constant for testing purposes.
	defaultReindexSlack uint64 = 1 << 12
)

type reindexContext struct {
	manager *reachabilityManager
	subTreeSizesCache map[externalapi.DomainHash]uint64
}

func newReindexContext(rt *reachabilityManager) reindexContext {
	return reindexContext{
		manager: rt,
		subTreeSizesCache: make(map[externalapi.DomainHash]uint64),
	}
}

func newReachabilityTreeData() model.ReachabilityData {
	// Please see the comment above model.ReachabilityTreeNode to understand why
	// we use these initial values.
	interval := newReachabilityInterval(1, math.MaxUint64-1)
	data := reachabilitydata.EmptyReachabilityData()
	data.SetInterval(interval)

	return data
}

func (rt *reachabilityManager) intervalRangeForChildAllocation(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	interval, err := rt.interval(node)
	if err != nil {
		return nil, err
	}

	// We subtract 1 from the end of the range to prevent the node from allocating
	// the entire interval to its child, so its interval would *strictly* contain the interval of its child.
	return newReachabilityInterval(interval.Start, interval.End-1), nil
}

func (rt *reachabilityManager) remainingIntervalBefore(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	childRange, err := rt.intervalRangeForChildAllocation(node)
	if err != nil {
		return nil, err
	}

	children, err := rt.children(node)
	if err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return childRange, nil
	}

	firstChildInterval, err := rt.interval(children[0])
	if err != nil {
		return nil, err
	}

	return newReachabilityInterval(childRange.Start, firstChildInterval.Start-1), nil
}

func (rt *reachabilityManager) remainingIntervalAfter(node *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	childRange, err := rt.intervalRangeForChildAllocation(node)
	if err != nil {
		return nil, err
	}

	children, err := rt.children(node)
	if err != nil {
		return nil, err
	}

	if len(children) == 0 {
		return childRange, nil
	}

	lastChildInterval, err := rt.interval(children[len(children)-1])
	if err != nil {
		return nil, err
	}

	return newReachabilityInterval(lastChildInterval.End+1, childRange.End), nil
}

func (rt *reachabilityManager) remainingSlackBefore(node *externalapi.DomainHash) (uint64, error) {
	interval, err := rt.remainingIntervalBefore(node)
	if err != nil {
		return 0, err
	}
	return intervalSize(interval), nil
}

func (rt *reachabilityManager) remainingSlackAfter(node *externalapi.DomainHash) (uint64, error) {
	interval, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return 0, err
	}
	return intervalSize(interval), nil
}

func (rt *reachabilityManager) hasSlackIntervalBefore(node *externalapi.DomainHash) (bool, error) {
	interval, err := rt.remainingIntervalBefore(node)
	if err != nil {
		return false, err
	}

	return intervalSize(interval) > 0, nil
}

func (rt *reachabilityManager) hasSlackIntervalAfter(node *externalapi.DomainHash) (bool, error) {
	interval, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return false, err
	}

	return intervalSize(interval) > 0, nil
}

// addChild adds child to this tree node. If this node has no
// remaining interval to allocate, a reindexing is triggered.
// This method returns a list of model.ReachabilityTreeNodes modified
// by it.
func (rt *reachabilityManager) addChild(node, child, root *externalapi.DomainHash) error {
	remaining, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return err
	}

	// Set the parent-child relationship
	err = rt.stageAddChild(node, child)
	if err != nil {
		return err
	}

	err = rt.stageParent(child, node)
	if err != nil {
		return err
	}

	// No allocation space left -- reindex
	if intervalSize(remaining) == 0 {

		// Temporarily set the child's interval to be empty, at
		// the start of node's remaining interval. This is done
		// so that child-of-node checks (e.g.
		// FindNextDescendantChainBlock) will not fail for node.
		err = rt.stageInterval(child, newReachabilityInterval(remaining.Start, remaining.Start-1))
		if err != nil {
			return err
		}

		rc := newReindexContext(rt)

		reindexStartTime := time.Now()
		err := rc.reindexIntervals(node, root)
		if err != nil {
			return err
		}

		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. Took %dms.",
			node, reindexTimeElapsed.Milliseconds())

		return nil
	}

	// Allocate from the remaining space
	allocated, _, err := intervalSplitInHalf(remaining)
	if err != nil {
		return err
	}

	return rt.stageInterval(child, allocated)
}

// reindexIntervals traverses the reachability subtree that's
// defined by this node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter. It does this by traversing down the reachability
// tree until it finds a node with a subtree size that's greater than
// its interval size. See propagateInterval for further details.
func (rc *reindexContext) reindexIntervals(node, root *externalapi.DomainHash) error {

	if res, _ := rc.manager.isStrictAncestorOf(node, root); res {
		// In this case we avoid reindexing the entire subtree and
		// use slacks along the chain up from parent to reindex root
		return rc.reindexIntervalsEarlierThanRoot(node, root, node, 1)
	}

	current := node

	// Find the first ancestor that has sufficient interval space
	for {
		currentInterval, err := rc.manager.interval(current)
		if err != nil {
			return err
		}

		currentIntervalSize := intervalSize(currentInterval)

		err = rc.countSubtrees(current)
		if err != nil {
			return err
		}

		currentSubtreeSize := rc.subTreeSizesCache[*current]

		if currentIntervalSize >= currentSubtreeSize {
			break
		}

		parent, err := rc.manager.parent(current)
		if err != nil {
			return err
		}

		if parent == nil {
			// If we ended up here it means that there are more
			// than 2^64 blocks, which shouldn't ever happen.
			return errors.Errorf("missing tree " +
				"parent during reindexing. Theoretically, this " +
				"should only ever happen if there are more " +
				"than 2^64 blocks in the DAG.")
		}

		if current.Equal(root) {
			return errors.Errorf("unexpected behavior: reindex root %s is out of capacity" +
				"during reindexing. Theoretically, this " +
				"should only ever happen if there are more " +
				"than 2^64 blocks in the DAG.", root.String())
		}

		if res, _ := rc.manager.isStrictAncestorOf(parent, root); res {
			// In this case we avoid reindexing the entire subtree and
			// use slacks along the chain up from parent to reindex root
			return rc.reindexIntervalsEarlierThanRoot(node, root, parent, currentSubtreeSize)
		}

		current = parent
	}

	// Propagate the interval to the subtree
	return rc.propagateInterval(current)
}

// countSubtrees counts the size of each subtree under this node,
// and populates the provided subTreeSizeMap with the results.
// It is equivalent to the following recursive implementation:
//
// func (rt *reachabilityManager) countSubtrees(node *model.ReachabilityTreeNode) uint64 {
//     subtreeSize := uint64(0)
//     for _, child := range node.children {
//         subtreeSize += child.countSubtrees()
//     }
//     return subtreeSize + 1
// }
//
// However, we are expecting (linearly) deep trees, and so a
// recursive stack-based approach is inefficient and will hit
// recursion limits. Instead, the same logic was implemented
// using a (queue-based) BFS method. At a high level, the
// algorithm uses BFS for reaching all leaves and pushes
// intermediate updates from leaves via parent chains until all
// size information is gathered at the root of the operation
// (i.e. at node).
func (rc *reindexContext) countSubtrees(node *externalapi.DomainHash) error {

	if _, ok := rc.subTreeSizesCache[*node]; ok {
		return nil
	}

	queue := []*externalapi.DomainHash{node}
	calculatedChildrenCount := make(map[externalapi.DomainHash]uint64)
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		children, err := rc.manager.children(current)
		if err != nil {
			return err
		}

		if len(children) == 0 {
			// We reached a leaf
			rc.subTreeSizesCache[*current] = 1
		} else if _, ok := rc.subTreeSizesCache[*current]; !ok {
			// We haven't yet calculated the subtree size of
			// the current node. Add all its children to the
			// queue
			queue = append(queue, children...)
			continue
		}

		// We reached a leaf or a pre-calculated subtree.
		// Push information up
		for !current.Equal(node) {
			current, err = rc.manager.parent(current)
			if err != nil {
				return err
			}

			// If the current is now nil, it means that the previous
			// `current` was the genesis block -- the only block that
			// does not have parents
			if current == nil {
				break
			}

			calculatedChildrenCount[*current]++

			children, err := rc.manager.children(current)
			if err != nil {
				return err
			}

			if calculatedChildrenCount[*current] != uint64(len(children)) {
				// Not all subtrees of the current node are ready
				break
			}
			// All children of `current` have calculated their subtree size.
			// Sum them all together and add 1 to get the sub tree size of
			// `current`.
			childSubtreeSizeSum := uint64(0)
			for _, child := range children {
				childSubtreeSizeSum += rc.subTreeSizesCache[*child]
			}
			rc.subTreeSizesCache[*current] = childSubtreeSizeSum + 1
		}
	}

	return nil
}

// propagateInterval propagates the new interval using a BFS traversal.
// Subtree intervals are recursively allocated according to subtree sizes and
// the allocation rule in splitWithExponentialBias. This method returns
// a list of model.ReachabilityTreeNodes modified by it.
func (rc *reindexContext) propagateInterval(node *externalapi.DomainHash) error {

	// Make sure subtrees are counted before propagating
	err := rc.countSubtrees(node)
	if err != nil {
		return err
	}

	queue := []*externalapi.DomainHash{node}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		children, err := rc.manager.children(current)
		if err != nil {
			return err
		}

		if len(children) > 0 {
			sizes := make([]uint64, len(children))
			for i, child := range children {
				sizes[i] = rc.subTreeSizesCache[*child]
			}

			interval, err := rc.manager.intervalRangeForChildAllocation(current)
			if err != nil {
				return err
			}

			intervals, err := intervalSplitWithExponentialBias(interval, sizes)
			if err != nil {
				return err
			}

			for i, child := range children {
				childInterval := intervals[i]
				err = rc.manager.stageInterval(child, childInterval)
				if err != nil {
					return err
				}
				queue = append(queue, child)
			}
		}
	}
	return nil
}

func (rc *reindexContext) reindexIntervalsEarlierThanRoot(
	node, root, common *externalapi.DomainHash, requiredAllocation uint64) error {

	// The chosen child is:
	// a. A reachability tree child of `common`
	// b. A reachability tree ancestor of `root`
	chosen, err := rc.manager.FindNextDescendantChainBlock(root, common)
	if err != nil {
		return err
	}

	nodeInterval, err := rc.manager.interval(node)
	if err != nil {
		return err
	}

	chosenInterval, err := rc.manager.interval(chosen)
	if err != nil {
		return err
	}

	if nodeInterval.End < chosenInterval.Start {
		// node is in the subtree before the chosen child
		return rc.reclaimIntervalBefore(node, common, chosen, root, requiredAllocation)
	}

	// node is either:
	// * in the subtree after the chosen child
	// * the common ancestor
	// In both cases we reclaim from the "after" subtree. In the
	// latter case this is arbitrary
	return rc.reclaimIntervalAfter(node, common, chosen, root, requiredAllocation)
}

func (rc *reindexContext) reclaimIntervalBefore(
	node, common, chosen, root *externalapi.DomainHash, requiredAllocation uint64) error {

	var slackSum uint64 = 0
	var pathLen uint64 = 0
	var pathSlackAlloc uint64 = 0

	current := chosen
	for {
		if current.Equal(root) {
			previousInterval, err := rc.manager.interval(current)
			if err != nil {
				return err
			}

			offset := requiredAllocation + rc.manager.reindexSlack * pathLen - slackSum
			err = rc.manager.stageInterval(current, previousInterval.IncreaseStart(offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(current)
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsBefore(node, common, current, offset)
			if err != nil {
				return err
			}

			pathSlackAlloc = rc.manager.reindexSlack
			break
		}

		slackBeforeCurrent, err := rc.manager.remainingSlackBefore(current)
		if err != nil {
			return err
		}
		slackSum += slackBeforeCurrent

		if slackSum >= requiredAllocation {
			previousInterval, err := rc.manager.interval(current)
			if err != nil {
				return err
			}

			offset := slackBeforeCurrent - (slackSum - requiredAllocation)

			err = rc.manager.stageInterval(current, previousInterval.IncreaseStart(offset))
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsBefore(node, common, current, offset)
			if err != nil {
				return err
			}

			break
		}

		current, err = rc.manager.FindNextDescendantChainBlock(root, current)
		if err != nil {
			return err
		}

		pathLen++
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree before the
	// current node with an interval that is smaller.
	// This is to make room for the new node.
	for {
		current, err := rc.manager.parent(current)
		if err != nil {
			return err
		}

		if current.Equal(common) {
			break
		}

		originalInterval, err := rc.manager.interval(current)
		if err != nil {
			return err
		}

		slackBeforeCurrent, err := rc.manager.remainingSlackBefore(current)
		if err != nil {
			return err
		}

		offset := slackBeforeCurrent - pathSlackAlloc
		err = rc.manager.stageInterval(current, originalInterval.IncreaseStart(offset))
		if err != nil {
			return err
		}

		err = rc.offsetSiblingsBefore(node, common, current, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) offsetSiblingsBefore(node, common, chosen *externalapi.DomainHash, offset uint64) error {
	parent, err := rc.manager.parent(chosen)
	if err != nil {
		return err
	}

	siblingsBefore, _, err := rc.manager.splitChildren(parent, chosen)
	if err != nil {
		return err
	}

	if parent.Equal(common) {
		previousInterval, err := rc.manager.interval(node)
		if err != nil {
			return err
		}

		err = rc.manager.stageInterval(node, previousInterval.IncreaseEnd(offset))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(node)
		if err != nil {
			return err
		}

		indexOfNode := -1
		for i, sibling := range siblingsBefore {
			if sibling.Equal(node) {
				indexOfNode = i
				break
			}
		}
		if indexOfNode < 0 {
			err = errors.Errorf("node %s is expected to be child of coomon %s", node, common)
			return err
		}
		siblingsBefore = siblingsBefore[indexOfNode+1:]
	}

	for _, sibling := range siblingsBefore {
		previousInterval, err := rc.manager.interval(sibling)
		if err != nil {
			return err
		}

		err = rc.manager.stageInterval(sibling, previousInterval.Increase(offset))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(sibling)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) reclaimIntervalAfter(
	node, common, chosen, root *externalapi.DomainHash, requiredAllocation uint64) error {

	var slackSum uint64 = 0
	var pathLen uint64 = 0
	var pathSlackAlloc uint64 = 0

	current := chosen
	for {
		if current.Equal(root) {
			previousInterval, err := rc.manager.interval(current)
			if err != nil {
				return err
			}

			offset := requiredAllocation + rc.manager.reindexSlack * pathLen - slackSum
			err = rc.manager.stageInterval(current, previousInterval.DecreaseEnd(offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(current)
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsAfter(node, common, current, offset)
			if err != nil {
				return err
			}

			pathSlackAlloc = rc.manager.reindexSlack
			break
		}

		slackAfterCurrent, err := rc.manager.remainingSlackAfter(current)
		if err != nil {
			return err
		}
		slackSum += slackAfterCurrent

		if slackSum >= requiredAllocation {
			previousInterval, err := rc.manager.interval(current)
			if err != nil {
				return err
			}

			offset := slackAfterCurrent - (slackSum - requiredAllocation)

			err = rc.manager.stageInterval(current, previousInterval.DecreaseEnd(offset))
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsAfter(node, common, current, offset)
			if err != nil {
				return err
			}

			break
		}

		current, err = rc.manager.FindNextDescendantChainBlock(root, current)
		if err != nil {
			return err
		}

		pathLen++
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree after the
	// current node with an interval that is smaller.
	// This is to make room for the new node.

	for {
		current, err := rc.manager.parent(current)
		if err != nil {
			return err
		}

		if current.Equal(common) {
			break
		}

		originalInterval, err := rc.manager.interval(current)
		if err != nil {
			return err
		}

		slackAfterCurrent, err := rc.manager.remainingSlackAfter(current)
		if err != nil {
			return err
		}

		offset := slackAfterCurrent - pathSlackAlloc
		err = rc.manager.stageInterval(current, originalInterval.DecreaseEnd(offset))
		if err != nil {
			return err
		}

		err = rc.offsetSiblingsAfter(node, common, current, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) offsetSiblingsAfter(node, common, chosen *externalapi.DomainHash, offset uint64) error {
	parent, err := rc.manager.parent(chosen)
	if err != nil {
		return err
	}

	_, siblingsAfter, err := rc.manager.splitChildren(parent, chosen)
	if err != nil {
		return err
	}

	for i, sibling := range siblingsAfter {
		if sibling.Equal(node) || (node.Equal(common) && i == len(siblingsAfter) - 1) {
			previousInterval, err := rc.manager.interval(sibling)
			if err != nil {
				return err
			}

			err = rc.manager.stageInterval(sibling, previousInterval.DecreaseStart(offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(sibling)
			if err != nil {
				return err
			}

			break
		}

		previousInterval, err := rc.manager.interval(sibling)
		if err != nil {
			return err
		}

		err = rc.manager.stageInterval(sibling, previousInterval.Decrease(offset))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(sibling)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsReachabilityTreeAncestorOf checks if this node is a reachability tree ancestor
// of the other node. Note that we use the graph theory convention
// here which defines that node is also an ancestor of itself.
func (rt *reachabilityManager) IsReachabilityTreeAncestorOf(node, other *externalapi.DomainHash) (bool, error) {
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return false, err
	}

	otherInterval, err := rt.interval(other)
	if err != nil {
		return false, err
	}

	return intervalContains(nodeInterval, otherInterval), nil
}

func (rt *reachabilityManager) isStrictAncestorOf(node, other *externalapi.DomainHash) (bool, error) {
	if node.Equal(other) {
		return false, nil
	}
	return rt.IsReachabilityTreeAncestorOf(node, other)
}

// findCommonAncestor finds the most recent reachability
// tree ancestor common to both node and the given reindex root. Note
// that we assume that almost always the chain between the reindex root
// and the common ancestor is longer than the chain between node and the
// common ancestor.
func (rt *reachabilityManager) findCommonAncestor(node, root *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	currentThis := node
	for {
		isAncestorOf, err := rt.IsReachabilityTreeAncestorOf(currentThis, root)
		if err != nil {
			return nil, err
		}

		if isAncestorOf {
			return currentThis, nil
		}

		currentThis, err = rt.parent(currentThis)
		if err != nil {
			return nil, err
		}
	}
}

// String returns a string representation of a reachability tree node
// and its children.
func (rt *reachabilityManager) String(node *externalapi.DomainHash) (string, error) {
	queue := []*externalapi.DomainHash{node}
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return "", err
	}

	lines := []string{nodeInterval.String()}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		currentChildren, err := rt.children(current)
		if err != nil {
			return "", err
		}

		if len(currentChildren) == 0 {
			continue
		}

		line := ""
		for _, child := range currentChildren {
			childInterval, err := rt.interval(child)
			if err != nil {
				return "", err
			}

			line += childInterval.String()
			queue = append(queue, child)
		}
		lines = append([]string{line}, lines...)
	}
	return strings.Join(lines, "\n"), nil
}

func (rt *reachabilityManager) updateReindexRoot(newTreeNode *externalapi.DomainHash) error {

	nextReindexRoot, err := rt.reindexRoot()
	if err != nil {
		return err
	}

	for {
		candidateReindexRoot, found, err := rt.maybeMoveReindexRoot(nextReindexRoot, newTreeNode)
		if err != nil {
			return err
		}
		if !found {
			break
		}
		nextReindexRoot = candidateReindexRoot
	}

	rt.stageReindexRoot(nextReindexRoot)
	return nil
}

func (rt *reachabilityManager) maybeMoveReindexRoot(reindexRoot, newTreeNode *externalapi.DomainHash) (
	newReindexRoot *externalapi.DomainHash, found bool, err error) {

	isAncestorOf, err := rt.IsReachabilityTreeAncestorOf(reindexRoot, newTreeNode)
	if err != nil {
		return nil, false, err
	}
	if !isAncestorOf {
		commonAncestor, err := rt.findCommonAncestor(newTreeNode, reindexRoot)
		if err != nil {
			return nil, false, err
		}

		return commonAncestor, true, nil
	}

	reindexRootChosenChild, err := rt.FindNextDescendantChainBlock(newTreeNode, reindexRoot)
	if err != nil {
		return nil, false, err
	}

	newTreeNodeGHOSTDAGData, err := rt.ghostdagDataStore.Get(rt.databaseContext, newTreeNode)
	if err != nil {
		return nil, false, err
	}

	reindexRootChosenChildGHOSTDAGData, err := rt.ghostdagDataStore.Get(rt.databaseContext, reindexRootChosenChild)
	if err != nil {
		return nil, false, err
	}

	if newTreeNodeGHOSTDAGData.BlueScore()-reindexRootChosenChildGHOSTDAGData.BlueScore() < rt.reindexWindow {
		return nil, false, nil
	}

	err = rt.concentrateInterval(reindexRoot, reindexRootChosenChild)
	if err != nil {
		return nil, false, err
	}

	return reindexRootChosenChild, true, nil
}

// NextDescendantChainBlock finds the reachability tree child
// of 'parent' which is the ancestor of 'descendant'.
func (rt *reachabilityManager) FindNextDescendantChainBlock(descendant, parent *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	children, err := rt.children(parent)
	if err != nil {
		return nil, err
	}

	ancestor, ok := rt.findAncestorOfNode(children, descendant)
	if !ok {
		return nil, errors.Errorf("parent is not an ancestor of descendant")
	}

	return ancestor, nil
}

func (rt *reachabilityManager) concentrateInterval(root, chosen *externalapi.DomainHash) error {
	rc := newReindexContext(rt)

	reindexRootChildNodesBeforeChosen, reindexRootChildNodesAfterChosen, err :=
		rt.splitChildren(root, chosen)
	if err != nil {
		return err
	}

	reindexRootChildNodesBeforeChosenSizesSum, err :=
		rc.tightenIntervalsBefore(root, reindexRootChildNodesBeforeChosen)
	if err != nil {
		return err
	}

	reindexRootChildNodesAfterChosenSizesSum, err :=
		rc.tightenIntervalsAfter(root, reindexRootChildNodesAfterChosen)
	if err != nil {
		return err
	}

	err = rc.expandIntervalToChosen(root, chosen,
		reindexRootChildNodesBeforeChosenSizesSum, reindexRootChildNodesAfterChosenSizesSum)
	if err != nil {
		return err
	}

	return nil
}

// splitChildren splits `node` into two slices: the nodes that are before
// `child` and the nodes that are after.
func (rt *reachabilityManager) splitChildren(node, pivot *externalapi.DomainHash) (
	nodesBeforePivot, nodesAfterPivot []*externalapi.DomainHash, err error) {

	children, err := rt.children(node)
	if err != nil {
		return nil, nil, err
	}

	for i, child := range children {
		if child.Equal(pivot) {
			return children[:i], children[i+1:], nil
		}
	}
	return nil, nil, errors.Errorf("pivot not a pivot of node")
}

func (rc *reindexContext) tightenIntervalsBefore(
	root *externalapi.DomainHash,
	nodesBeforeChosen []*externalapi.DomainHash) (sizesSum uint64,
	err error) {

	reindexRootChildNodesBeforeChosenSizes, sizesSum :=
		rc.calcReachabilityTreeNodeSizes(nodesBeforeChosen)

	reindexRootInterval, err := rc.manager.interval(root)
	if err != nil {
		return 0, err
	}

	intervalBeforeReindexRootStart := newReachabilityInterval(
		reindexRootInterval.Start+rc.manager.reindexSlack,
		reindexRootInterval.Start+rc.manager.reindexSlack+sizesSum-1,
	)

	err = rc.propagateChildIntervals(
		intervalBeforeReindexRootStart, nodesBeforeChosen, reindexRootChildNodesBeforeChosenSizes)
	if err != nil {
		return 0, err
	}
	return sizesSum, nil
}

func (rc *reindexContext) tightenIntervalsAfter(
	root *externalapi.DomainHash, nodesAfterChosen []*externalapi.DomainHash) (sizesSum uint64, err error) {

	reindexRootChildNodesAfterChosenSizes, sizesSum :=
		rc.calcReachabilityTreeNodeSizes(nodesAfterChosen)

	reindexRootInterval, err := rc.manager.interval(root)
	if err != nil {
		return 0, err
	}

	intervalAfterReindexRootEnd := newReachabilityInterval(
		reindexRootInterval.End-rc.manager.reindexSlack-sizesSum,
		reindexRootInterval.End-rc.manager.reindexSlack-1,
	)

	err = rc.propagateChildIntervals(
		intervalAfterReindexRootEnd, nodesAfterChosen, reindexRootChildNodesAfterChosenSizes)
	if err != nil {
		return 0, err
	}
	return sizesSum, nil
}

func (rc *reindexContext) expandIntervalToChosen(
	root, chosen *externalapi.DomainHash,
	sizesSumBefore, sizesSumAfter uint64) error {

	rootInterval, err := rc.manager.interval(root)
	if err != nil {
		return err
	}

	newChosenInterval := newReachabilityInterval(
		rootInterval.Start+sizesSumBefore+rc.manager.reindexSlack,
		rootInterval.End-sizesSumAfter-rc.manager.reindexSlack-1,
	)

	currentChosenInterval, err := rc.manager.interval(chosen)
	if err != nil {
		return err
	}

	if !intervalContains(newChosenInterval, currentChosenInterval) {
		// New interval doesn't contain the previous one, propagation is required

		// We assign slack on both sides as an optimization. Were we to
		// assign a tight interval, the next time the reindex root moves we
		// would need to propagate intervals again. That is to say, when we
		// do allocate slack, next time
		// expandIntervalToChosen is called (next time the
		// reindex root moves), newChosenInterval is likely to
		// contain chosen.Interval.
		err := rc.manager.stageInterval(chosen, newReachabilityInterval(
			newChosenInterval.Start+rc.manager.reindexSlack,
			newChosenInterval.End-rc.manager.reindexSlack,
		))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(chosen)
		if err != nil {
			return err
		}
	}

	err = rc.manager.stageInterval(chosen, newChosenInterval)
	if err != nil {
		return err
	}
	return nil
}

func (rc *reindexContext) calcReachabilityTreeNodeSizes(treeNodes []*externalapi.DomainHash) (
	sizes []uint64, sum uint64) {

	sizes = make([]uint64, len(treeNodes))
	sum = 0
	for i, node := range treeNodes {
		err := rc.countSubtrees(node)
		if err != nil {
			return nil, 0
		}

		subtreeSize := rc.subTreeSizesCache[*node]
		sizes[i] = subtreeSize
		sum += subtreeSize
	}
	return sizes, sum
}

func (rc *reindexContext) propagateChildIntervals(
	interval *model.ReachabilityInterval, childNodes []*externalapi.DomainHash, sizes []uint64) error {

	childIntervalSizes, err := intervalSplitExact(interval, sizes)
	if err != nil {
		return err
	}

	for i, child := range childNodes {
		childInterval := childIntervalSizes[i]
		err := rc.manager.stageInterval(child, childInterval)
		if err != nil {
			return err
		}

		err = rc.propagateInterval(child)
		if err != nil {
			return err
		}
	}

	return nil
}
