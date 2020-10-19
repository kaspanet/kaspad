package blocknode

import (
	"container/heap"
)

// baseHeap is an implementation for heap.Interface that sorts blocks by their height
type baseHeap []*Node

func (h baseHeap) Len() int      { return len(h) }
func (h baseHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *baseHeap) Push(x interface{}) {
	*h = append(*h, x.(*Node))
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
	return h.baseHeap[i].Less(h.baseHeap[j])
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h downHeap) Less(i, j int) bool {
	return !h.baseHeap[i].Less(h.baseHeap[j])
}

// BlockHeap represents a mutable heap of Blocks, sorted by their height
type BlockHeap struct {
	impl heap.Interface
}

// NewDownHeap initializes and returns a new BlockHeap
func NewDownHeap() BlockHeap {
	h := BlockHeap{impl: &downHeap{}}
	heap.Init(h.impl)
	return h
}

// NewUpHeap initializes and returns a new BlockHeap
func NewUpHeap() BlockHeap {
	h := BlockHeap{impl: &upHeap{}}
	heap.Init(h.impl)
	return h
}

// Pop removes the block with lowest height from this heap and returns it
func (bh BlockHeap) Pop() *Node {
	return heap.Pop(bh.impl).(*Node)
}

// Push pushes the block onto the heap
func (bh BlockHeap) Push(block *Node) {
	heap.Push(bh.impl, block)
}

// PushSet pushes a blockset to the heap.
func (bh BlockHeap) PushSet(bs Set) {
	for block := range bs {
		heap.Push(bh.impl, block)
	}
}

// Len returns the length of this heap
func (bh BlockHeap) Len() int {
	return bh.impl.Len()
}
