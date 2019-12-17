package blockdag

import "fmt"

type reachabilityInterval struct {
	start uint64
	end   uint64
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
	start := block.interval.start
	i := fb.bisect(block)
	if i > 0 {
		candidate := (*fb)[i-1]
		end := block.interval.end
		if candidate.interval.start <= end && end <= candidate.interval.end {
			// candidate is an ancestor of block, no need to insert
			return
		}
		if start <= candidate.interval.end && candidate.interval.end <= end {
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
	end := block.interval.end
	return candidate.interval.start <= end && end <= candidate.interval.end
}

// bisect finds the appropriate index for the given block's reachability
// interval.
func (fb futureBlocks) bisect(block *blockNode) int {
	end := block.interval.end

	low := 0
	high := len(fb)
	for low < high {
		middle := (low + high) / 2
		if end < fb[middle].interval.start {
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
		intervalsString += fmt.Sprintf("[%d,%d]", block.interval.start, block.interval.end)
	}
	return intervalsString
}
