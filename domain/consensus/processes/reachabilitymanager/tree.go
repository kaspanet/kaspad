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

	// slackReachabilityIntervalForReclaiming is the slack interval to
	// reclaim during reachability reindexes earlier than the reindex root.
	// See reclaimIntervalBeforeChosenChild for further details. Note that
	// this is not a constant for testing purposes.
	slackReachabilityIntervalForReclaiming uint64 = 1
)

// exponentialFractions returns a fraction of each size in sizes
// as follows:
//   fraction[i] = 2^size[i] / sum_j(2^size[j])
// In the code below the above equation is divided by 2^max(size)
// to avoid exploding numbers. Note that in 1 / 2^(max(size)-size[i])
// we divide 1 by potentially a very large number, which will
// result in loss of float precision. This is not a problem - all
// numbers close to 0 bear effectively the same weight.
func exponentialFractions(sizes []uint64) []float64 {
	maxSize := uint64(0)
	for _, size := range sizes {
		if size > maxSize {
			maxSize = size
		}
	}
	fractions := make([]float64, len(sizes))
	for i, size := range sizes {
		fractions[i] = 1 / math.Pow(2, float64(maxSize-size))
	}
	fractionsSum := float64(0)
	for _, fraction := range fractions {
		fractionsSum += fraction
	}
	for i, fraction := range fractions {
		fractions[i] = fraction / fractionsSum
	}
	return fractions
}

func newReachabilityTreeData() model.ReachabilityData {
	// Please see the comment above model.ReachabilityTreeNode to understand why
	// we use these initial values.
	interval := newReachabilityInterval(1, math.MaxUint64-1)
	data := reachabilitydata.EmptyReachabilityData()
	data.SetInterval(interval)

	return data
}

func (rt *reachabilityManager) intervalRangeForChildAllocation(hash *externalapi.DomainHash) (*model.ReachabilityInterval, error) {
	interval, err := rt.interval(hash)
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
func (rt *reachabilityManager) addChild(node, child, reindexRoot *externalapi.DomainHash) error {
	remaining, err := rt.remainingIntervalAfter(node)
	if err != nil {
		return err
	}

	// Set the parent-child relationship
	err = rt.addChildAndStage(node, child)
	if err != nil {
		return err
	}

	err = rt.stageParent(child, node)
	if err != nil {
		return err
	}

	// Temporarily set the child's interval to be empty, at
	// the start of node's remaining interval. This is done
	// so that child-of-node checks (e.g.
	// FindAncestorOfThisAmongChildrenOfOther) will not fail for node.
	err = rt.stageInterval(child, newReachabilityInterval(remaining.Start, remaining.Start-1))
	if err != nil {
		return err
	}

	// Handle node not being a descendant of the reindex root.
	// Note that we check node here instead of child because
	// at this point we don't yet know child's interval.
	isReindexRootAncestorOfNode, err := rt.IsReachabilityTreeAncestorOf(reindexRoot, node)
	if err != nil {
		return err
	}

	if !isReindexRootAncestorOfNode {
		reindexStartTime := time.Now()
		err := rt.reindexIntervalsEarlierThanReindexRoot(node, reindexRoot)
		if err != nil {
			return err
		}
		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. This block is not a child of the current "+
			"reindex root %s. Took %dms.",
			node, reindexRoot, reindexTimeElapsed.Milliseconds())
		return nil
	}

	// No allocation space left -- reindex
	if intervalSize(remaining) == 0 {
		reindexStartTime := time.Now()
		err := rt.reindexIntervals(node)
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
// tree until it finds a node with a subreeSize that's greater than
// its interval size. See propagateInterval for further details.
// This method returns a list of model.ReachabilityTreeNodes modified by it.
func (rt *reachabilityManager) reindexIntervals(node *externalapi.DomainHash) error {
	current := node

	// Initial interval and subtree sizes
	currentInterval, err := rt.interval(node)
	if err != nil {
		return err
	}

	size := intervalSize(currentInterval)
	subTreeSizeMap := make(map[externalapi.DomainHash]uint64)
	err = rt.countSubtrees(current, subTreeSizeMap)
	if err != nil {
		return err
	}

	currentSubtreeSize := subTreeSizeMap[*current]

	// Find the first ancestor that has sufficient interval space
	for size < currentSubtreeSize {
		currentParent, err := rt.parent(current)
		if err != nil {
			return err
		}

		if currentParent == nil {
			// If we ended up here it means that there are more
			// than 2^64 blocks, which shouldn't ever happen.
			return errors.Errorf("missing tree " +
				"parent during reindexing. Theoretically, this " +
				"should only ever happen if there are more " +
				"than 2^64 blocks in the DAG.")
		}
		current = currentParent
		currentInterval, err := rt.interval(current)
		if err != nil {
			return err
		}

		size = intervalSize(currentInterval)
		err = rt.countSubtrees(current, subTreeSizeMap)
		if err != nil {
			return err
		}

		currentSubtreeSize = subTreeSizeMap[*current]
	}

	// Propagate the interval to the subtree
	return rt.propagateInterval(current, subTreeSizeMap)
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
func (rt *reachabilityManager) countSubtrees(node *externalapi.DomainHash, subTreeSizeMap map[externalapi.DomainHash]uint64) error {
	queue := []*externalapi.DomainHash{node}
	calculatedChildrenCount := make(map[externalapi.DomainHash]uint64)
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		currentChildren, err := rt.children(current)
		if err != nil {
			return err
		}

		if len(currentChildren) == 0 {
			// We reached a leaf
			subTreeSizeMap[*current] = 1
		} else if _, ok := subTreeSizeMap[*current]; !ok {
			// We haven't yet calculated the subtree size of
			// the current node. Add all its children to the
			// queue
			queue = append(queue, currentChildren...)
			continue
		}

		// We reached a leaf or a pre-calculated subtree.
		// Push information up
		for !current.Equal(node) {
			current, err = rt.parent(current)
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

			currentChildren, err := rt.children(current)
			if err != nil {
				return err
			}

			if calculatedChildrenCount[*current] != uint64(len(currentChildren)) {
				// Not all subtrees of the current node are ready
				break
			}
			// All children of `current` have calculated their subtree size.
			// Sum them all together and add 1 to get the sub tree size of
			// `current`.
			childSubtreeSizeSum := uint64(0)
			for _, child := range currentChildren {
				childSubtreeSizeSum += subTreeSizeMap[*child]
			}
			subTreeSizeMap[*current] = childSubtreeSizeSum + 1
		}
	}

	return nil
}

// propagateInterval propagates the new interval using a BFS traversal.
// Subtree intervals are recursively allocated according to subtree sizes and
// the allocation rule in splitWithExponentialBias. This method returns
// a list of model.ReachabilityTreeNodes modified by it.
func (rt *reachabilityManager) propagateInterval(node *externalapi.DomainHash, subTreeSizeMap map[externalapi.DomainHash]uint64) error {

	queue := []*externalapi.DomainHash{node}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		currentChildren, err := rt.children(current)
		if err != nil {
			return err
		}

		if len(currentChildren) > 0 {
			sizes := make([]uint64, len(currentChildren))
			for i, child := range currentChildren {
				sizes[i] = subTreeSizeMap[*child]
			}

			interval, err := rt.intervalRangeForChildAllocation(current)
			if err != nil {
				return err
			}

			intervals, err := intervalSplitWithExponentialBias(interval, sizes)
			if err != nil {
				return err
			}
			for i, child := range currentChildren {
				childInterval := intervals[i]
				err = rt.stageInterval(child, childInterval)
				if err != nil {
					return err
				}
				queue = append(queue, child)
			}
		}
	}
	return nil
}

func (rt *reachabilityManager) reindexIntervalsEarlierThanReindexRoot(node,
	reindexRoot *externalapi.DomainHash) error {

	// Find the common ancestor for both node and the reindex root
	commonAncestor, err := rt.findCommonAncestorWithReindexRoot(node, reindexRoot)
	if err != nil {
		return err
	}

	// The chosen child is:
	// a. A reachability tree child of `commonAncestor`
	// b. A reachability tree ancestor of `reindexRoot`
	commonAncestorChosenChild, err := rt.FindAncestorOfThisAmongChildrenOfOther(reindexRoot, commonAncestor)
	if err != nil {
		return err
	}

	nodeInterval, err := rt.interval(node)
	if err != nil {
		return err
	}

	commonAncestorChosenChildInterval, err := rt.interval(commonAncestorChosenChild)
	if err != nil {
		return err
	}

	if nodeInterval.End < commonAncestorChosenChildInterval.Start {
		// node is in the subtree before the chosen child
		return rt.reclaimIntervalBeforeChosenChild(node, commonAncestor,
			commonAncestorChosenChild, reindexRoot)
	}

	// node is either:
	// * in the subtree after the chosen child
	// * the common ancestor
	// In both cases we reclaim from the "after" subtree. In the
	// latter case this is arbitrary
	return rt.reclaimIntervalAfterChosenChild(node, commonAncestor,
		commonAncestorChosenChild, reindexRoot)
}

func (rt *reachabilityManager) reclaimIntervalBeforeChosenChild(rtn, commonAncestor, commonAncestorChosenChild,
	reindexRoot *externalapi.DomainHash) error {

	current := commonAncestorChosenChild

	commonAncestorChosenChildHasSlackIntervalBefore, err := rt.hasSlackIntervalBefore(commonAncestorChosenChild)
	if err != nil {
		return err
	}

	if !commonAncestorChosenChildHasSlackIntervalBefore {
		// The common ancestor ran out of slack before its chosen child.
		// Climb up the reachability tree toward the reindex root until
		// we find a node that has enough slack.
		for {
			currentHasSlackIntervalBefore, err := rt.hasSlackIntervalBefore(current)
			if err != nil {
				return err
			}

			if currentHasSlackIntervalBefore || current.Equal(reindexRoot) {
				break
			}

			current, err = rt.FindAncestorOfThisAmongChildrenOfOther(reindexRoot, current)
			if err != nil {
				return err
			}
		}

		if current.Equal(reindexRoot) {
			// "Deallocate" an interval of slackReachabilityIntervalForReclaiming
			// from this node. This is the interval that we'll use for the new
			// node.
			originalInterval, err := rt.interval(current)
			if err != nil {
				return err
			}

			err = rt.stageInterval(current, newReachabilityInterval(
				originalInterval.Start+slackReachabilityIntervalForReclaiming,
				originalInterval.End,
			))
			if err != nil {
				return err
			}

			err = rt.countSubtreesAndPropagateInterval(current)
			if err != nil {
				return err
			}

			err = rt.stageInterval(current, originalInterval)
			if err != nil {
				return err
			}
		}
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree before the
	// current node with an interval that is smaller by
	// slackReachabilityIntervalForReclaiming. This is to make room
	// for the new node.
	for !current.Equal(commonAncestor) {
		currentInterval, err := rt.interval(current)
		if err != nil {
			return err
		}

		err = rt.stageInterval(current, newReachabilityInterval(
			currentInterval.Start+slackReachabilityIntervalForReclaiming,
			currentInterval.End,
		))
		if err != nil {
			return err
		}

		currentParent, err := rt.parent(current)
		if err != nil {
			return err
		}

		err = rt.reindexIntervalsBeforeNode(currentParent, current)
		if err != nil {
			return err
		}
		current, err = rt.parent(current)
		if err != nil {
			return err
		}
	}

	return nil
}

// reindexIntervalsBeforeNode applies a tight interval to the reachability
// subtree before `node`. Note that `node` itself is unaffected.
func (rt *reachabilityManager) reindexIntervalsBeforeNode(rtn, node *externalapi.DomainHash) error {

	childrenBeforeNode, _, err := rt.splitChildrenAroundChild(rtn, node)
	if err != nil {
		return err
	}

	childrenBeforeNodeSizes, childrenBeforeNodeSubtreeSizeMaps, childrenBeforeNodeSizesSum :=
		rt.calcReachabilityTreeNodeSizes(childrenBeforeNode)

	// Apply a tight interval
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return err
	}

	newIntervalEnd := nodeInterval.Start - 1
	newInterval := newReachabilityInterval(newIntervalEnd-childrenBeforeNodeSizesSum+1, newIntervalEnd)
	intervals, err := intervalSplitExact(newInterval, childrenBeforeNodeSizes)
	if err != nil {
		return err
	}
	return rt.propagateIntervals(childrenBeforeNode, intervals, childrenBeforeNodeSubtreeSizeMaps)
}

func (rt *reachabilityManager) reclaimIntervalAfterChosenChild(node, commonAncestor, commonAncestorChosenChild,
	reindexRoot *externalapi.DomainHash) error {

	current := commonAncestorChosenChild
	commonAncestorChosenChildHasSlackIntervalAfter, err := rt.hasSlackIntervalAfter(commonAncestorChosenChild)
	if err != nil {
		return err
	}

	if !commonAncestorChosenChildHasSlackIntervalAfter {
		// The common ancestor ran out of slack after its chosen child.
		// Climb up the reachability tree toward the reindex root until
		// we find a node that has enough slack.
		for {
			currentHasSlackIntervalAfter, err := rt.hasSlackIntervalAfter(commonAncestorChosenChild)
			if err != nil {
				return err
			}

			if currentHasSlackIntervalAfter || current.Equal(reindexRoot) {
				break
			}

			current, err = rt.FindAncestorOfThisAmongChildrenOfOther(reindexRoot, current)
			if err != nil {
				return err
			}
		}

		if current.Equal(reindexRoot) {
			// "Deallocate" an interval of slackReachabilityIntervalForReclaiming
			// from this node. This is the interval that we'll use for the new
			// node.
			originalInterval, err := rt.interval(current)
			if err != nil {
				return err
			}

			err = rt.stageInterval(current, newReachabilityInterval(
				originalInterval.Start,
				originalInterval.End-slackReachabilityIntervalForReclaiming,
			))
			if err != nil {
				return err
			}

			err = rt.countSubtreesAndPropagateInterval(current)
			if err != nil {
				return err
			}

			err = rt.stageInterval(current, originalInterval)
			if err != nil {
				return err
			}
		}
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree after the
	// current node with an interval that is smaller by
	// slackReachabilityIntervalForReclaiming. This is to make room
	// for the new node.
	for !current.Equal(commonAncestor) {
		currentInterval, err := rt.interval(current)
		if err != nil {
			return err
		}

		err = rt.stageInterval(current, newReachabilityInterval(
			currentInterval.Start,
			currentInterval.End-slackReachabilityIntervalForReclaiming,
		))
		if err != nil {
			return err
		}

		currentParent, err := rt.parent(current)
		if err != nil {
			return err
		}

		err = rt.reindexIntervalsAfterNode(currentParent, current)
		if err != nil {
			return err
		}
		current = currentParent
	}

	return nil
}

// reindexIntervalsAfterNode applies a tight interval to the reachability
// subtree after `node`. Note that `node` itself is unaffected.
func (rt *reachabilityManager) reindexIntervalsAfterNode(rtn, node *externalapi.DomainHash) error {

	_, childrenAfterNode, err := rt.splitChildrenAroundChild(rtn, node)
	if err != nil {
		return err
	}

	childrenAfterNodeSizes, childrenAfterNodeSubtreeSizeMaps, childrenAfterNodeSizesSum :=
		rt.calcReachabilityTreeNodeSizes(childrenAfterNode)

	// Apply a tight interval
	nodeInterval, err := rt.interval(node)
	if err != nil {
		return err
	}

	newIntervalStart := nodeInterval.End + 1
	newInterval := newReachabilityInterval(newIntervalStart, newIntervalStart+childrenAfterNodeSizesSum-1)
	intervals, err := intervalSplitExact(newInterval, childrenAfterNodeSizes)
	if err != nil {
		return err
	}
	return rt.propagateIntervals(childrenAfterNode, intervals, childrenAfterNodeSubtreeSizeMaps)
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

// findCommonAncestorWithReindexRoot finds the most recent reachability
// tree ancestor common to both node and the given reindex root. Note
// that we assume that almost always the chain between the reindex root
// and the common ancestor is longer than the chain between node and the
// common ancestor.
func (rt *reachabilityManager) findCommonAncestorWithReindexRoot(node, reindexRoot *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	currentThis := node
	for {
		isAncestorOf, err := rt.IsReachabilityTreeAncestorOf(currentThis, reindexRoot)
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
		commonAncestor, err := rt.findCommonAncestorWithReindexRoot(newTreeNode, reindexRoot)
		if err != nil {
			return nil, false, err
		}

		return commonAncestor, true, nil
	}

	reindexRootChosenChild, err := rt.FindAncestorOfThisAmongChildrenOfOther(newTreeNode, reindexRoot)
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

	err = rt.concentrateIntervalAroundReindexRootChosenChild(reindexRoot, reindexRootChosenChild)
	if err != nil {
		return nil, false, err
	}

	return reindexRootChosenChild, true, nil
}

// FindAncestorOfThisAmongChildrenOfOther finds the reachability tree child
// of node that is the ancestor of node.
func (rt *reachabilityManager) FindAncestorOfThisAmongChildrenOfOther(this, other *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	otherChildren, err := rt.children(other)
	if err != nil {
		return nil, err
	}

	ancestor, ok := rt.findAncestorOfNode(otherChildren, this)
	if !ok {
		return nil, errors.Errorf("node is not an ancestor of this")
	}

	return ancestor, nil
}

func (rt *reachabilityManager) concentrateIntervalAroundReindexRootChosenChild(reindexRoot,
	reindexRootChosenChild *externalapi.DomainHash) error {

	reindexRootChildNodesBeforeChosen, reindexRootChildNodesAfterChosen, err :=
		rt.splitChildrenAroundChild(reindexRoot, reindexRootChosenChild)
	if err != nil {
		return err
	}

	reindexRootChildNodesBeforeChosenSizesSum, err :=
		rt.tightenIntervalsBeforeReindexRootChosenChild(reindexRoot, reindexRootChildNodesBeforeChosen)
	if err != nil {
		return err
	}

	reindexRootChildNodesAfterChosenSizesSum, err :=
		rt.tightenIntervalsAfterReindexRootChosenChild(reindexRoot, reindexRootChildNodesAfterChosen)
	if err != nil {
		return err
	}

	err = rt.expandIntervalInReindexRootChosenChild(reindexRoot, reindexRootChosenChild,
		reindexRootChildNodesBeforeChosenSizesSum, reindexRootChildNodesAfterChosenSizesSum)
	if err != nil {
		return err
	}

	return nil
}

// splitChildrenAroundChild splits `node` into two slices: the nodes that are before
// `child` and the nodes that are after.
func (rt *reachabilityManager) splitChildrenAroundChild(node, child *externalapi.DomainHash) (
	nodesBeforeChild, nodesAfterChild []*externalapi.DomainHash, err error) {

	nodeChildren, err := rt.children(node)
	if err != nil {
		return nil, nil, err
	}

	for i, candidateChild := range nodeChildren {
		if candidateChild.Equal(child) {
			return nodeChildren[:i], nodeChildren[i+1:], nil
		}
	}
	return nil, nil, errors.Errorf("child not a child of node")
}

func (rt *reachabilityManager) tightenIntervalsBeforeReindexRootChosenChild(
	reindexRoot *externalapi.DomainHash,
	reindexRootChildNodesBeforeChosen []*externalapi.DomainHash) (reindexRootChildNodesBeforeChosenSizesSum uint64,
	err error) {

	reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps, reindexRootChildNodesBeforeChosenSizesSum :=
		rt.calcReachabilityTreeNodeSizes(reindexRootChildNodesBeforeChosen)

	reindexRootInterval, err := rt.interval(reindexRoot)
	if err != nil {
		return 0, err
	}

	intervalBeforeReindexRootStart := newReachabilityInterval(
		reindexRootInterval.Start+rt.reindexSlack,
		reindexRootInterval.Start+rt.reindexSlack+reindexRootChildNodesBeforeChosenSizesSum-1,
	)

	err = rt.propagateChildIntervals(intervalBeforeReindexRootStart, reindexRootChildNodesBeforeChosen,
		reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps)
	if err != nil {
		return 0, err
	}
	return reindexRootChildNodesBeforeChosenSizesSum, nil
}

func (rt *reachabilityManager) tightenIntervalsAfterReindexRootChosenChild(
	reindexRoot *externalapi.DomainHash,
	reindexRootChildNodesAfterChosen []*externalapi.DomainHash) (reindexRootChildNodesAfterChosenSizesSum uint64,
	err error) {

	reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps,
		reindexRootChildNodesAfterChosenSizesSum :=
		rt.calcReachabilityTreeNodeSizes(reindexRootChildNodesAfterChosen)

	reindexRootInterval, err := rt.interval(reindexRoot)
	if err != nil {
		return 0, err
	}

	intervalAfterReindexRootEnd := newReachabilityInterval(
		reindexRootInterval.End-rt.reindexSlack-reindexRootChildNodesAfterChosenSizesSum,
		reindexRootInterval.End-rt.reindexSlack-1,
	)

	err = rt.propagateChildIntervals(intervalAfterReindexRootEnd, reindexRootChildNodesAfterChosen,
		reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps)
	if err != nil {
		return 0, err
	}
	return reindexRootChildNodesAfterChosenSizesSum, nil
}

func (rt *reachabilityManager) expandIntervalInReindexRootChosenChild(reindexRoot,
	reindexRootChosenChild *externalapi.DomainHash, reindexRootChildNodesBeforeChosenSizesSum uint64,
	reindexRootChildNodesAfterChosenSizesSum uint64) error {

	reindexRootInterval, err := rt.interval(reindexRoot)
	if err != nil {
		return err
	}

	newReindexRootChildInterval := newReachabilityInterval(
		reindexRootInterval.Start+reindexRootChildNodesBeforeChosenSizesSum+rt.reindexSlack,
		reindexRootInterval.End-reindexRootChildNodesAfterChosenSizesSum-rt.reindexSlack-1,
	)

	reindexRootChosenChildInterval, err := rt.interval(reindexRootChosenChild)
	if err != nil {
		return err
	}

	if !intervalContains(newReindexRootChildInterval, reindexRootChosenChildInterval) {
		// New interval doesn't contain the previous one, propagation is required

		// We assign slack on both sides as an optimization. Were we to
		// assign a tight interval, the next time the reindex root moves we
		// would need to propagate intervals again. That is to say, When we
		// DO allocate slack, next time
		// expandIntervalInReindexRootChosenChild is called (next time the
		// reindex root moves), newReindexRootChildInterval is likely to
		// contain reindexRootChosenChild.Interval.
		err := rt.stageInterval(reindexRootChosenChild, newReachabilityInterval(
			newReindexRootChildInterval.Start+rt.reindexSlack,
			newReindexRootChildInterval.End-rt.reindexSlack,
		))
		if err != nil {
			return err
		}

		err = rt.countSubtreesAndPropagateInterval(reindexRootChosenChild)
		if err != nil {
			return err
		}
	}

	err = rt.stageInterval(reindexRootChosenChild, newReindexRootChildInterval)
	if err != nil {
		return err
	}
	return nil
}

func (rt *reachabilityManager) countSubtreesAndPropagateInterval(node *externalapi.DomainHash) error {
	subtreeSizeMap := make(map[externalapi.DomainHash]uint64)
	err := rt.countSubtrees(node, subtreeSizeMap)
	if err != nil {
		return err
	}

	return rt.propagateInterval(node, subtreeSizeMap)
}

func (rt *reachabilityManager) calcReachabilityTreeNodeSizes(treeNodes []*externalapi.DomainHash) (
	sizes []uint64, subtreeSizeMaps []map[externalapi.DomainHash]uint64, sum uint64) {

	sizes = make([]uint64, len(treeNodes))
	subtreeSizeMaps = make([]map[externalapi.DomainHash]uint64, len(treeNodes))
	sum = 0
	for i, node := range treeNodes {
		subtreeSizeMap := make(map[externalapi.DomainHash]uint64)
		err := rt.countSubtrees(node, subtreeSizeMap)
		if err != nil {
			return nil, nil, 0
		}

		subtreeSize := subtreeSizeMap[*node]
		sizes[i] = subtreeSize
		subtreeSizeMaps[i] = subtreeSizeMap
		sum += subtreeSize
	}
	return sizes, subtreeSizeMaps, sum
}

func (rt *reachabilityManager) propagateChildIntervals(interval *model.ReachabilityInterval,
	childNodes []*externalapi.DomainHash, sizes []uint64, subtreeSizeMaps []map[externalapi.DomainHash]uint64) error {

	childIntervalSizes, err := intervalSplitExact(interval, sizes)
	if err != nil {
		return err
	}

	for i, child := range childNodes {
		childInterval := childIntervalSizes[i]
		err := rt.stageInterval(child, childInterval)
		if err != nil {
			return err
		}

		childSubtreeSizeMap := subtreeSizeMaps[i]
		err = rt.propagateInterval(child, childSubtreeSizeMap)
		if err != nil {
			return err
		}
	}

	return nil
}
