package blockdag

import (
	"container/heap"
)

// baseHeap is an implementation for heap.Interface that sorts blocks by their height
type baseHeap []*blockNode

func (h baseHeap) Len() int      { return len(h) }
func (h baseHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *baseHeap) Push(x interface{}) {
	*h = append(*h, x.(*blockNode))
}

func (h *baseHeap) Pop() interface{} {
	oldHeap := *h
	oldLength := len(oldHeap)
	popped := oldHeap[oldLength-1]
	*h = oldHeap[0 : oldLength-1]
	return popped
}

// upHeap extends baseHeap to include Less operation that traverses from bottom to top
type upHeap struct{ baseHeap }

func (h upHeap) Less(i, j int) bool {
	return h.baseHeap[i].less(h.baseHeap[j])
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h downHeap) Less(i, j int) bool {
	return !h.baseHeap[i].less(h.baseHeap[j])
}

// blockHeap represents a mutable heap of Blocks, sorted by their height
type blockHeap struct {
	impl heap.Interface
}

// newDownHeap initializes and returns a new blockHeap
func newDownHeap() blockHeap {
	h := blockHeap{impl: &downHeap{}}
	heap.Init(h.impl)
	return h
}

// newUpHeap initializes and returns a new blockHeap
func newUpHeap() blockHeap {
	h := blockHeap{impl: &upHeap{}}
	heap.Init(h.impl)
	return h
}

// pop removes the block with lowest height from this heap and returns it
func (bh blockHeap) pop() *blockNode {
	return heap.Pop(bh.impl).(*blockNode)
}

// Push pushes the block onto the heap
func (bh blockHeap) Push(block *blockNode) {
	heap.Push(bh.impl, block)
}

// pushSet pushes a blockset to the heap.
func (bh blockHeap) pushSet(bs blockSet) {
	for block := range bs {
		heap.Push(bh.impl, block)
	}
}

// Len returns the length of this heap
func (bh blockHeap) Len() int {
	return bh.impl.Len()
}
