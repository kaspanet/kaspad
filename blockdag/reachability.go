package blockdag

import "fmt"

type reachabilityInterval struct {
	start uint64
	end   uint64
}

type futureBlocks []*blockNode

// insertFutureBlock inserts the given block into this futureBlocks
// if required.
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

// isFutureBlock resolves whether the given block is in the future of
// any block in this futureBlocks.
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

// bisect finds the appropriate index for the givev block's reachability
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
