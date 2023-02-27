package reachabilitymanager

import (
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

// reindexContext is a struct used during reindex operations. It represents a temporary context
// for caching subtree information during the *current* reindex operation only
type reindexContext struct {
	manager           *reachabilityManager
	subTreeSizesCache map[externalapi.DomainHash]uint64
}

// newReindexContext creates a new empty reindex context
func newReindexContext(rt *reachabilityManager) reindexContext {
	return reindexContext{
		manager:           rt,
		subTreeSizesCache: make(map[externalapi.DomainHash]uint64),
	}
}

/*

Core (BFS) algorithms used during reindexing

*/

// countSubtrees counts the size of each subtree under this node,
// and populates the provided subTreeSizeMap with the results.
// It is equivalent to the following recursive implementation:
//
//	func (rt *reachabilityManager) countSubtrees(node *model.ReachabilityTreeNode) uint64 {
//	    subtreeSize := uint64(0)
//	    for _, child := range node.children {
//	        subtreeSize += child.countSubtrees()
//	    }
//	    return subtreeSize + 1
//	}
//
// However, we are expecting (linearly) deep trees, and so a
// recursive stack-based approach is inefficient and will hit
// recursion limits. Instead, the same logic was implemented
// using a (queue-based) BFS method. At a high level, the
// algorithm uses BFS for reaching all leaves and pushes
// intermediate updates from leaves via parent chains until all
// size information is gathered at the root of the operation
// (i.e. at node).
func (rc *reindexContext) countSubtrees(stagingArea *model.StagingArea, node *externalapi.DomainHash) error {

	if _, ok := rc.subTreeSizesCache[*node]; ok {
		return nil
	}

	queue := []*externalapi.DomainHash{node}
	calculatedChildrenCount := make(map[externalapi.DomainHash]uint64)
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]
		children, err := rc.manager.children(stagingArea, current)
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
			current, err = rc.manager.parent(stagingArea, current)
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

			children, err := rc.manager.children(stagingArea, current)
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
// the allocation rule in splitWithExponentialBias.
func (rc *reindexContext) propagateInterval(stagingArea *model.StagingArea, node *externalapi.DomainHash) error {

	// Make sure subtrees are counted before propagating
	err := rc.countSubtrees(stagingArea, node)
	if err != nil {
		return err
	}

	queue := []*externalapi.DomainHash{node}
	for len(queue) > 0 {
		var current *externalapi.DomainHash
		current, queue = queue[0], queue[1:]

		children, err := rc.manager.children(stagingArea, current)
		if err != nil {
			return err
		}

		if len(children) > 0 {
			sizes := make([]uint64, len(children))
			for i, child := range children {
				sizes[i] = rc.subTreeSizesCache[*child]
			}

			interval, err := rc.manager.intervalRangeForChildAllocation(stagingArea, current)
			if err != nil {
				return err
			}

			intervals, err := intervalSplitWithExponentialBias(interval, sizes)
			if err != nil {
				return err
			}

			for i, child := range children {
				childInterval := intervals[i]
				err = rc.manager.stageInterval(stagingArea, child, childInterval)
				if err != nil {
					return err
				}
				queue = append(queue, child)
			}
		}
	}
	return nil
}

/*

Functions for handling reindex triggered by adding child block

*/

// reindexIntervals traverses the reachability subtree that's
// defined by the new child node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter. It does this by traversing down the reachability
// tree until it finds a node with a subtree size that's greater than
// its interval size. See propagateInterval for further details.
func (rc *reindexContext) reindexIntervals(stagingArea *model.StagingArea, newChild, reindexRoot *externalapi.DomainHash) error {
	current := newChild
	// Search for the first ancestor with sufficient interval space
	for {
		currentInterval, err := rc.manager.interval(stagingArea, current)
		if err != nil {
			return err
		}

		currentIntervalSize := intervalSize(currentInterval)

		err = rc.countSubtrees(stagingArea, current)
		if err != nil {
			return err
		}

		currentSubtreeSize := rc.subTreeSizesCache[*current]

		// Current has sufficient space, break and propagate
		if currentIntervalSize >= currentSubtreeSize {
			break
		}

		parent, err := rc.manager.parent(stagingArea, current)
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

		if current.Equal(reindexRoot) {
			// Reindex root is expected to hold enough capacity as long as there are less
			// than ~2^52 blocks in the DAG, which should never happen in our lifetimes
			// even if block rate per second is above 100. The calculation follows from the allocation of
			// 2^12 (which equals 2^64/2^52) for slack per chain block below the reindex root.
			return errors.Errorf("unexpected behavior: reindex root %s is out of capacity"+
				"during reindexing. Theoretically, this "+
				"should only ever happen if there are more "+
				"than ~2^52 blocks in the DAG.", reindexRoot.String())
		}

		isParentStrictAncestorOfRoot, err := rc.manager.isStrictAncestorOf(stagingArea, parent, reindexRoot)
		if err != nil {
			return err
		}

		if isParentStrictAncestorOfRoot {
			// In this case parent is guaranteed to have sufficient interval space,
			// however we avoid reindexing the entire subtree above parent
			// (which includes root and thus majority of blocks mined since)
			// and use slacks along the chain up from parent to reindex root.
			// Notes:
			// 1. we set requiredAllocation=currentSubtreeSize in order to double the
			// current interval capacity
			// 2. it might be the case that current is the `newChild` itself
			return rc.reindexIntervalsEarlierThanRoot(stagingArea, current, reindexRoot, parent, currentSubtreeSize)
		}

		current = parent
	}

	// Propagate the interval to the subtree
	return rc.propagateInterval(stagingArea, current)
}

// reindexIntervalsEarlierThanRoot implements the reindex algorithm for the case where the
// new child node is not in reindex root's subtree. The function is expected to allocate
// `requiredAllocation` to be added to interval of `allocationNode`. `commonAncestor` is
// expected to be a direct parent of `allocationNode` and an ancestor of `reindexRoot`.
func (rc *reindexContext) reindexIntervalsEarlierThanRoot(stagingArea *model.StagingArea,
	allocationNode, reindexRoot, commonAncestor *externalapi.DomainHash, requiredAllocation uint64) error {

	// The chosen child is:
	// a. A reachability tree child of `commonAncestor`
	// b. A reachability tree ancestor of `reindexRoot` or `reindexRoot` itself
	chosenChild, err := rc.manager.FindNextAncestor(stagingArea, reindexRoot, commonAncestor)
	if err != nil {
		return err
	}

	nodeInterval, err := rc.manager.interval(stagingArea, allocationNode)
	if err != nil {
		return err
	}

	chosenInterval, err := rc.manager.interval(stagingArea, chosenChild)
	if err != nil {
		return err
	}

	if nodeInterval.Start < chosenInterval.Start {
		// allocationNode is in the subtree before the chosen child
		return rc.reclaimIntervalBefore(stagingArea, allocationNode, commonAncestor, chosenChild, reindexRoot, requiredAllocation)
	}

	// allocationNode is in the subtree after the chosen child
	return rc.reclaimIntervalAfter(stagingArea, allocationNode, commonAncestor, chosenChild, reindexRoot, requiredAllocation)
}

func (rc *reindexContext) reclaimIntervalBefore(stagingArea *model.StagingArea,
	allocationNode, commonAncestor, chosenChild, reindexRoot *externalapi.DomainHash, requiredAllocation uint64) error {

	var slackSum uint64 = 0
	var pathLen uint64 = 0
	var pathSlackAlloc uint64 = 0

	var err error
	current := chosenChild

	// Walk up the chain from common ancestor's chosen child towards reindex root
	for {
		if current.Equal(reindexRoot) {
			// Reached reindex root. In this case, since we reached (the unlimited) root,
			// we also re-allocate new slack for the chain we just traversed

			previousInterval, err := rc.manager.interval(stagingArea, current)
			if err != nil {
				return err
			}

			offset := requiredAllocation + rc.manager.reindexSlack*pathLen - slackSum
			err = rc.manager.stageInterval(stagingArea, current, intervalIncreaseStart(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(stagingArea, current)
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsBefore(stagingArea, allocationNode, current, offset)
			if err != nil {
				return err
			}

			// Set the slack for each chain block to be reserved below during the chain walk-down
			pathSlackAlloc = rc.manager.reindexSlack
			break
		}

		slackBeforeCurrent, err := rc.manager.remainingSlackBefore(stagingArea, current)
		if err != nil {
			return err
		}
		slackSum += slackBeforeCurrent

		if slackSum >= requiredAllocation {
			previousInterval, err := rc.manager.interval(stagingArea, current)
			if err != nil {
				return err
			}

			// Set offset to be just enough to satisfy required allocation
			offset := slackBeforeCurrent - (slackSum - requiredAllocation)

			err = rc.manager.stageInterval(stagingArea, current, intervalIncreaseStart(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsBefore(stagingArea, allocationNode, current, offset)
			if err != nil {
				return err
			}

			break
		}

		current, err = rc.manager.FindNextAncestor(stagingArea, reindexRoot, current)
		if err != nil {
			return err
		}

		pathLen++
	}

	// Go back down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree before the
	// current node with an interval that is smaller.
	// This is to make room for the required allocation.
	for {
		current, err = rc.manager.parent(stagingArea, current)
		if err != nil {
			return err
		}

		if current.Equal(commonAncestor) {
			break
		}

		originalInterval, err := rc.manager.interval(stagingArea, current)
		if err != nil {
			return err
		}

		slackBeforeCurrent, err := rc.manager.remainingSlackBefore(stagingArea, current)
		if err != nil {
			return err
		}

		offset := slackBeforeCurrent - pathSlackAlloc
		err = rc.manager.stageInterval(stagingArea, current, intervalIncreaseStart(originalInterval, offset))
		if err != nil {
			return err
		}

		err = rc.offsetSiblingsBefore(stagingArea, allocationNode, current, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) offsetSiblingsBefore(stagingArea *model.StagingArea,
	allocationNode, current *externalapi.DomainHash, offset uint64) error {

	parent, err := rc.manager.parent(stagingArea, current)
	if err != nil {
		return err
	}

	siblingsBefore, _, err := rc.manager.splitChildren(stagingArea, parent, current)
	if err != nil {
		return err
	}

	// Iterate over the slice in reverse order in order to break if reaching `allocationNode`
	for i := len(siblingsBefore) - 1; i >= 0; i-- {
		sibling := siblingsBefore[i]
		if sibling.Equal(allocationNode) {
			// We reached our final destination, allocate `offset` to `allocationNode` by increasing end and break
			previousInterval, err := rc.manager.interval(stagingArea, allocationNode)
			if err != nil {
				return err
			}

			err = rc.manager.stageInterval(stagingArea, allocationNode, intervalIncreaseEnd(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(stagingArea, allocationNode)
			if err != nil {
				return err
			}

			break
		}

		previousInterval, err := rc.manager.interval(stagingArea, sibling)
		if err != nil {
			return err
		}

		err = rc.manager.stageInterval(stagingArea, sibling, intervalIncrease(previousInterval, offset))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(stagingArea, sibling)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) reclaimIntervalAfter(stagingArea *model.StagingArea,
	allocationNode, commonAncestor, chosenChild, reindexRoot *externalapi.DomainHash, requiredAllocation uint64) error {

	var slackSum uint64 = 0
	var pathLen uint64 = 0
	var pathSlackAlloc uint64 = 0

	var err error
	current := chosenChild

	// Walk up the chain from common ancestor's chosen child towards reindex root
	for {
		if current.Equal(reindexRoot) {
			// Reached reindex root. In this case, since we reached (the unlimited) root,
			// we also re-allocate new slack for the chain we just traversed

			previousInterval, err := rc.manager.interval(stagingArea, current)
			if err != nil {
				return err
			}

			offset := requiredAllocation + rc.manager.reindexSlack*pathLen - slackSum
			err = rc.manager.stageInterval(stagingArea, current, intervalDecreaseEnd(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(stagingArea, current)
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsAfter(stagingArea, allocationNode, current, offset)
			if err != nil {
				return err
			}

			// Set the slack for each chain block to be reserved below during the chain walk-down
			pathSlackAlloc = rc.manager.reindexSlack
			break
		}

		slackAfterCurrent, err := rc.manager.remainingSlackAfter(stagingArea, current)
		if err != nil {
			return err
		}
		slackSum += slackAfterCurrent

		if slackSum >= requiredAllocation {
			previousInterval, err := rc.manager.interval(stagingArea, current)
			if err != nil {
				return err
			}

			// Set offset to be just enough to satisfy required allocation
			offset := slackAfterCurrent - (slackSum - requiredAllocation)

			err = rc.manager.stageInterval(stagingArea, current, intervalDecreaseEnd(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.offsetSiblingsAfter(stagingArea, allocationNode, current, offset)
			if err != nil {
				return err
			}

			break
		}

		current, err = rc.manager.FindNextAncestor(stagingArea, reindexRoot, current)
		if err != nil {
			return err
		}

		pathLen++
	}

	// Go back down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree before the
	// current node with an interval that is smaller.
	// This is to make room for the required allocation.
	for {
		current, err = rc.manager.parent(stagingArea, current)
		if err != nil {
			return err
		}

		if current.Equal(commonAncestor) {
			break
		}

		originalInterval, err := rc.manager.interval(stagingArea, current)
		if err != nil {
			return err
		}

		slackAfterCurrent, err := rc.manager.remainingSlackAfter(stagingArea, current)
		if err != nil {
			return err
		}

		offset := slackAfterCurrent - pathSlackAlloc
		err = rc.manager.stageInterval(stagingArea, current, intervalDecreaseEnd(originalInterval, offset))
		if err != nil {
			return err
		}

		err = rc.offsetSiblingsAfter(stagingArea, allocationNode, current, offset)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *reindexContext) offsetSiblingsAfter(stagingArea *model.StagingArea,
	allocationNode, current *externalapi.DomainHash, offset uint64) error {

	parent, err := rc.manager.parent(stagingArea, current)
	if err != nil {
		return err
	}

	_, siblingsAfter, err := rc.manager.splitChildren(stagingArea, parent, current)
	if err != nil {
		return err
	}

	for _, sibling := range siblingsAfter {
		if sibling.Equal(allocationNode) {
			// We reached our final destination, allocate `offset` to `allocationNode` by decreasing start and break
			previousInterval, err := rc.manager.interval(stagingArea, allocationNode)
			if err != nil {
				return err
			}

			err = rc.manager.stageInterval(stagingArea, allocationNode, intervalDecreaseStart(previousInterval, offset))
			if err != nil {
				return err
			}

			err = rc.propagateInterval(stagingArea, allocationNode)
			if err != nil {
				return err
			}

			break
		}

		previousInterval, err := rc.manager.interval(stagingArea, sibling)
		if err != nil {
			return err
		}

		err = rc.manager.stageInterval(stagingArea, sibling, intervalDecrease(previousInterval, offset))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(stagingArea, sibling)
		if err != nil {
			return err
		}
	}

	return nil
}

/*

Functions for handling reindex triggered by moving reindex root

*/

func (rc *reindexContext) concentrateInterval(stagingArea *model.StagingArea,
	reindexRoot, chosenChild *externalapi.DomainHash, isFinalReindexRoot bool) error {

	siblingsBeforeChosen, siblingsAfterChosen, err := rc.manager.splitChildren(stagingArea, reindexRoot, chosenChild)
	if err != nil {
		return err
	}

	siblingsBeforeSizesSum, err := rc.tightenIntervalsBefore(stagingArea, reindexRoot, siblingsBeforeChosen)
	if err != nil {
		return err
	}

	siblingsAfterSizesSum, err := rc.tightenIntervalsAfter(stagingArea, reindexRoot, siblingsAfterChosen)
	if err != nil {
		return err
	}

	err = rc.expandIntervalToChosen(stagingArea,
		reindexRoot, chosenChild, siblingsBeforeSizesSum, siblingsAfterSizesSum, isFinalReindexRoot)
	if err != nil {
		return err
	}

	return nil
}

func (rc *reindexContext) tightenIntervalsBefore(stagingArea *model.StagingArea,
	reindexRoot *externalapi.DomainHash, siblingsBeforeChosen []*externalapi.DomainHash) (sizesSum uint64, err error) {

	siblingSubtreeSizes, sizesSum := rc.countChildrenSubtrees(stagingArea, siblingsBeforeChosen)

	rootInterval, err := rc.manager.interval(stagingArea, reindexRoot)
	if err != nil {
		return 0, err
	}

	intervalBeforeChosen := newReachabilityInterval(
		rootInterval.Start+rc.manager.reindexSlack,
		rootInterval.Start+rc.manager.reindexSlack+sizesSum-1,
	)

	err = rc.propagateChildrenIntervals(stagingArea, intervalBeforeChosen, siblingsBeforeChosen, siblingSubtreeSizes)
	if err != nil {
		return 0, err
	}

	return sizesSum, nil
}

func (rc *reindexContext) tightenIntervalsAfter(stagingArea *model.StagingArea,
	reindexRoot *externalapi.DomainHash, siblingsAfterChosen []*externalapi.DomainHash) (sizesSum uint64, err error) {

	siblingSubtreeSizes, sizesSum := rc.countChildrenSubtrees(stagingArea, siblingsAfterChosen)

	rootInterval, err := rc.manager.interval(stagingArea, reindexRoot)
	if err != nil {
		return 0, err
	}

	intervalAfterChosen := newReachabilityInterval(
		rootInterval.End-rc.manager.reindexSlack-sizesSum,
		rootInterval.End-rc.manager.reindexSlack-1,
	)

	err = rc.propagateChildrenIntervals(stagingArea, intervalAfterChosen, siblingsAfterChosen, siblingSubtreeSizes)
	if err != nil {
		return 0, err
	}

	return sizesSum, nil
}

func (rc *reindexContext) expandIntervalToChosen(stagingArea *model.StagingArea,
	reindexRoot, chosenChild *externalapi.DomainHash, sizesSumBefore, sizesSumAfter uint64, isFinalReindexRoot bool) error {

	rootInterval, err := rc.manager.interval(stagingArea, reindexRoot)
	if err != nil {
		return err
	}

	newChosenInterval := newReachabilityInterval(
		rootInterval.Start+sizesSumBefore+rc.manager.reindexSlack,
		rootInterval.End-sizesSumAfter-rc.manager.reindexSlack-1,
	)

	currentChosenInterval, err := rc.manager.interval(stagingArea, chosenChild)
	if err != nil {
		return err
	}

	// Propagate interval only if chosenChild is the final reindex root
	if isFinalReindexRoot && !intervalContains(newChosenInterval, currentChosenInterval) {
		// New interval doesn't contain the previous one, propagation is required

		// We assign slack on both sides as an optimization. Were we to
		// assign a tight interval, the next time the reindex root moves we
		// would need to propagate intervals again. That is to say, when we
		// do allocate slack, next time
		// expandIntervalToChosen is called (next time the
		// reindex root moves), newChosenInterval is likely to
		// contain currentChosenInterval.
		err := rc.manager.stageInterval(stagingArea, chosenChild, newReachabilityInterval(
			newChosenInterval.Start+rc.manager.reindexSlack,
			newChosenInterval.End-rc.manager.reindexSlack,
		))
		if err != nil {
			return err
		}

		err = rc.propagateInterval(stagingArea, chosenChild)
		if err != nil {
			return err
		}
	}

	err = rc.manager.stageInterval(stagingArea, chosenChild, newChosenInterval)
	if err != nil {
		return err
	}

	return nil
}

func (rc *reindexContext) countChildrenSubtrees(stagingArea *model.StagingArea, children []*externalapi.DomainHash) (
	sizes []uint64, sum uint64) {

	sizes = make([]uint64, len(children))
	sum = 0
	for i, node := range children {
		err := rc.countSubtrees(stagingArea, node)
		if err != nil {
			return nil, 0
		}

		subtreeSize := rc.subTreeSizesCache[*node]
		sizes[i] = subtreeSize
		sum += subtreeSize
	}
	return sizes, sum
}

func (rc *reindexContext) propagateChildrenIntervals(stagingArea *model.StagingArea,
	interval *model.ReachabilityInterval, children []*externalapi.DomainHash, sizes []uint64) error {

	childIntervals, err := intervalSplitExact(interval, sizes)
	if err != nil {
		return err
	}

	for i, child := range children {
		childInterval := childIntervals[i]
		err := rc.manager.stageInterval(stagingArea, child, childInterval)
		if err != nil {
			return err
		}

		err = rc.propagateInterval(stagingArea, child)
		if err != nil {
			return err
		}
	}

	return nil
}
