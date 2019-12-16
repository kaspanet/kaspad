package blockdag

type reachabilityInterval struct {
	start uint64
	end   uint64
}

func bisect(futureBlocks []*blockNode, intervalEnd uint64) int {
	low := 0
	high := len(futureBlocks)
	for low < high {
		middle := (low + high) / 2
		if intervalEnd < futureBlocks[middle].interval.start {
			high = middle
		} else {
			low = middle + 1
		}
	}
	return low
}

func insertFutureBlock(futureBlocks []*blockNode, block *blockNode) []*blockNode {
	start := block.interval.start
	end := block.interval.end
	i := bisect(futureBlocks, end)
	if i > 0 {
		candidate := futureBlocks[i-1]
		if candidate.interval.start <= end && end <= candidate.interval.end {
			// candidate is an ancestor of block, no need to insert
			return futureBlocks
		}
		if start <= candidate.interval.end && candidate.interval.end <= end {
			// block is ancestor of candidate, and can thus replace it
			futureBlocks[i-1] = block
			return futureBlocks
		}
	}
	// Insert block in the correct index to maintain future_blocks as a sorted-by-interval list
	// (note that i might be equal to len(future_blocks))
	//futureBlocks.insert(i, block)
	return futureBlocks
}
