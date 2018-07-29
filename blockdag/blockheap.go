package blockdag

import "container/heap"

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

func (h baseHeap) Less(i, j int) bool {
	if h[i].height == h[j].height {
		return HashToBig(&h[i].hash).Cmp(HashToBig(&h[j].hash)) > 0
	}

	return h[i].height > h[j].height
}

// BlockHeap represents a mutable heap of Blocks, sorted by their height
type BlockHeap struct {
	impl heap.Interface
}

// NewHeap initializes and returns a new BlockHeap
func NewHeap() BlockHeap {
	h := BlockHeap{impl: &baseHeap{}}
	heap.Init(h.impl)
	return h
}

// Pop removes the block with lowest height from this heap and returns it
func (bh BlockHeap) Pop() *blockNode {
	return heap.Pop(bh.impl).(*blockNode)
}

// Push pushes the block onto the heap
func (bh BlockHeap) Push(block *blockNode) {
	heap.Push(bh.impl, block)
}

// Len returns the length of this heap
func (bh BlockHeap) Len() int {
	return bh.impl.Len()
}
