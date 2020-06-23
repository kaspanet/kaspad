package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"strings"
	"time"
)

var (
	// reachabilityReindexWindow is the target window size for reachability
	// reindexes. Note that this is not a constant for testing purposes.
	reachabilityReindexWindow uint64 = 200

	// reachabilityReindexSlack is the slack interval given to reachability
	// tree nodes not in the selected parent chain. Note that this is not
	// a constant for testing purposes.
	reachabilityReindexSlack uint64 = 1 << 12
)

// modifiedTreeNodes are a set of reachabilityTreeNodes that's bubbled up
// from any function that modifies them, so that the original caller may
// update the database accordingly. This is a set rather than a slice due
// to frequent duplicate treeNodes between operations.
type modifiedTreeNodes map[*reachabilityTreeNode]struct{}

func newModifiedTreeNodes(nodes ...*reachabilityTreeNode) modifiedTreeNodes {
	modifiedNodes := make(modifiedTreeNodes)
	for _, node := range nodes {
		modifiedNodes[node] = struct{}{}
	}
	return modifiedNodes
}

// copyAllFrom copies all the reachabilityTreeNodes from `other`
// into `mtn`. Note that `other` is not affected.
func (mtn modifiedTreeNodes) copyAllFrom(other modifiedTreeNodes) {
	for node := range other {
		mtn[node] = struct{}{}
	}
}

// reachabilityInterval represents an interval to be used within the
// tree reachability algorithm. See reachabilityTreeNode for further
// details.
type reachabilityInterval struct {
	start uint64
	end   uint64
}

func newReachabilityInterval(start uint64, end uint64) *reachabilityInterval {
	return &reachabilityInterval{start: start, end: end}
}

// size returns the size of this interval. Note that intervals are
// inclusive from both sides.
func (ri *reachabilityInterval) size() uint64 {
	return ri.end - ri.start + 1
}

// splitInHalf splits this interval by a fraction of 0.5.
// See splitFraction for further details.
func (ri *reachabilityInterval) splitInHalf() (
	left *reachabilityInterval, right *reachabilityInterval, err error) {

	return ri.splitFraction(0.5)
}

// splitFraction splits this interval to two parts such that their
// union is equal to the original interval and the first (left) part
// contains the given fraction of the original interval's size.
// Note: if the split results in fractional parts, this method rounds
// the first part up and the last part down.
func (ri *reachabilityInterval) splitFraction(fraction float64) (
	left *reachabilityInterval, right *reachabilityInterval, err error) {

	if fraction < 0 || fraction > 1 {
		return nil, nil, errors.Errorf("fraction must be between 0 and 1")
	}
	if ri.size() == 0 {
		return nil, nil, errors.Errorf("cannot split an empty interval")
	}

	allocationSize := uint64(math.Ceil(float64(ri.size()) * fraction))
	left = newReachabilityInterval(ri.start, ri.start+allocationSize-1)
	right = newReachabilityInterval(ri.start+allocationSize, ri.end)
	return left, right, nil
}

// splitExact splits this interval to exactly |sizes| parts where
// |part_i| = sizes[i]. This method expects sum(sizes) to be exactly
// equal to the interval's size.
func (ri *reachabilityInterval) splitExact(sizes []uint64) ([]*reachabilityInterval, error) {
	sizesSum := uint64(0)
	for _, size := range sizes {
		sizesSum += size
	}
	if sizesSum != ri.size() {
		return nil, errors.Errorf("sum of sizes must be equal to the interval's size")
	}

	intervals := make([]*reachabilityInterval, len(sizes))
	start := ri.start
	for i, size := range sizes {
		intervals[i] = newReachabilityInterval(start, start+size-1)
		start += size
	}
	return intervals, nil
}

// splitWithExponentialBias splits this interval to |sizes| parts
// by the allocation rule described below. This method expects sum(sizes)
// to be smaller or equal to the interval's size. Every part_i is
// allocated at least sizes[i] capacity. The remaining budget is
// split by an exponentially biased rule described below.
//
// This rule follows the GHOSTDAG protocol behavior where the child
// with the largest subtree is expected to dominate the competition
// for new blocks and thus grow the most. However, we may need to
// add slack for non-largest subtrees in order to make CPU reindexing
// attacks unworthy.
func (ri *reachabilityInterval) splitWithExponentialBias(sizes []uint64) ([]*reachabilityInterval, error) {
	intervalSize := ri.size()
	sizesSum := uint64(0)
	for _, size := range sizes {
		sizesSum += size
	}
	if sizesSum > intervalSize {
		return nil, errors.Errorf("sum of sizes must be less than or equal to the interval's size")
	}
	if sizesSum == intervalSize {
		return ri.splitExact(sizes)
	}

	// Add a fractional bias to every size in the given sizes
	totalBias := intervalSize - sizesSum
	remainingBias := totalBias
	biasedSizes := make([]uint64, len(sizes))
	fractions := exponentialFractions(sizes)
	for i, fraction := range fractions {
		var bias uint64
		if i == len(fractions)-1 {
			bias = remainingBias
		} else {
			bias = uint64(math.Round(float64(totalBias) * fraction))
			if bias > remainingBias {
				bias = remainingBias
			}
		}
		biasedSizes[i] = sizes[i] + bias
		remainingBias -= bias
	}
	return ri.splitExact(biasedSizes)
}

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

// isAncestorOf checks if this interval's node is a reachability tree
// ancestor of the other interval's node. The condition below is relying on the
// property of reachability intervals that intervals are either completely disjoint,
// or one strictly contains the other.
func (ri *reachabilityInterval) isAncestorOf(other *reachabilityInterval) bool {
	return ri.start <= other.end && other.end <= ri.end
}

// String returns a string representation of the interval.
func (ri *reachabilityInterval) String() string {
	return fmt.Sprintf("[%d,%d]", ri.start, ri.end)
}

// reachabilityTreeNode represents a node in the reachability tree
// of some DAG block. It mainly provides the ability to query *tree*
// reachability with O(1) query time. It does so by managing an
// index interval for each node and making sure all nodes in its
// subtree are indexed within the interval, so the query
// B ∈ subtree(A) simply becomes B.interval ⊂ A.interval.
//
// The main challenge of maintaining such intervals is that our tree
// is an ever-growing tree and as such pre-allocated intervals may
// not suffice as per future events. This is where the reindexing
// algorithm below comes into place.
// We use the reasonable assumption that the initial root interval
// (e.g., [0, 2^64-1]) should always suffice for any practical use-
// case, and so reindexing should always succeed unless more than
// 2^64 blocks are added to the DAG/tree.
type reachabilityTreeNode struct {
	blockNode *blockNode

	children []*reachabilityTreeNode
	parent   *reachabilityTreeNode

	// interval is the index interval containing all intervals of
	// blocks in this node's subtree
	interval *reachabilityInterval
}

func newReachabilityTreeNode(blockNode *blockNode) *reachabilityTreeNode {
	// Please see the comment above reachabilityTreeNode to understand why
	// we use these initial values.
	interval := newReachabilityInterval(1, math.MaxUint64-1)
	return &reachabilityTreeNode{blockNode: blockNode, interval: interval}
}

func (rtn *reachabilityTreeNode) childIntervalAllocationRange() *reachabilityInterval {
	// We subtract 1 from the end of the range to prevent the node from allocating
	// the entire interval to its child, so its interval would *strictly* contain the interval of its child.
	return newReachabilityInterval(rtn.interval.start, rtn.interval.end-1)
}

func (rtn *reachabilityTreeNode) remainingIntervalBefore() *reachabilityInterval {
	childRange := rtn.childIntervalAllocationRange()
	if len(rtn.children) == 0 {
		return childRange
	}
	return newReachabilityInterval(childRange.start, rtn.children[0].interval.start-1)
}

func (rtn *reachabilityTreeNode) remainingIntervalAfter() *reachabilityInterval {
	childRange := rtn.childIntervalAllocationRange()
	if len(rtn.children) == 0 {
		return childRange
	}
	return newReachabilityInterval(rtn.children[len(rtn.children)-1].interval.end+1, childRange.end)
}

func (rtn *reachabilityTreeNode) hasSlackIntervalBefore() bool {
	return rtn.remainingIntervalBefore().size() > 0
}

func (rtn *reachabilityTreeNode) hasSlackIntervalAfter() bool {
	return rtn.remainingIntervalAfter().size() > 0
}

// addChild adds child to this tree node. If this node has no
// remaining interval to allocate, a reindexing is triggered.
// This method returns a list of reachabilityTreeNodes modified
// by it.
func (rtn *reachabilityTreeNode) addChild(child *reachabilityTreeNode, reindexRoot *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	remaining := rtn.remainingIntervalAfter()

	// Set the parent-child relationship
	rtn.children = append(rtn.children, child)
	child.parent = rtn

	// Handle rtn not being a descendant of the reindex root.
	// Note that we check rtn here instead of child because
	// at this point we don't yet know child's interval.
	if !reindexRoot.isAncestorOf(rtn) {
		reindexStartTime := time.Now()
		modifiedNodes, err := rtn.reindexIntervalsEarlierThanReindexRoot(reindexRoot)
		if err != nil {
			return nil, err
		}
		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. This block is not a child of the current "+
			"reindex root %s. Modified %d tree nodes and took %dms.",
			rtn.blockNode.hash, reindexRoot.blockNode.hash,
			len(modifiedNodes), reindexTimeElapsed.Milliseconds())
		return modifiedNodes, nil
	}

	// No allocation space left -- reindex
	if remaining.size() == 0 {
		reindexStartTime := time.Now()
		modifiedNodes, err := rtn.reindexIntervals()
		if err != nil {
			return nil, err
		}
		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. Modified %d tree nodes and took %dms.",
			rtn.blockNode.hash, len(modifiedNodes), reindexTimeElapsed.Milliseconds())
		return modifiedNodes, nil
	}

	// Allocate from the remaining space
	allocated, _, err := remaining.splitInHalf()
	if err != nil {
		return nil, err
	}
	child.interval = allocated
	return newModifiedTreeNodes(rtn, child), nil
}

// reindexIntervals traverses the reachability subtree that's
// defined by this node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter. It does this by traversing down the reachability
// tree until it finds a node with a subreeSize that's greater than
// its interval size. See propagateInterval for further details.
// This method returns a list of reachabilityTreeNodes modified by it.
func (rtn *reachabilityTreeNode) reindexIntervals() (modifiedTreeNodes, error) {
	current := rtn

	// Initial interval and subtree sizes
	intervalSize := current.interval.size()
	subTreeSizeMap := make(map[*reachabilityTreeNode]uint64)
	current.countSubtrees(subTreeSizeMap)
	currentSubtreeSize := subTreeSizeMap[current]

	// Find the first ancestor that has sufficient interval space
	for intervalSize < currentSubtreeSize {
		if current.parent == nil {
			// If we ended up here it means that there are more
			// than 2^64 blocks, which shouldn't ever happen.
			return nil, errors.Errorf("missing tree " +
				"parent during reindexing. Theoretically, this " +
				"should only ever happen if there are more " +
				"than 2^64 blocks in the DAG.")
		}
		current = current.parent
		intervalSize = current.interval.size()
		current.countSubtrees(subTreeSizeMap)
		currentSubtreeSize = subTreeSizeMap[current]
	}

	// Propagate the interval to the subtree
	return current.propagateInterval(subTreeSizeMap)
}

// countSubtrees counts the size of each subtree under this node,
// and populates the provided subTreeSizeMap with the results.
// It is equivalent to the following recursive implementation:
//
// func (rtn *reachabilityTreeNode) countSubtrees() uint64 {
//     subtreeSize := uint64(0)
//     for _, child := range rtn.children {
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
// (i.e. at rtn).
func (rtn *reachabilityTreeNode) countSubtrees(subTreeSizeMap map[*reachabilityTreeNode]uint64) {
	queue := []*reachabilityTreeNode{rtn}
	calculatedChildrenCount := make(map[*reachabilityTreeNode]uint64)
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) == 0 {
			// We reached a leaf
			subTreeSizeMap[current] = 1
		} else if _, ok := subTreeSizeMap[current]; !ok {
			// We haven't yet calculated the subtree size of
			// the current node. Add all its children to the
			// queue
			queue = append(queue, current.children...)
			continue
		}

		// We reached a leaf or a pre-calculated subtree.
		// Push information up
		for current != rtn {
			current = current.parent
			calculatedChildrenCount[current]++
			if calculatedChildrenCount[current] != uint64(len(current.children)) {
				// Not all subtrees of the current node are ready
				break
			}
			// All children of `current` have calculated their subtree size.
			// Sum them all together and add 1 to get the sub tree size of
			// `current`.
			childSubtreeSizeSum := uint64(0)
			for _, child := range current.children {
				childSubtreeSizeSum += subTreeSizeMap[child]
			}
			subTreeSizeMap[current] = childSubtreeSizeSum + 1
		}
	}
}

// propagateInterval propagates the new interval using a BFS traversal.
// Subtree intervals are recursively allocated according to subtree sizes and
// the allocation rule in splitWithExponentialBias. This method returns
// a list of reachabilityTreeNodes modified by it.
func (rtn *reachabilityTreeNode) propagateInterval(subTreeSizeMap map[*reachabilityTreeNode]uint64) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()
	queue := []*reachabilityTreeNode{rtn}
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) > 0 {
			sizes := make([]uint64, len(current.children))
			for i, child := range current.children {
				sizes[i] = subTreeSizeMap[child]
			}
			intervals, err := current.childIntervalAllocationRange().splitWithExponentialBias(sizes)
			if err != nil {
				return nil, err
			}
			for i, child := range current.children {
				childInterval := intervals[i]
				child.interval = childInterval
				queue = append(queue, child)
			}
		}

		modifiedTreeNodes[current] = struct{}{}
	}
	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) reindexIntervalsEarlierThanReindexRoot(
	reindexRoot *reachabilityTreeNode) (modifiedTreeNodes, error) {

	commonAncestor := rtn.findCommonAncestor(reindexRoot)
	commonAncestorChosenChild, err := commonAncestor.findAncestorAmongChildren(reindexRoot)
	if err != nil {
		return nil, err
	}

	if rtn.interval.end < commonAncestorChosenChild.interval.start {
		// rtn is in the subtree before the chosen child
		return rtn.reclaimIntervalBeforeChosenChild(commonAncestor, commonAncestorChosenChild, reindexRoot)
	}
	if commonAncestorChosenChild.interval.end < rtn.interval.start {
		// rtn is in the subtree after the chosen child
		return rtn.reclaimIntervalAfterChosenChild(commonAncestor, commonAncestorChosenChild, reindexRoot)
	}
	return nil, errors.Errorf("rtn is in the chosen child's subtree")
}

func (rtn *reachabilityTreeNode) reclaimIntervalBeforeChosenChild(
	commonAncestor *reachabilityTreeNode, commonAncestorChosenChild *reachabilityTreeNode, reindexRoot *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	current := commonAncestorChosenChild
	for !current.hasSlackIntervalBefore() {
		if current == reindexRoot {
			originalInterval := current.interval
			current.interval = newReachabilityInterval(current.interval.start+1, current.interval.end)

			modifiedNodes, err := current.countSubtreesAndPropagateInterval()
			if err != nil {
				return nil, err
			}
			modifiedTreeNodes.copyAllFrom(modifiedNodes)

			current.interval = originalInterval
			break
		}

		var err error
		current, err = current.findAncestorAmongChildren(reindexRoot)
		if err != nil {
			return nil, err
		}
	}

	for current != commonAncestor {
		current.interval = newReachabilityInterval(current.interval.start+1, current.interval.end)

		modifiedNodes, err := current.parent.reindexIntervalsBeforeChosenChild(current)
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)

		current = current.parent
	}

	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) reindexIntervalsBeforeChosenChild(chosenChild *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	childrenBeforeChosen, _, err := rtn.splitChildrenAroundChosenChild(chosenChild)
	if err != nil {
		return nil, err
	}

	childrenBeforeChosenSizes, childrenBeforeChosenSubtreeSizeMaps, childrenBeforeChosenSizesSum :=
		calcReachabilityTreeNodeSizes(childrenBeforeChosen)

	// Apply a tight interval
	newIntervalEnd := childrenBeforeChosen[len(childrenBeforeChosen)-1].interval.end + 1
	newInterval := newReachabilityInterval(newIntervalEnd-childrenBeforeChosenSizesSum+1, newIntervalEnd)
	intervals, err := newInterval.splitExact(childrenBeforeChosenSizes)
	if err != nil {
		return nil, err
	}
	for i, child := range childrenBeforeChosen {
		interval := intervals[i]
		subtreeSizeMap := childrenBeforeChosenSubtreeSizeMaps[i]
		child.interval = interval
		modifiedNodes, err := child.propagateInterval(subtreeSizeMap)
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)
	}

	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) reclaimIntervalAfterChosenChild(
	commonAncestor *reachabilityTreeNode, commonAncestorChosenChild *reachabilityTreeNode, reindexRoot *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	current := commonAncestorChosenChild
	for !current.hasSlackIntervalAfter() {
		if current == reindexRoot {
			originalInterval := current.interval
			current.interval = newReachabilityInterval(current.interval.start, current.interval.end-1)

			modifiedNodes, err := current.countSubtreesAndPropagateInterval()
			if err != nil {
				return nil, err
			}
			modifiedTreeNodes.copyAllFrom(modifiedNodes)

			current.interval = originalInterval
			break
		}

		var err error
		current, err = current.findAncestorAmongChildren(reindexRoot)
		if err != nil {
			return nil, err
		}
	}

	for current != commonAncestor {
		current.interval = newReachabilityInterval(current.interval.start, current.interval.end-1)

		modifiedNodes, err := current.parent.reindexIntervalsAfterChosenChild(current)
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)

		current = current.parent
	}

	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) reindexIntervalsAfterChosenChild(chosenChild *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	_, childrenAfterChosen, err := rtn.splitChildrenAroundChosenChild(chosenChild)
	if err != nil {
		return nil, err
	}

	childrenAfterChosenSizes, childrenAfterChosenSubtreeSizeMaps, childrenAfterChosenSizesSum :=
		calcReachabilityTreeNodeSizes(childrenAfterChosen)

	// Apply a tight interval
	newIntervalStart := childrenAfterChosen[0].interval.start - 1
	newInterval := newReachabilityInterval(newIntervalStart, newIntervalStart+childrenAfterChosenSizesSum-1)
	intervals, err := newInterval.splitExact(childrenAfterChosenSizes)
	if err != nil {
		return nil, err
	}
	for i, child := range childrenAfterChosen {
		interval := intervals[i]
		subtreeSizeMap := childrenAfterChosenSubtreeSizeMaps[i]
		child.interval = interval
		modifiedNodes, err := child.propagateInterval(subtreeSizeMap)
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)
	}

	return modifiedTreeNodes, nil
}

// isAncestorOf checks if this node is a reachability tree ancestor
// of the other node.
func (rtn *reachabilityTreeNode) isAncestorOf(other *reachabilityTreeNode) bool {
	return rtn.interval.isAncestorOf(other.interval)
}

// findCommonAncestor finds the most recent reachability tree ancestor
// common to both rtn and other.
func (rtn *reachabilityTreeNode) findCommonAncestor(other *reachabilityTreeNode) *reachabilityTreeNode {
	currentThis := rtn
	currentOther := other
	for {
		if currentThis.isAncestorOf(other) {
			return currentThis
		}
		if currentOther.isAncestorOf(rtn) {
			return currentOther
		}
		currentThis = currentThis.parent
		currentOther = currentOther.parent
	}
}

// String returns a string representation of a reachability tree node
// and its children.
func (rtn *reachabilityTreeNode) String() string {
	queue := []*reachabilityTreeNode{rtn}
	lines := []string{rtn.interval.String()}
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) == 0 {
			continue
		}

		line := ""
		for _, child := range current.children {
			line += child.interval.String()
			queue = append(queue, child)
		}
		lines = append([]string{line}, lines...)
	}
	return strings.Join(lines, "\n")
}

// futureCoveringBlockSet represents a collection of blocks in the future of
// a certain block. Once a block B is added to the DAG, every block A_i in
// B's selected parent anticone must register B in its futureCoveringBlockSet. This allows
// to relatively quickly (O(log(|futureCoveringBlockSet|))) query whether B
// is a descendent (is in the "future") of any block that previously
// registered it.
//
// Note that futureCoveringBlockSet is meant to be queried only if B is not
// a reachability tree descendant of the block in question, as reachability
// tree queries are always O(1).
//
// See insertBlock, isInFuture, and dag.isAncestorOf for further details.
type futureCoveringBlockSet []*futureCoveringBlock

// futureCoveringBlock represents a block in the future of some other block.
type futureCoveringBlock struct {
	blockNode *blockNode
	treeNode  *reachabilityTreeNode
}

func futureCoveringBlockSetFromReachabilityTreeNodes(nodes []*reachabilityTreeNode) futureCoveringBlockSet {
	futureCoveringBlocks := make([]*futureCoveringBlock, len(nodes))
	for i, node := range nodes {
		futureCoveringBlocks[i] = &futureCoveringBlock{
			blockNode: node.blockNode,
			treeNode:  node,
		}
	}
	return futureCoveringBlocks
}

// insertBlock inserts the given block into this futureCoveringBlockSet
// while keeping futureCoveringBlockSet ordered by interval.
// If a block B ∈ futureCoveringBlockSet exists such that its interval
// contains block's interval, block need not be added. If block's
// interval contains B's interval, it replaces it.
//
// Notes:
// * Intervals never intersect unless one contains the other
//   (this follows from the tree structure and the indexing rule).
// * Since futureCoveringBlockSet is kept ordered, a binary search can be
//   used for insertion/queries.
// * Although reindexing may change a block's interval, the
//   is-superset relation will by definition
//   be always preserved.
func (fb *futureCoveringBlockSet) insertBlock(block *futureCoveringBlock) {
	blockInterval := block.treeNode.interval
	i := fb.findIndex(block)
	if i > 0 {
		candidate := (*fb)[i-1]
		candidateInterval := candidate.treeNode.interval
		if candidateInterval.isAncestorOf(blockInterval) {
			// candidate is an ancestor of block, no need to insert
			return
		}
		if blockInterval.isAncestorOf(candidateInterval) {
			// block is an ancestor of candidate, and can thus replace it
			(*fb)[i-1] = block
			return
		}
	}

	// Insert block in the correct index to maintain futureCoveringBlockSet as
	// a sorted-by-interval list.
	// Note that i might be equal to len(futureCoveringBlockSet)
	left := (*fb)[:i]
	right := append([]*futureCoveringBlock{block}, (*fb)[i:]...)
	*fb = append(left, right...)
}

// isInFuture resolves whether the given block is in the subtree of
// any block in this futureCoveringBlockSet.
// See insertBlock method for the complementary insertion behavior.
//
// Like the insert method, this method also relies on the fact that
// futureCoveringBlockSet is kept ordered by interval to efficiently perform a
// binary search over futureCoveringBlockSet and answer the query in
// O(log(|futureCoveringBlockSet|)).
func (fb futureCoveringBlockSet) isInFuture(block *futureCoveringBlock) bool {
	i := fb.findIndex(block)
	if i == 0 {
		// No candidate to contain block
		return false
	}

	candidate := fb[i-1]
	return candidate.treeNode.isAncestorOf(block.treeNode)
}

// findIndex finds the index of the block with the maximum start that is below
// the given block.
func (fb futureCoveringBlockSet) findIndex(block *futureCoveringBlock) int {
	blockInterval := block.treeNode.interval
	end := blockInterval.end

	low := 0
	high := len(fb)
	for low < high {
		middle := (low + high) / 2
		middleInterval := fb[middle].treeNode.interval
		if end < middleInterval.start {
			high = middle
		} else {
			low = middle + 1
		}
	}
	return low
}

// String returns a string representation of the intervals in this futureCoveringBlockSet.
func (fb futureCoveringBlockSet) String() string {
	intervalsString := ""
	for _, block := range fb {
		intervalsString += block.treeNode.interval.String()
	}
	return intervalsString
}

func (dag *BlockDAG) updateReachability(node *blockNode, selectedParentAnticone []*blockNode) error {
	// Allocate a new reachability tree node
	newTreeNode := newReachabilityTreeNode(node)

	// If this is the genesis node, simply initialize it and return
	if node.isGenesis() {
		dag.reachabilityStore.setTreeNode(newTreeNode)
		dag.reachabilityTree.reindexRoot = newTreeNode
		return nil
	}

	// Insert the node into the selected parent's reachability tree
	selectedParentTreeNode, err := dag.reachabilityStore.treeNodeByBlockNode(node.selectedParent)
	if err != nil {
		return err
	}
	modifiedTreeNodes, err := selectedParentTreeNode.addChild(newTreeNode, dag.reachabilityTree.reindexRoot)
	if err != nil {
		return err
	}
	for modifiedTreeNode := range modifiedTreeNodes {
		dag.reachabilityStore.setTreeNode(modifiedTreeNode)
	}

	// Add the block to the futureCoveringSets of all the blocks
	// in the selected parent's anticone
	for _, current := range selectedParentAnticone {
		currentFutureCoveringSet, err := dag.reachabilityStore.futureCoveringSetByBlockNode(current)
		if err != nil {
			return err
		}
		currentFutureCoveringSet.insertBlock(&futureCoveringBlock{blockNode: node, treeNode: newTreeNode})
		err = dag.reachabilityStore.setFutureCoveringSet(current, currentFutureCoveringSet)
		if err != nil {
			return err
		}
	}

	// Update the reindex root.
	// Note that we check for blue score here in order to find out
	// whether the new node is going to be the virtual's selected
	// parent. We don't check node == virtual.selectedParent because
	// at this stage the virtual had not yet been updated.
	if node.blueScore > dag.SelectedTipBlueScore() {
		updateStartTime := time.Now()
		modifiedTreeNodes, err := dag.reachabilityTree.updateReindexRoot(newTreeNode)
		if err != nil {
			return err
		}
		if len(modifiedTreeNodes) > 0 {
			updateTimeElapsed := time.Since(updateStartTime)
			log.Debugf("Reachability reindex root updated to %s. "+
				"Modified %d tree nodes and took %dms.",
				dag.reachabilityTree.reindexRoot.blockNode.hash,
				len(modifiedTreeNodes), updateTimeElapsed.Milliseconds())
			for modifiedTreeNode := range modifiedTreeNodes {
				dag.reachabilityStore.setTreeNode(modifiedTreeNode)
			}
		}
	}

	return nil
}

type reachabilityTree struct {
	reindexRoot *reachabilityTreeNode
}

func newReachabilityTree(reindexRoot *reachabilityTreeNode) *reachabilityTree {
	return &reachabilityTree{reindexRoot: reindexRoot}
}

func (rt *reachabilityTree) updateReindexRoot(newTreeNode *reachabilityTreeNode) (modifiedTreeNodes, error) {
	modifiedTreeNodes := newModifiedTreeNodes()

	nextReindexRoot := rt.reindexRoot
	for {
		candidateReindexRoot, modifiedNodes, found, err := rt.maybeMoveReindexRoot(nextReindexRoot, newTreeNode)
		if err != nil {
			return nil, err
		}
		if !found {
			break
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)
		nextReindexRoot = candidateReindexRoot
	}

	rt.reindexRoot = nextReindexRoot
	return modifiedTreeNodes, nil
}

func (rt *reachabilityTree) maybeMoveReindexRoot(
	reindexRoot *reachabilityTreeNode, newTreeNode *reachabilityTreeNode) (
	newReindexRoot *reachabilityTreeNode, modifiedTreeNodes modifiedTreeNodes, found bool, err error) {

	if !reindexRoot.isAncestorOf(newTreeNode) {
		commonAncestor := reindexRoot.findCommonAncestor(newTreeNode)
		return commonAncestor, nil, true, nil
	}

	chosenReindexRootChild, err := reindexRoot.findAncestorAmongChildren(newTreeNode)
	if err != nil {
		return nil, nil, false, err
	}
	if newTreeNode.blockNode.blueScore-chosenReindexRootChild.blockNode.blueScore < reachabilityReindexWindow {
		return nil, nil, false, nil
	}
	modifiedTreeNodes, err = rt.concentrateIntervalAroundReindexRootChosenChild(reindexRoot, chosenReindexRootChild)
	if err != nil {
		return nil, nil, false, err
	}

	return chosenReindexRootChild, modifiedTreeNodes, true, nil
}

// findAncestorAmongChildren finds the reachability tree child
// of rtn that is the ancestor of node.
func (rtn *reachabilityTreeNode) findAncestorAmongChildren(node *reachabilityTreeNode) (*reachabilityTreeNode, error) {
	rootChildrenFutureCoveringSet := futureCoveringBlockSetFromReachabilityTreeNodes(rtn.children)
	i := rootChildrenFutureCoveringSet.findIndex(&futureCoveringBlock{blockNode: node.blockNode, treeNode: node})
	if i == 0 {
		return nil, errors.Errorf("rtn is not an ancestor of node")
	}

	return rootChildrenFutureCoveringSet[i-1].treeNode, nil
}

func (rt *reachabilityTree) concentrateIntervalAroundReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, chosenReindexRootChild *reachabilityTreeNode) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	reindexRootChildNodesBeforeChosen, reindexRootChildNodesAfterChosen, err :=
		reindexRoot.splitChildrenAroundChosenChild(chosenReindexRootChild)
	if err != nil {
		return nil, err
	}

	reindexRootChildNodesBeforeChosenSizesSum, modifiedNodes, err :=
		rt.tightenIntervalsBeforeReindexRootChosenChild(reindexRoot, reindexRootChildNodesBeforeChosen)
	if err != nil {
		return nil, err
	}
	modifiedTreeNodes.copyAllFrom(modifiedNodes)

	reindexRootChildNodesAfterChosenSizesSum, modifiedNodes, err :=
		rt.tightenIntervalsAfterReindexRootChosenChild(reindexRoot, reindexRootChildNodesAfterChosen)
	if err != nil {
		return nil, err
	}
	modifiedTreeNodes.copyAllFrom(modifiedNodes)

	modifiedNodes, err = rt.expandIntervalInReindexRootChosenChild(
		reindexRoot, chosenReindexRootChild, reindexRootChildNodesBeforeChosenSizesSum, reindexRootChildNodesAfterChosenSizesSum)
	if err != nil {
		return nil, err
	}
	modifiedTreeNodes.copyAllFrom(modifiedNodes)

	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) splitChildrenAroundChosenChild(chosenChild *reachabilityTreeNode) (
	nodesBeforeChosen []*reachabilityTreeNode, nodesAfterChosen []*reachabilityTreeNode, err error) {

	chosenIndex := -1
	for i, child := range rtn.children {
		if child == chosenChild {
			chosenIndex = i
			break
		}
	}
	if chosenIndex == -1 {
		return nil, nil, errors.Errorf("chosenChild not a child of rtn")
	}
	return rtn.children[:chosenIndex], rtn.children[chosenIndex+1:], nil
}

func (rt *reachabilityTree) tightenIntervalsBeforeReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, reindexRootChildNodesBeforeChosen []*reachabilityTreeNode) (
	reindexRootChildNodesBeforeChosenSizesSum uint64, modifiedTreeNodes modifiedTreeNodes, err error) {

	reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps, reindexRootChildNodesBeforeChosenSizesSum :=
		calcReachabilityTreeNodeSizes(reindexRootChildNodesBeforeChosen)

	reindexRootStart := reindexRoot.interval.start
	targetRangeBeforeReindexRootStart := reindexRootStart + reachabilityReindexSlack
	targetRangeBeforeReindexRootEnd := targetRangeBeforeReindexRootStart + reindexRootChildNodesBeforeChosenSizesSum - 1
	intervalBeforeReindexRootStart := newReachabilityInterval(targetRangeBeforeReindexRootStart, targetRangeBeforeReindexRootEnd)

	modifiedTreeNodes, err = rt.propagateChildIntervals(intervalBeforeReindexRootStart, reindexRootChildNodesBeforeChosen,
		reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps)
	if err != nil {
		return 0, nil, err
	}
	return reindexRootChildNodesBeforeChosenSizesSum, modifiedTreeNodes, nil
}

func (rt *reachabilityTree) tightenIntervalsAfterReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, reindexRootChildNodesAfterChosen []*reachabilityTreeNode) (
	reindexRootChildNodesAfterChosenSizesSum uint64, modifiedTreeNodes modifiedTreeNodes, err error) {

	reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps, reindexRootChildNodesAfterChosenSizesSum :=
		calcReachabilityTreeNodeSizes(reindexRootChildNodesAfterChosen)

	reindexRootEnd := reindexRoot.interval.end
	targetRangeAfterReindexRootEnd := reindexRootEnd - reachabilityReindexSlack
	targetRangeAfterReindexRootStart := targetRangeAfterReindexRootEnd - reindexRootChildNodesAfterChosenSizesSum
	intervalAfterReindexRootEnd := newReachabilityInterval(targetRangeAfterReindexRootStart, targetRangeAfterReindexRootEnd-1)

	modifiedTreeNodes, err = rt.propagateChildIntervals(intervalAfterReindexRootEnd, reindexRootChildNodesAfterChosen,
		reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps)
	if err != nil {
		return 0, nil, err
	}
	return reindexRootChildNodesAfterChosenSizesSum, modifiedTreeNodes, nil
}

func (rt *reachabilityTree) expandIntervalInReindexRootChosenChild(reindexRoot *reachabilityTreeNode,
	chosenReindexRootChild *reachabilityTreeNode, reindexRootChildNodesBeforeChosenSizesSum uint64,
	reindexRootChildNodesAfterChosenSizesSum uint64) (modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	reindexRootStart := reindexRoot.interval.start
	reindexRootEnd := reindexRoot.interval.end
	targetRangeForReindexRootChildStart :=
		reindexRootStart + reindexRootChildNodesBeforeChosenSizesSum + reachabilityReindexSlack
	targetRangeForReindexRootChildEnd :=
		reindexRootEnd - reindexRootChildNodesAfterChosenSizesSum - reachabilityReindexSlack - 1
	newReindexRootChildInterval := newReachabilityInterval(targetRangeForReindexRootChildStart, targetRangeForReindexRootChildEnd)

	if targetRangeForReindexRootChildStart > reindexRootStart || targetRangeForReindexRootChildEnd < reindexRootEnd {
		// New interval doesn't contain the previous one, propagation is required
		chosenReindexRootChild.interval = newReachabilityInterval(
			targetRangeForReindexRootChildStart+reachabilityReindexSlack,
			targetRangeForReindexRootChildEnd-reachabilityReindexSlack-1,
		)
		modifiedNodes, err := chosenReindexRootChild.countSubtreesAndPropagateInterval()
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)
	}

	chosenReindexRootChild.interval = newReindexRootChildInterval
	modifiedTreeNodes[chosenReindexRootChild] = struct{}{}
	return modifiedTreeNodes, nil
}

func (rtn *reachabilityTreeNode) countSubtreesAndPropagateInterval() (modifiedTreeNodes, error) {
	subtreeSizeMap := make(map[*reachabilityTreeNode]uint64)
	rtn.countSubtrees(subtreeSizeMap)
	return rtn.propagateInterval(subtreeSizeMap)
}

func calcReachabilityTreeNodeSizes(treeNodes []*reachabilityTreeNode) (
	sizes []uint64, subtreeSizeMaps []map[*reachabilityTreeNode]uint64, sum uint64) {

	sizes = make([]uint64, len(treeNodes))
	subtreeSizeMaps = make([]map[*reachabilityTreeNode]uint64, len(treeNodes))
	sum = 0
	for i, node := range treeNodes {
		subtreeSizeMap := make(map[*reachabilityTreeNode]uint64)
		node.countSubtrees(subtreeSizeMap)
		subtreeSize := subtreeSizeMap[node]
		sizes[i] = subtreeSize
		subtreeSizeMaps[i] = subtreeSizeMap
		sum += subtreeSize
	}
	return sizes, subtreeSizeMaps, sum
}

func (rt *reachabilityTree) propagateChildIntervals(interval *reachabilityInterval,
	childNodes []*reachabilityTreeNode, sizes []uint64, subtreeSizeMaps []map[*reachabilityTreeNode]uint64) (
	modifiedTreeNodes, error) {

	modifiedTreeNodes := newModifiedTreeNodes()

	childIntervalSizes, err := interval.splitExact(sizes)
	if err != nil {
		return nil, err
	}

	for i, child := range childNodes {
		childInterval := childIntervalSizes[i]
		child.interval = childInterval

		childSubtreeSizeMap := subtreeSizeMaps[i]
		modifiedNodes, err := child.propagateInterval(childSubtreeSizeMap)
		if err != nil {
			return nil, err
		}
		modifiedTreeNodes.copyAllFrom(modifiedNodes)
	}

	return modifiedTreeNodes, nil
}

// isAncestorOf returns true if this node is in the past of the other node
// in the DAG. The complexity of this method is O(log(|this.futureCoveringBlockSet|))
func (dag *BlockDAG) isAncestorOf(this *blockNode, other *blockNode) (bool, error) {
	// First, check if this node is a reachability tree ancestor of the
	// other node
	thisTreeNode, err := dag.reachabilityStore.treeNodeByBlockNode(this)
	if err != nil {
		return false, err
	}
	otherTreeNode, err := dag.reachabilityStore.treeNodeByBlockNode(other)
	if err != nil {
		return false, err
	}
	if thisTreeNode.isAncestorOf(otherTreeNode) {
		return true, nil
	}

	// Otherwise, use previously registered future blocks to complete the
	// reachability test
	thisFutureCoveringSet, err := dag.reachabilityStore.futureCoveringSetByBlockNode(this)
	if err != nil {
		return false, err
	}
	return thisFutureCoveringSet.isInFuture(&futureCoveringBlock{blockNode: other, treeNode: otherTreeNode}), nil
}
