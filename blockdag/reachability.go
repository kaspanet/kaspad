package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dbaccess"
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

	// slackReachabilityIntervalForReclaiming is the slack interval to
	// reclaim during reachability reindexes earlier than the reindex root.
	// See reclaimIntervalBeforeChosenChild for further details. Note that
	// this is not a constant for testing purposes.
	slackReachabilityIntervalForReclaiming uint64 = 1
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

// contains returns true if ri contains other.
func (ri *reachabilityInterval) contains(other *reachabilityInterval) bool {
	return ri.start <= other.start && other.end <= ri.end
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

func (rtn *reachabilityTreeNode) intervalRangeForChildAllocation() *reachabilityInterval {
	// We subtract 1 from the end of the range to prevent the node from allocating
	// the entire interval to its child, so its interval would *strictly* contain the interval of its child.
	return newReachabilityInterval(rtn.interval.start, rtn.interval.end-1)
}

func (rtn *reachabilityTreeNode) remainingIntervalBefore() *reachabilityInterval {
	childRange := rtn.intervalRangeForChildAllocation()
	if len(rtn.children) == 0 {
		return childRange
	}
	return newReachabilityInterval(childRange.start, rtn.children[0].interval.start-1)
}

func (rtn *reachabilityTreeNode) remainingIntervalAfter() *reachabilityInterval {
	childRange := rtn.intervalRangeForChildAllocation()
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
func (rtn *reachabilityTreeNode) addChild(child *reachabilityTreeNode, reindexRoot *reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) error {

	remaining := rtn.remainingIntervalAfter()

	// Set the parent-child relationship
	rtn.children = append(rtn.children, child)
	child.parent = rtn

	// Handle rtn not being a descendant of the reindex root.
	// Note that we check rtn here instead of child because
	// at this point we don't yet know child's interval.
	if !reindexRoot.isAncestorOf(rtn) {
		reindexStartTime := time.Now()
		err := rtn.reindexIntervalsEarlierThanReindexRoot(reindexRoot, modifiedNodes)
		if err != nil {
			return err
		}
		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. This block is not a child of the current "+
			"reindex root %s. Modified %d tree nodes and took %dms.",
			rtn.blockNode.hash, reindexRoot.blockNode.hash,
			len(modifiedNodes), reindexTimeElapsed.Milliseconds())
		return nil
	}

	// No allocation space left -- reindex
	if remaining.size() == 0 {
		reindexStartTime := time.Now()
		err := rtn.reindexIntervals(modifiedNodes)
		if err != nil {
			return err
		}
		reindexTimeElapsed := time.Since(reindexStartTime)
		log.Debugf("Reachability reindex triggered for "+
			"block %s. Modified %d tree nodes and took %dms.",
			rtn.blockNode.hash, len(modifiedNodes), reindexTimeElapsed.Milliseconds())
		return nil
	}

	// Allocate from the remaining space
	allocated, _, err := remaining.splitInHalf()
	if err != nil {
		return err
	}
	child.interval = allocated
	modifiedNodes[rtn] = struct{}{}
	modifiedNodes[child] = struct{}{}
	return nil
}

// reindexIntervals traverses the reachability subtree that's
// defined by this node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter. It does this by traversing down the reachability
// tree until it finds a node with a subreeSize that's greater than
// its interval size. See propagateInterval for further details.
// This method returns a list of reachabilityTreeNodes modified by it.
func (rtn *reachabilityTreeNode) reindexIntervals(modifiedNodes modifiedTreeNodes) error {
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
			return errors.Errorf("missing tree " +
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
	return current.propagateInterval(subTreeSizeMap, modifiedNodes)
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
func (rtn *reachabilityTreeNode) propagateInterval(subTreeSizeMap map[*reachabilityTreeNode]uint64,
	modifiedNodes modifiedTreeNodes) error {

	queue := []*reachabilityTreeNode{rtn}
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) > 0 {
			sizes := make([]uint64, len(current.children))
			for i, child := range current.children {
				sizes[i] = subTreeSizeMap[child]
			}
			intervals, err := current.intervalRangeForChildAllocation().splitWithExponentialBias(sizes)
			if err != nil {
				return err
			}
			for i, child := range current.children {
				childInterval := intervals[i]
				child.interval = childInterval
				queue = append(queue, child)
			}
		}

		modifiedNodes[current] = struct{}{}
	}
	return nil
}

func (rtn *reachabilityTreeNode) reindexIntervalsEarlierThanReindexRoot(
	reindexRoot *reachabilityTreeNode, modifiedNodes modifiedTreeNodes) error {

	// Find the common ancestor for both rtn and the reindex root
	commonAncestor := rtn.findCommonAncestorWithReindexRoot(reindexRoot)

	// The chosen child is:
	// a. A reachability tree child of `commonAncestor`
	// b. A reachability tree ancestor of `reindexRoot`
	commonAncestorChosenChild, err := commonAncestor.findAncestorAmongChildren(reindexRoot)
	if err != nil {
		return err
	}

	if rtn.interval.end < commonAncestorChosenChild.interval.start {
		// rtn is in the subtree before the chosen child
		return rtn.reclaimIntervalBeforeChosenChild(commonAncestor,
			commonAncestorChosenChild, reindexRoot, modifiedNodes)
	}

	// rtn is either:
	// * in the subtree after the chosen child
	// * the common ancestor
	// In both cases we reclaim from the "after" subtree. In the
	// latter case this is arbitrary
	return rtn.reclaimIntervalAfterChosenChild(commonAncestor,
		commonAncestorChosenChild, reindexRoot, modifiedNodes)
}

func (rtn *reachabilityTreeNode) reclaimIntervalBeforeChosenChild(
	commonAncestor *reachabilityTreeNode, commonAncestorChosenChild *reachabilityTreeNode,
	reindexRoot *reachabilityTreeNode, modifiedNodes modifiedTreeNodes) error {

	current := commonAncestorChosenChild
	if !commonAncestorChosenChild.hasSlackIntervalBefore() {
		// The common ancestor ran out of slack before its chosen child.
		// Climb up the reachability tree toward the reindex root until
		// we find a node that has enough slack.
		for !current.hasSlackIntervalBefore() && current != reindexRoot {
			var err error
			current, err = current.findAncestorAmongChildren(reindexRoot)
			if err != nil {
				return err
			}
		}

		if current == reindexRoot {
			// "Deallocate" an interval of slackReachabilityIntervalForReclaiming
			// from this node. This is the interval that we'll use for the new
			// node.
			originalInterval := current.interval
			current.interval = newReachabilityInterval(
				current.interval.start+slackReachabilityIntervalForReclaiming,
				current.interval.end,
			)
			err := current.countSubtreesAndPropagateInterval(modifiedNodes)
			if err != nil {
				return err
			}
			current.interval = originalInterval
		}
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree before the
	// current node with an interval that is smaller by
	// slackReachabilityIntervalForReclaiming. This is to make room
	// for the new node.
	for current != commonAncestor {
		current.interval = newReachabilityInterval(
			current.interval.start+slackReachabilityIntervalForReclaiming,
			current.interval.end,
		)
		err := current.parent.reindexIntervalsBeforeNode(current, modifiedNodes)
		if err != nil {
			return err
		}
		current = current.parent
	}

	return nil
}

// reindexIntervalsBeforeNode applies a tight interval to the reachability
// subtree before `node`. Note that `node` itself is unaffected.
func (rtn *reachabilityTreeNode) reindexIntervalsBeforeNode(node *reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) error {

	childrenBeforeNode, _, err := rtn.splitChildrenAroundChild(node)
	if err != nil {
		return err
	}

	childrenBeforeNodeSizes, childrenBeforeNodeSubtreeSizeMaps, childrenBeforeNodeSizesSum :=
		calcReachabilityTreeNodeSizes(childrenBeforeNode)

	// Apply a tight interval
	newIntervalEnd := node.interval.start - 1
	newInterval := newReachabilityInterval(newIntervalEnd-childrenBeforeNodeSizesSum+1, newIntervalEnd)
	intervals, err := newInterval.splitExact(childrenBeforeNodeSizes)
	if err != nil {
		return err
	}
	return orderedTreeNodeSet(childrenBeforeNode).
		propagateIntervals(intervals, childrenBeforeNodeSubtreeSizeMaps, modifiedNodes)
}

func (rtn *reachabilityTreeNode) reclaimIntervalAfterChosenChild(
	commonAncestor *reachabilityTreeNode, commonAncestorChosenChild *reachabilityTreeNode,
	reindexRoot *reachabilityTreeNode, modifiedNodes modifiedTreeNodes) error {

	current := commonAncestorChosenChild
	if !commonAncestorChosenChild.hasSlackIntervalAfter() {
		// The common ancestor ran out of slack after its chosen child.
		// Climb up the reachability tree toward the reindex root until
		// we find a node that has enough slack.
		for !current.hasSlackIntervalAfter() && current != reindexRoot {
			var err error
			current, err = current.findAncestorAmongChildren(reindexRoot)
			if err != nil {
				return err
			}
		}

		if current == reindexRoot {
			// "Deallocate" an interval of slackReachabilityIntervalForReclaiming
			// from this node. This is the interval that we'll use for the new
			// node.
			originalInterval := current.interval
			current.interval = newReachabilityInterval(
				current.interval.start,
				current.interval.end-slackReachabilityIntervalForReclaiming,
			)
			modifiedNodes[current] = struct{}{}

			err := current.countSubtreesAndPropagateInterval(modifiedNodes)
			if err != nil {
				return err
			}
			current.interval = originalInterval
		}
	}

	// Go down the reachability tree towards the common ancestor.
	// On every hop we reindex the reachability subtree after the
	// current node with an interval that is smaller by
	// slackReachabilityIntervalForReclaiming. This is to make room
	// for the new node.
	for current != commonAncestor {
		current.interval = newReachabilityInterval(
			current.interval.start,
			current.interval.end-slackReachabilityIntervalForReclaiming,
		)
		modifiedNodes[current] = struct{}{}

		err := current.parent.reindexIntervalsAfterNode(current, modifiedNodes)
		if err != nil {
			return err
		}
		current = current.parent
	}

	return nil
}

// reindexIntervalsAfterNode applies a tight interval to the reachability
// subtree after `node`. Note that `node` itself is unaffected.
func (rtn *reachabilityTreeNode) reindexIntervalsAfterNode(node *reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) error {

	_, childrenAfterNode, err := rtn.splitChildrenAroundChild(node)
	if err != nil {
		return err
	}

	childrenAfterNodeSizes, childrenAfterNodeSubtreeSizeMaps, childrenAfterNodeSizesSum :=
		calcReachabilityTreeNodeSizes(childrenAfterNode)

	// Apply a tight interval
	newIntervalStart := node.interval.end + 1
	newInterval := newReachabilityInterval(newIntervalStart, newIntervalStart+childrenAfterNodeSizesSum-1)
	intervals, err := newInterval.splitExact(childrenAfterNodeSizes)
	if err != nil {
		return err
	}
	return orderedTreeNodeSet(childrenAfterNode).
		propagateIntervals(intervals, childrenAfterNodeSubtreeSizeMaps, modifiedNodes)
}

func (tns orderedTreeNodeSet) propagateIntervals(intervals []*reachabilityInterval,
	subtreeSizeMaps []map[*reachabilityTreeNode]uint64, modifiedNodes modifiedTreeNodes) error {

	for i, node := range tns {
		node.interval = intervals[i]
		subtreeSizeMap := subtreeSizeMaps[i]
		err := node.propagateInterval(subtreeSizeMap, modifiedNodes)
		if err != nil {
			return err
		}
	}
	return nil
}

// isAncestorOf checks if this node is a reachability tree ancestor
// of the other node. Note that we use the graph theory convention
// here which defines that rtn is also an ancestor of itself.
func (rtn *reachabilityTreeNode) isAncestorOf(other *reachabilityTreeNode) bool {
	return rtn.interval.contains(other.interval)
}

// findCommonAncestorWithReindexRoot finds the most recent reachability
// tree ancestor common to both rtn and the given reindex root. Note
// that we assume that almost always the chain between the reindex root
// and the common ancestor is longer than the chain between rtn and the
// common ancestor.
func (rtn *reachabilityTreeNode) findCommonAncestorWithReindexRoot(reindexRoot *reachabilityTreeNode) *reachabilityTreeNode {
	currentThis := rtn
	for {
		if currentThis.isAncestorOf(reindexRoot) {
			return currentThis
		}
		currentThis = currentThis.parent
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

// orderedTreeNodeSet is an ordered set of reachabilityTreeNodes
// Note that this type does not validate order validity. It's the
// responsibility of the caller to construct instances of this
// type properly.
type orderedTreeNodeSet []*reachabilityTreeNode

// futureCoveringTreeNodeSet represents a collection of blocks in the future of
// a certain block. Once a block B is added to the DAG, every block A_i in
// B's selected parent anticone must register B in its futureCoveringTreeNodeSet. This allows
// to relatively quickly (O(log(|futureCoveringTreeNodeSet|))) query whether B
// is a descendent (is in the "future") of any block that previously
// registered it.
//
// Note that futureCoveringTreeNodeSet is meant to be queried only if B is not
// a reachability tree descendant of the block in question, as reachability
// tree queries are always O(1).
//
// See insertNode, hasAncestorOf, and reachabilityTree.isInPast for further
// details.
type futureCoveringTreeNodeSet orderedTreeNodeSet

// insertNode inserts the given block into this futureCoveringTreeNodeSet
// while keeping futureCoveringTreeNodeSet ordered by interval.
// If a block B ∈ futureCoveringTreeNodeSet exists such that its interval
// contains block's interval, block need not be added. If block's
// interval contains B's interval, it replaces it.
//
// Notes:
// * Intervals never intersect unless one contains the other
//   (this follows from the tree structure and the indexing rule).
// * Since futureCoveringTreeNodeSet is kept ordered, a binary search can be
//   used for insertion/queries.
// * Although reindexing may change a block's interval, the
//   is-superset relation will by definition
//   be always preserved.
func (fb *futureCoveringTreeNodeSet) insertNode(node *reachabilityTreeNode) {
	ancestorIndex, ok := orderedTreeNodeSet(*fb).findAncestorIndexOfNode(node)
	if !ok {
		*fb = append([]*reachabilityTreeNode{node}, *fb...)
		return
	}

	candidate := (*fb)[ancestorIndex]
	if candidate.isAncestorOf(node) {
		// candidate is an ancestor of node, no need to insert
		return
	}
	if node.isAncestorOf(candidate) {
		// node is an ancestor of candidate, and can thus replace it
		(*fb)[ancestorIndex] = node
		return
	}

	// Insert node in the correct index to maintain futureCoveringTreeNodeSet as
	// a sorted-by-interval list.
	// Note that ancestorIndex might be equal to len(futureCoveringTreeNodeSet)
	left := (*fb)[:ancestorIndex+1]
	right := append([]*reachabilityTreeNode{node}, (*fb)[ancestorIndex+1:]...)
	*fb = append(left, right...)
}

// hasAncestorOf resolves whether the given node is in the subtree of
// any node in this futureCoveringTreeNodeSet.
// See insertNode method for the complementary insertion behavior.
//
// Like the insert method, this method also relies on the fact that
// futureCoveringTreeNodeSet is kept ordered by interval to efficiently perform a
// binary search over futureCoveringTreeNodeSet and answer the query in
// O(log(|futureCoveringTreeNodeSet|)).
func (fb futureCoveringTreeNodeSet) hasAncestorOf(node *reachabilityTreeNode) bool {
	ancestorIndex, ok := orderedTreeNodeSet(fb).findAncestorIndexOfNode(node)
	if !ok {
		// No candidate to contain node
		return false
	}

	candidate := fb[ancestorIndex]
	return candidate.isAncestorOf(node)
}

// findAncestorOfNode finds the reachability tree ancestor of `node`
// among the nodes in `tns`.
func (tns orderedTreeNodeSet) findAncestorOfNode(node *reachabilityTreeNode) (*reachabilityTreeNode, bool) {
	ancestorIndex, ok := tns.findAncestorIndexOfNode(node)
	if !ok {
		return nil, false
	}
	return tns[ancestorIndex], true
}

// findAncestorIndexOfNode finds the index of the reachability tree
// ancestor of `node` among the nodes in `tns`. It does so by finding
// the index of the block with the maximum start that is below the
// given block.
func (tns orderedTreeNodeSet) findAncestorIndexOfNode(node *reachabilityTreeNode) (int, bool) {
	blockInterval := node.interval
	end := blockInterval.end

	low := 0
	high := len(tns)
	for low < high {
		middle := (low + high) / 2
		middleInterval := tns[middle].interval
		if end < middleInterval.start {
			high = middle
		} else {
			low = middle + 1
		}
	}

	if low == 0 {
		return 0, false
	}
	return low - 1, true
}

// String returns a string representation of the intervals in this futureCoveringTreeNodeSet.
func (fb futureCoveringTreeNodeSet) String() string {
	intervalsString := ""
	for _, node := range fb {
		intervalsString += node.interval.String()
	}
	return intervalsString
}

func (rt *reachabilityTree) addBlock(node *blockNode, selectedParentAnticone []*blockNode) error {
	// Allocate a new reachability tree node
	newTreeNode := newReachabilityTreeNode(node)

	// If this is the genesis node, simply initialize it and return
	if node.isGenesis() {
		rt.store.setTreeNode(newTreeNode)
		rt.reindexRoot = newTreeNode
		return nil
	}

	// Insert the node into the selected parent's reachability tree
	selectedParentTreeNode, err := rt.store.treeNodeByBlockNode(node.selectedParent)
	if err != nil {
		return err
	}
	modifiedNodes := newModifiedTreeNodes()
	err = selectedParentTreeNode.addChild(newTreeNode, rt.reindexRoot, modifiedNodes)
	if err != nil {
		return err
	}
	for modifiedNode := range modifiedNodes {
		rt.store.setTreeNode(modifiedNode)
	}

	// Add the block to the futureCoveringSets of all the blocks
	// in the selected parent's anticone
	for _, current := range selectedParentAnticone {
		currentFutureCoveringSet, err := rt.store.futureCoveringSetByBlockNode(current)
		if err != nil {
			return err
		}
		currentFutureCoveringSet.insertNode(newTreeNode)
		err = rt.store.setFutureCoveringSet(current, currentFutureCoveringSet)
		if err != nil {
			return err
		}
	}

	// Update the reindex root.
	// Note that we check for blue score here in order to find out
	// whether the new node is going to be the virtual's selected
	// parent. We don't check node == virtual.selectedParent because
	// at this stage the virtual had not yet been updated.
	if node.blueScore > rt.dag.SelectedTipBlueScore() {
		updateStartTime := time.Now()
		modifiedNodes := newModifiedTreeNodes()
		err := rt.updateReindexRoot(newTreeNode, modifiedNodes)
		if err != nil {
			return err
		}
		if len(modifiedNodes) > 0 {
			updateTimeElapsed := time.Since(updateStartTime)
			log.Debugf("Reachability reindex root updated to %s. "+
				"Modified %d tree nodes and took %dms.",
				rt.reindexRoot.blockNode.hash,
				len(modifiedNodes), updateTimeElapsed.Milliseconds())
			for modifiedNode := range modifiedNodes {
				rt.store.setTreeNode(modifiedNode)
			}
		}
	}

	return nil
}

type reachabilityTree struct {
	dag         *BlockDAG
	store       *reachabilityStore
	reindexRoot *reachabilityTreeNode
}

func newReachabilityTree(dag *BlockDAG) *reachabilityTree {
	store := newReachabilityStore(dag)
	return &reachabilityTree{
		dag:         dag,
		store:       store,
		reindexRoot: nil,
	}
}

func (rt *reachabilityTree) init(dbContext dbaccess.Context) error {
	// Init the store
	err := rt.store.init(dbContext)
	if err != nil {
		return err
	}

	// Fetch the reindex root hash. If missing, use the genesis hash
	reindexRootHash, err := dbaccess.FetchReachabilityReindexRoot(dbContext)
	if err != nil {
		if !dbaccess.IsNotFoundError(err) {
			return err
		}
		reindexRootHash = rt.dag.dagParams.GenesisHash
	}

	// Init the reindex root
	reachabilityReindexRootNode, ok := rt.dag.index.LookupNode(reindexRootHash)
	if !ok {
		return errors.Errorf("reachability reindex root block %s "+
			"does not exist in the DAG", reindexRootHash)
	}
	rt.reindexRoot, err = rt.store.treeNodeByBlockNode(reachabilityReindexRootNode)
	if err != nil {
		return errors.Wrapf(err, "cannot set reachability reindex root")
	}

	return nil
}

func (rt *reachabilityTree) storeState(dbTx *dbaccess.TxContext) error {
	// Flush the store
	err := rt.dag.reachabilityTree.store.flushToDB(dbTx)
	if err != nil {
		return err
	}

	// Store the reindex root
	err = dbaccess.StoreReachabilityReindexRoot(dbTx, rt.reindexRoot.blockNode.hash)
	if err != nil {
		return err
	}

	return nil
}

func (rt *reachabilityTree) updateReindexRoot(newTreeNode *reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) error {

	nextReindexRoot := rt.reindexRoot
	for {
		candidateReindexRoot, found, err := rt.maybeMoveReindexRoot(nextReindexRoot, newTreeNode, modifiedNodes)
		if err != nil {
			return err
		}
		if !found {
			break
		}
		nextReindexRoot = candidateReindexRoot
	}

	rt.reindexRoot = nextReindexRoot
	return nil
}

func (rt *reachabilityTree) maybeMoveReindexRoot(
	reindexRoot *reachabilityTreeNode, newTreeNode *reachabilityTreeNode, modifiedNodes modifiedTreeNodes) (
	newReindexRoot *reachabilityTreeNode, found bool, err error) {

	if !reindexRoot.isAncestorOf(newTreeNode) {
		commonAncestor := newTreeNode.findCommonAncestorWithReindexRoot(reindexRoot)
		return commonAncestor, true, nil
	}

	reindexRootChosenChild, err := reindexRoot.findAncestorAmongChildren(newTreeNode)
	if err != nil {
		return nil, false, err
	}
	if newTreeNode.blockNode.blueScore-reindexRootChosenChild.blockNode.blueScore < reachabilityReindexWindow {
		return nil, false, nil
	}
	err = rt.concentrateIntervalAroundReindexRootChosenChild(reindexRoot, reindexRootChosenChild, modifiedNodes)
	if err != nil {
		return nil, false, err
	}

	return reindexRootChosenChild, true, nil
}

// findAncestorAmongChildren finds the reachability tree child
// of rtn that is the ancestor of node.
func (rtn *reachabilityTreeNode) findAncestorAmongChildren(node *reachabilityTreeNode) (*reachabilityTreeNode, error) {
	ancestor, ok := orderedTreeNodeSet(rtn.children).findAncestorOfNode(node)
	if !ok {
		return nil, errors.Errorf("rtn is not an ancestor of node")
	}

	return ancestor, nil
}

func (rt *reachabilityTree) concentrateIntervalAroundReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, reindexRootChosenChild *reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) error {

	reindexRootChildNodesBeforeChosen, reindexRootChildNodesAfterChosen, err :=
		reindexRoot.splitChildrenAroundChild(reindexRootChosenChild)
	if err != nil {
		return err
	}

	reindexRootChildNodesBeforeChosenSizesSum, err :=
		rt.tightenIntervalsBeforeReindexRootChosenChild(reindexRoot, reindexRootChildNodesBeforeChosen, modifiedNodes)
	if err != nil {
		return err
	}

	reindexRootChildNodesAfterChosenSizesSum, err :=
		rt.tightenIntervalsAfterReindexRootChosenChild(reindexRoot, reindexRootChildNodesAfterChosen, modifiedNodes)
	if err != nil {
		return err
	}

	err = rt.expandIntervalInReindexRootChosenChild(reindexRoot, reindexRootChosenChild,
		reindexRootChildNodesBeforeChosenSizesSum, reindexRootChildNodesAfterChosenSizesSum, modifiedNodes)
	if err != nil {
		return err
	}

	return nil
}

// splitChildrenAroundChild splits `rtn` into two slices: the nodes that are before
// `child` and the nodes that are after.
func (rtn *reachabilityTreeNode) splitChildrenAroundChild(child *reachabilityTreeNode) (
	nodesBeforeChild []*reachabilityTreeNode, nodesAfterChild []*reachabilityTreeNode, err error) {

	for i, candidateChild := range rtn.children {
		if candidateChild == child {
			return rtn.children[:i], rtn.children[i+1:], nil
		}
	}
	return nil, nil, errors.Errorf("child not a child of rtn")
}

func (rt *reachabilityTree) tightenIntervalsBeforeReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, reindexRootChildNodesBeforeChosen []*reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) (reindexRootChildNodesBeforeChosenSizesSum uint64, err error) {

	reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps, reindexRootChildNodesBeforeChosenSizesSum :=
		calcReachabilityTreeNodeSizes(reindexRootChildNodesBeforeChosen)

	intervalBeforeReindexRootStart := newReachabilityInterval(
		reindexRoot.interval.start+reachabilityReindexSlack,
		reindexRoot.interval.start+reachabilityReindexSlack+reindexRootChildNodesBeforeChosenSizesSum-1,
	)

	err = rt.propagateChildIntervals(intervalBeforeReindexRootStart, reindexRootChildNodesBeforeChosen,
		reindexRootChildNodesBeforeChosenSizes, reindexRootChildNodesBeforeChosenSubtreeSizeMaps, modifiedNodes)
	if err != nil {
		return 0, err
	}
	return reindexRootChildNodesBeforeChosenSizesSum, nil
}

func (rt *reachabilityTree) tightenIntervalsAfterReindexRootChosenChild(
	reindexRoot *reachabilityTreeNode, reindexRootChildNodesAfterChosen []*reachabilityTreeNode,
	modifiedNodes modifiedTreeNodes) (reindexRootChildNodesAfterChosenSizesSum uint64, err error) {

	reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps, reindexRootChildNodesAfterChosenSizesSum :=
		calcReachabilityTreeNodeSizes(reindexRootChildNodesAfterChosen)

	intervalAfterReindexRootEnd := newReachabilityInterval(
		reindexRoot.interval.end-reachabilityReindexSlack-reindexRootChildNodesAfterChosenSizesSum,
		reindexRoot.interval.end-reachabilityReindexSlack-1,
	)

	err = rt.propagateChildIntervals(intervalAfterReindexRootEnd, reindexRootChildNodesAfterChosen,
		reindexRootChildNodesAfterChosenSizes, reindexRootChildNodesAfterChosenSubtreeSizeMaps, modifiedNodes)
	if err != nil {
		return 0, err
	}
	return reindexRootChildNodesAfterChosenSizesSum, nil
}

func (rt *reachabilityTree) expandIntervalInReindexRootChosenChild(reindexRoot *reachabilityTreeNode,
	reindexRootChosenChild *reachabilityTreeNode, reindexRootChildNodesBeforeChosenSizesSum uint64,
	reindexRootChildNodesAfterChosenSizesSum uint64, modifiedNodes modifiedTreeNodes) error {

	newReindexRootChildInterval := newReachabilityInterval(
		reindexRoot.interval.start+reindexRootChildNodesBeforeChosenSizesSum+reachabilityReindexSlack,
		reindexRoot.interval.end-reindexRootChildNodesAfterChosenSizesSum-reachabilityReindexSlack-1,
	)

	if !newReindexRootChildInterval.contains(reindexRootChosenChild.interval) {
		// New interval doesn't contain the previous one, propagation is required

		// We assign slack on both sides as an optimization. Were we to
		// assign a tight interval, the next time the reindex root moves we
		// would need to propagate intervals again. That is to say, When we
		// DO allocate slack, next time
		// expandIntervalInReindexRootChosenChild is called (next time the
		// reindex root moves), newReindexRootChildInterval is likely to
		// contain reindexRootChosenChild.interval.
		reindexRootChosenChild.interval = newReachabilityInterval(
			newReindexRootChildInterval.start+reachabilityReindexSlack,
			newReindexRootChildInterval.end-reachabilityReindexSlack,
		)
		err := reindexRootChosenChild.countSubtreesAndPropagateInterval(modifiedNodes)
		if err != nil {
			return err
		}
	}

	reindexRootChosenChild.interval = newReindexRootChildInterval
	modifiedNodes[reindexRootChosenChild] = struct{}{}
	return nil
}

func (rtn *reachabilityTreeNode) countSubtreesAndPropagateInterval(modifiedNodes modifiedTreeNodes) error {
	subtreeSizeMap := make(map[*reachabilityTreeNode]uint64)
	rtn.countSubtrees(subtreeSizeMap)
	return rtn.propagateInterval(subtreeSizeMap, modifiedNodes)
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
	childNodes []*reachabilityTreeNode, sizes []uint64, subtreeSizeMaps []map[*reachabilityTreeNode]uint64,
	modifiedNodes modifiedTreeNodes) error {

	childIntervalSizes, err := interval.splitExact(sizes)
	if err != nil {
		return err
	}

	for i, child := range childNodes {
		childInterval := childIntervalSizes[i]
		child.interval = childInterval

		childSubtreeSizeMap := subtreeSizeMaps[i]
		err := child.propagateInterval(childSubtreeSizeMap, modifiedNodes)
		if err != nil {
			return err
		}
	}

	return nil
}

// isInPast returns true if `this` is in the past (exclusive) of `other`
// in the DAG.
// The complexity of this method is O(log(|this.futureCoveringTreeNodeSet|))
func (rt *reachabilityTree) isInPast(this *blockNode, other *blockNode) (bool, error) {
	// By definition, a node is not in the past of itself.
	if this == other {
		return false, nil
	}

	// Check if this node is a reachability tree ancestor of the
	// other node
	isReachabilityTreeAncestor, err := rt.isReachabilityTreeAncestorOf(this, other)
	if err != nil {
		return false, err
	}
	if isReachabilityTreeAncestor {
		return true, nil
	}

	// Otherwise, use previously registered future blocks to complete the
	// reachability test
	thisFutureCoveringSet, err := rt.store.futureCoveringSetByBlockNode(this)
	if err != nil {
		return false, err
	}
	otherTreeNode, err := rt.store.treeNodeByBlockNode(other)
	if err != nil {
		return false, err
	}
	return thisFutureCoveringSet.hasAncestorOf(otherTreeNode), nil
}

// isReachabilityTreeAncestorOf returns whether `this` is in the selected parent chain of `other`.
func (rt *reachabilityTree) isReachabilityTreeAncestorOf(this *blockNode, other *blockNode) (bool, error) {
	thisTreeNode, err := rt.store.treeNodeByBlockNode(this)
	if err != nil {
		return false, err
	}
	otherTreeNode, err := rt.store.treeNodeByBlockNode(other)
	if err != nil {
		return false, err
	}
	return thisTreeNode.isAncestorOf(otherTreeNode), nil
}
