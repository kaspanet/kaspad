package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
	"strings"
	"time"
)

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

	// remainingInterval is the not-yet allocated interval (within
	// this node's interval) awaiting new children
	remainingInterval *reachabilityInterval
}

func newReachabilityTreeNode(blockNode *blockNode) *reachabilityTreeNode {
	// Please see the comment above reachabilityTreeNode to understand why
	// we use these initial values.
	interval := newReachabilityInterval(1, math.MaxUint64-1)
	// We subtract 1 from the end of the remaining interval to prevent the node from allocating
	// the entire interval to its child, so its interval would *strictly* contain the interval of its child.
	remainingInterval := newReachabilityInterval(interval.start, interval.end-1)
	return &reachabilityTreeNode{blockNode: blockNode, interval: interval, remainingInterval: remainingInterval}
}

// addChild adds child to this tree node. If this node has no
// remaining interval to allocate, a reindexing is triggered.
// This method returns a list of reachabilityTreeNodes modified
// by it.
func (rtn *reachabilityTreeNode) addChild(child *reachabilityTreeNode) ([]*reachabilityTreeNode, error) {
	// Set the parent-child relationship
	rtn.children = append(rtn.children, child)
	child.parent = rtn

	// No allocation space left -- reindex
	if rtn.remainingInterval.size() == 0 {
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
	allocated, remaining, err := rtn.remainingInterval.splitInHalf()
	if err != nil {
		return nil, err
	}
	child.setInterval(allocated)
	rtn.remainingInterval = remaining
	return []*reachabilityTreeNode{rtn, child}, nil
}

// setInterval sets the reachability interval for this node.
func (rtn *reachabilityTreeNode) setInterval(interval *reachabilityInterval) {
	rtn.interval = interval

	// Reserve a single interval index for the current node. This
	// is necessary to ensure that ancestor intervals are strictly
	// supersets of any descendant intervals and not equal
	rtn.remainingInterval = newReachabilityInterval(interval.start, interval.end-1)
}

// reindexIntervals traverses the reachability subtree that's
// defined by this node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter. It does this by traversing down the reachability
// tree until it finds a node with a subreeSize that's greater than
// its interval size. See propagateInterval for further details.
// This method returns a list of reachabilityTreeNodes modified by it.
func (rtn *reachabilityTreeNode) reindexIntervals() ([]*reachabilityTreeNode, error) {
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
func (rtn *reachabilityTreeNode) propagateInterval(subTreeSizeMap map[*reachabilityTreeNode]uint64) ([]*reachabilityTreeNode, error) {
	// We set the interval to reset its remainingInterval, so we could reallocate it while reindexing.
	rtn.setInterval(rtn.interval)
	queue := []*reachabilityTreeNode{rtn}
	var modifiedNodes []*reachabilityTreeNode
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) > 0 {
			sizes := make([]uint64, len(current.children))
			for i, child := range current.children {
				sizes[i] = subTreeSizeMap[child]
			}
			intervals, err := current.remainingInterval.splitWithExponentialBias(sizes)
			if err != nil {
				return nil, err
			}
			for i, child := range current.children {
				childInterval := intervals[i]
				child.setInterval(childInterval)
				queue = append(queue, child)
			}

			// Empty up remaining interval
			current.remainingInterval.start = current.remainingInterval.end + 1
		}

		modifiedNodes = append(modifiedNodes, current)
	}
	return modifiedNodes, nil
}

// isAncestorOf checks if this node is a reachability tree ancestor
// of the other node.
func (rtn *reachabilityTreeNode) isAncestorOf(other *reachabilityTreeNode) bool {
	return rtn.interval.isAncestorOf(other.interval)
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
		return nil
	}

	// Insert the node into the selected parent's reachability tree
	selectedParentTreeNode, err := dag.reachabilityStore.treeNodeByBlockNode(node.selectedParent)
	if err != nil {
		return err
	}
	modifiedTreeNodes, err := selectedParentTreeNode.addChild(newTreeNode)
	if err != nil {
		return err
	}
	for _, modifiedTreeNode := range modifiedTreeNodes {
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
	return nil
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
