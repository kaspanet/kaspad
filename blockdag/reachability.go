package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
)

type reachabilityTreeNode struct {
	children []*reachabilityTreeNode
	parent   *reachabilityTreeNode

	// interval is the index interval containing all intervals of
	// blocks in this node's subtree
	interval reachabilityInterval

	// remainingInterval is the not-yet allocated interval (within
	// this node's interval) awaiting new children
	remainingInterval reachabilityInterval

	// subtreeSize is a helper field used only during reindexing
	// (expected to be 0 any other time).
	// See countSubtreesUp for further details.
	subtreeSize uint64
}

// addTreeChild adds child to this tree node. If this node has no
// remaining interval to allocate, a reindexing is triggered.
func (rtn *reachabilityTreeNode) addTreeChild(child *reachabilityTreeNode) error {
	// Set the parent-child relationship
	rtn.children = append(rtn.children, child)
	child.parent = rtn

	allocated, remaining, err := rtn.remainingInterval.splitInHalf()
	if err != nil {
		return err
	}

	// No allocation space left -- reindex
	if allocated.start > allocated.end {
		return rtn.reindexTreeIntervals()
	}

	// Allocate from the remaining space
	child.setTreeInterval(allocated)
	rtn.remainingInterval = *remaining
	return nil
}

// setTreeInterval sets the reachability interval for this node.
func (rtn *reachabilityTreeNode) setTreeInterval(interval *reachabilityInterval) {
	rtn.interval = *interval

	// Reserve a single interval index for the current node. This
	// is necessary to ensure that ancestor intervals are strictly
	// supersets of any descendant intervals and not equal
	rtn.remainingInterval = reachabilityInterval{start: interval.start, end: interval.end - 1}
}

// reindexTreeInterval traverses the reachability subtree that's
// defined by this node and reallocates reachability interval space
// such that another reindexing is unlikely to occur shortly
// thereafter.
func (rtn *reachabilityTreeNode) reindexTreeIntervals() error {
	current := rtn

	// Initial interval and subtree sizes
	intervalSize := current.interval.size()
	subtreeSize := current.countSubtreesUp()

	// Find the first ancestor that has sufficient interval space
	for intervalSize < subtreeSize {
		if current.parent == nil {
			// If we ended up here it means that there are more
			// than 2^64 blocks inside the finality window,
			// something that shouldn't ever happen.
			return errors.Errorf("missing tree parent")
		}
		current = current.parent
		intervalSize = current.interval.size()
		subtreeSize = current.countSubtreesUp()
	}

	// Apply the interval down the subtree
	return current.applyIntervalDown(&current.interval)
}

// This method counts the size of each subtree under this node.
// The method outcome is exactly equal to the following recursive
// implementation:
//
// func (rtn *reachabilityTreeNode) countSubtreesUp() uint64 {
//     subtreeSize := uint64(0)
//     for _, child := range rtn.children {
//         subtreeSize += child.countSubtreesUp()
//     }
//     return subtreeSize
// }
//
// However we are expecting (linearly) deep trees, and so a
// recursive stack-based approach is inefficient and will hit
// recursion limits. Instead, the same logic was implemented
// using a (queue-based) BFS method. At a high level, the
// algorithm uses BFS for reaching all leafs and pushes
// intermediate updates from leafs via parent chains until all
// size information is gathered at the root of the operation
// (i.e. at self).
//
// Note the role of the subtreeSize field in the algorithm.
// For each block B this field is initialized to 0. The field
// has two possible states:
// * rtn.subtreeSize > |B.children|:
//	 this indicated that B's subtree size is already known and
//	 calculated.
// * rtn.subtreeSize <= |B.children|:
//   we are still in the counting stage of tracking who of B's
//   children has already calculated its subtree size. This way,
//   once B.subtree_size = |B.children| we know we can pull
//   subtree sizes from children and continue pushing the
// 	 readiness signal further up
func (rtn *reachabilityTreeNode) countSubtreesUp() uint64 {
	queue := []*reachabilityTreeNode{rtn}
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) == 0 {
			// We reached a leaf
			current.subtreeSize = 1
		}
		if current.subtreeSize <= uint64(len(current.children)) {
			// We haven't yet calculated the subtree size of
			// the current node. Add all its children to the
			// queue
			for _, child := range current.children {
				queue = append(queue, child)
			}
			continue
		}

		// We reached a leaf or a pre-calculated subtree.
		// Push information up
		for current != rtn {
			current = current.parent
			current.subtreeSize++
			if current.subtreeSize != uint64(len(current.children)) {
				// Not all subtrees of the current node are ready
				break
			}
			// All subtrees of current have reported readiness.
			// Count actual subtree size and continue pushing up.
			childSubtreeSizeSum := uint64(0)
			for _, child := range current.children {
				childSubtreeSizeSum += child.subtreeSize
			}
			current.subtreeSize = childSubtreeSizeSum + 1
		}
	}
	return rtn.subtreeSize
}

// applyIntervalDown applies new intervals using a BFS traversal.
// The intervals are allocated according to subtree sizes and the
// 'split' allocation rule (see the split() method for further
// details)
func (rtn *reachabilityTreeNode) applyIntervalDown(interval *reachabilityInterval) error {
	rtn.setTreeInterval(interval)

	queue := []*reachabilityTreeNode{rtn}
	for len(queue) > 0 {
		var current *reachabilityTreeNode
		current, queue = queue[0], queue[1:]
		if len(current.children) > 0 {
			sizes := make([]uint64, len(current.children))
			for i, child := range current.children {
				sizes[i] = child.subtreeSize
			}
			intervals, err := current.remainingInterval.split(sizes)
			if err != nil {
				return err
			}
			for i, child := range current.children {
				childInterval := intervals[i]
				child.setTreeInterval(childInterval)
				queue = append(queue, child)
			}

			// Empty up remaining interval
			current.remainingInterval.start = current.remainingInterval.end + 1
		}

		// Cleanup temp info for future reindexing
		current.subtreeSize = 0
	}
	return nil
}

type reachabilityInterval struct {
	start uint64
	end   uint64
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
// contains the given fraction of the original interval's capacity.
func (ri *reachabilityInterval) splitFraction(fraction float64) (
	left *reachabilityInterval, right *reachabilityInterval, err error) {
	if fraction < 0 || fraction > 1 {
		return nil, nil, errors.Errorf("fraction must be between 0 and 1")
	}
	if ri.end < ri.start {
		return ri, ri, nil
	}

	allocationSize := uint64(math.Ceil(float64(ri.size()) * fraction))
	left = &reachabilityInterval{start: ri.start, end: ri.start + allocationSize - 1}
	right = &reachabilityInterval{start: ri.start + allocationSize, end: ri.end}
	return left, right, nil
}

// splitExact splits this interval to exactly |sizes| parts where
// |part_i| = sizes[i].	This method expects sum(sizes) to be exactly
// equal to the interval's capacity.
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
		intervals[i] = &reachabilityInterval{start: start, end: start + size - 1}
		start += size
	}
	return intervals, nil
}

// split splits this interval to |sizes| parts by some allocation
// rule. This method expects sum(sizes)	to be smaller or equal to
// the interval's capacity. Every part_i is allocated at least
// sizes[i] capacity. The remaining budget is split by an
// exponential rule described below.
//
// This rule follows the GHOSTDAG protocol behavior where the child
// with the largest subtree is expected to dominate the competition
// for new blocks and thus grow the most. However, we may need to
// add slack for non-largest subtrees in order to make CPU reindexing
// attacks unworthy.
func (ri *reachabilityInterval) split(sizes []uint64) ([]*reachabilityInterval, error) {
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

	// Give exponentially proportional allocation:
	//   f_i = 2^x_i / sum(2^x_j)
	// In the code below the above equation is divided by 2^max(x_i)
	// to avoid exploding numbers.
	maxSize := uint64(0)
	for _, size := range sizes {
		if size > maxSize {
			maxSize = size
		}
	}
	fractions := make([]float64, len(sizes))
	for i, size := range sizes {
		fractions[i] = math.Pow(2, float64(size-maxSize))
	}
	fractionsSum := float64(0)
	for _, fraction := range fractions {
		fractionsSum += fraction
	}
	for i, fraction := range fractions {
		fractions[i] = fraction / fractionsSum
	}

	// Add a fractional bias to every size in the given sizes
	totalBias := float64(intervalSize - sizesSum)
	remainingBias := totalBias
	biasedSizes := make([]uint64, len(sizes))
	for i, fraction := range fractions {
		var bias float64
		if i == len(fractions)-1 {
			bias = remainingBias
		} else {
			bias = math.Min(math.Round(totalBias*fraction), remainingBias)
		}
		biasedSizes[i] = sizes[i] + uint64(bias)
		remainingBias -= bias
	}
	return ri.splitExact(biasedSizes)
}

// isAncestorOf checks if this interval's node is a reachability tree
// ancestor of the other interval's node.
func (ri *reachabilityInterval) isAncestorOf(other *reachabilityInterval) bool {
	return ri.start <= other.end && other.end <= ri.end
}

type futureBlocks []*blockNode

// insertFutureBlock inserts the given block into this futureBlocks
// while keeping futureBlocks ordered by interval.
// If a block B ∈ futureBlocks exists s.t. its interval contains
// block's interval, block need not be added. If block's interval
// contains B's interval, it replaces it.
//
// Notes:
// * Intervals never intersect unless one contains the other
//   (this follows from the tree structure and the indexing rule).
// * Since futureBlocks is kept ordered, a binary search can be
//   used for insertion/queries.
// * Although reindexing may change a block's interval, the
//   is-superset relation will by definition
// be always preserved.
func (fb *futureBlocks) insertFutureBlock(block *blockNode) {
	blockInterval := block.reachabilityTreeNode.interval
	i := fb.bisect(block)
	if i > 0 {
		candidate := (*fb)[i-1]
		candidateInterval := candidate.reachabilityTreeNode.interval
		if candidateInterval.isAncestorOf(&blockInterval) {
			// candidate is an ancestor of block, no need to insert
			return
		}
		if blockInterval.isAncestorOf(&candidateInterval) {
			// block is an ancestor of candidate, and can thus replace it
			(*fb)[i-1] = block
			return
		}
	}

	// Insert block in the correct index to maintain futureBlocks as
	// a sorted-by-interval list.
	// Note that i might be equal to len(futureBlocks)
	left := (*fb)[:i]
	right := append([]*blockNode{block}, (*fb)[i:]...)
	*fb = append(left, right...)
}

// isFutureBlock resolves whether the given block is in the subtree of
// any block in this futureBlocks.
// See insertFutureBlock method for the complementary insertion behavior.
//
// Like the insert method, this method also relies on the fact that
// futureBlocks is kept ordered by interval to efficiently perform a
// binary search over futureBlocks and answer the query in
// O(log(|future_blocks|)).
func (fb futureBlocks) isFutureBlock(block *blockNode) bool {
	i := fb.bisect(block)
	if i == 0 {
		// No candidate to contain block
		return false
	}

	candidate := fb[i-1]
	blockInterval := block.reachabilityTreeNode.interval
	candidateInterval := candidate.reachabilityTreeNode.interval
	return candidateInterval.isAncestorOf(&blockInterval)
}

// bisect finds the appropriate index for the given block's reachability
// interval.
func (fb futureBlocks) bisect(block *blockNode) int {
	blockInterval := block.reachabilityTreeNode.interval
	end := blockInterval.end

	low := 0
	high := len(fb)
	for low < high {
		middle := (low + high) / 2
		middleInterval := fb[middle].reachabilityTreeNode.interval
		if end < middleInterval.start {
			high = middle
		} else {
			low = middle + 1
		}
	}
	return low
}

// String returns a string representation of the intervals in this futureBlocks.
func (fb futureBlocks) String() string {
	intervalsString := ""
	for _, block := range fb {
		blockInterval := block.reachabilityTreeNode.interval
		intervalsString += fmt.Sprintf("[%d,%d]", blockInterval.start, blockInterval.end)
	}
	return intervalsString
}
