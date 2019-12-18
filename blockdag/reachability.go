package blockdag

import (
	"fmt"
	"github.com/pkg/errors"
	"math"
)

type reachabilityInterval struct {
	start uint64
	end   uint64
}

// size returns the size of this interval. Note that intervals are
// inclusive from both sides.
func (ri *reachabilityInterval) size() uint64 {
	return ri.end - ri.start + 1
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
	start := block.treeInterval.start
	i := fb.bisect(block)
	if i > 0 {
		candidate := (*fb)[i-1]
		end := block.treeInterval.end
		if candidate.treeInterval.start <= end && end <= candidate.treeInterval.end {
			// candidate is an ancestor of block, no need to insert
			return
		}
		if start <= candidate.treeInterval.end && candidate.treeInterval.end <= end {
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
	end := block.treeInterval.end
	return candidate.treeInterval.start <= end && end <= candidate.treeInterval.end
}

// bisect finds the appropriate index for the given block's reachability
// interval.
func (fb futureBlocks) bisect(block *blockNode) int {
	end := block.treeInterval.end

	low := 0
	high := len(fb)
	for low < high {
		middle := (low + high) / 2
		if end < fb[middle].treeInterval.start {
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
		intervalsString += fmt.Sprintf("[%d,%d]", block.treeInterval.start, block.treeInterval.end)
	}
	return intervalsString
}
