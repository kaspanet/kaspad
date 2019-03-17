package blockdag

import (
	"container/heap"

	"github.com/daglabs/btcd/dagconfig/daghash"
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
	if h.baseHeap[i].height == h.baseHeap[j].height {
		return daghash.HashToBig(&h.baseHeap[i].hash).Cmp(daghash.HashToBig(&h.baseHeap[j].hash)) < 0
	}

	return h.baseHeap[i].height < h.baseHeap[j].height
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h downHeap) Less(i, j int) bool {
	if h.baseHeap[i].height == h.baseHeap[j].height {
		return daghash.HashToBig(&h.baseHeap[i].hash).Cmp(daghash.HashToBig(&h.baseHeap[j].hash)) > 0
	}

	return h.baseHeap[i].height > h.baseHeap[j].height
}

// BlockHeap represents a mutable heap of Blocks, sorted by their height
type BlockHeap struct {
	impl heap.Interface
}

// HeapDirection represents the direction the heap traverses it's children
type HeapDirection bool

// HeapDirection possible values
const (
	HeapDirectionUp   HeapDirection = true
	HeapDirectionDown HeapDirection = false
)

// NewHeap initializes and returns a new BlockHeap
func NewHeap(direction HeapDirection) BlockHeap {
	var h BlockHeap
	if direction == HeapDirectionUp {
		h = BlockHeap{impl: &upHeap{}}
	} else {
		h = BlockHeap{impl: &downHeap{}}
	}
	heap.Init(h.impl)
	return h
}

// pop removes the block with lowest height from this heap and returns it
func (bh BlockHeap) pop() *blockNode {
	return heap.Pop(bh.impl).(*blockNode)
}

// Push pushes the block onto the heap
func (bh BlockHeap) Push(block *blockNode) {
	heap.Push(bh.impl, block)
}

// pushMany pushes a slice of blocks to the heap.
func (bh BlockHeap) pushMany(blocks []*blockNode) {
	for _, block := range blocks {
		heap.Push(bh.impl, block)
	}
}

// Len returns the length of this heap
func (bh BlockHeap) Len() int {
	return bh.impl.Len()
}
