package blockdag

import (
	"container/heap"
)

type baseDelayedBlocksHeap []*delayedBlock

func (h baseDelayedBlocksHeap) Len() int {
	return len(h)
}
func (h baseDelayedBlocksHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *baseDelayedBlocksHeap) Push(x interface{}) {
	*h = append(*h, x.(*delayedBlock))
}

func (h *baseDelayedBlocksHeap) Pop() interface{} {
	oldHeap := *h
	oldLength := len(oldHeap)
	popped := oldHeap[oldLength-1]
	*h = oldHeap[0 : oldLength-1]
	return popped
}

func (h baseDelayedBlocksHeap) Less(i, j int) bool {
	return h[j].processTime.After(h[i].processTime)
}

type delayedBlocksHeap struct {
	baseDelayedBlocksHeap
	impl heap.Interface
}

// newdDelayedBlocksHeap initializes and returns a new delayedBlocksHeap
func newdDelayedBlocksHeap() delayedBlocksHeap {
	h := delayedBlocksHeap{impl: &baseDelayedBlocksHeap{}}
	heap.Init(h.impl)
	return h
}

// pop removes the block with lowest height from this heap and returns it
func (dbh delayedBlocksHeap) pop() *delayedBlock {
	return heap.Pop(dbh.impl).(*delayedBlock)
}

// Push pushes the block onto the heap
func (dbh delayedBlocksHeap) Push(block *delayedBlock) {
	heap.Push(dbh.impl, block)
}

// Len returns the length of this heap
func (dbh delayedBlocksHeap) Len() int {
	return dbh.impl.Len()
}

// Peek return the topmost block in the heap (block with the earliest process time)
func (dbh delayedBlocksHeap) Peek() *delayedBlock {
	if dbh.baseDelayedBlocksHeap.Len() > 0 {
		return dbh.baseDelayedBlocksHeap[dbh.baseDelayedBlocksHeap.Len()-1]
	}
	return nil
}
