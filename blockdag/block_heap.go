package blockdag

import "container/heap"

// baseHeap is an implementation for heap.Interface that sorts blocks by their height
// baseHeap doesn't implement Less, because it is implemented by sub-classes, depending on direction
type baseHeap []*blockNode

func (h baseHeap) Len() int      { return len(h) }
func (h baseHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *baseHeap) Push(x interface{}) {
	*h = append(*h, x.(*blockNode))
}

func (h *baseHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// upHeap extends baseHeap to include Less operation that traverses from bottom to top
type upHeap struct{ baseHeap }

func (h upHeap) Less(i, j int) bool {
	return h.baseHeap[i].Height < h.baseHeap[j].Height || (h.baseHeap[i].Height == h.baseHeap[j].Height && h.baseHeap[i].ID < h.baseHeap[j].ID)
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h downHeap) Less(i, j int) bool {
	return h.baseHeap[i].Height > h.baseHeap[j].Height || (h.baseHeap[i].Height == h.baseHeap[j].Height && h.baseHeap[i].ID > h.baseHeap[j].ID)
}

// BlockHeap represents a mutable heap of Blocks, sorted by their height
type BlockHeap struct {
	impl heap.Interface
}

// HeapDirection represents the direction the heap traverses it's children
type blockHeapDirection bool

// HeapDirection possible values
const (
	blockHeapDirectionUp   blockHeapDirection = true
	blockHeapDirectionDown blockHeapDirection = false
)

// NewHeap initializes and returns a new BlockHeap
func newBlockHeap(direction HeapDirection) BlockHeap {
	var h BlockHeap
	if direction == HeapDirectionUp {
		h = BlockHeap{impl: &upHeap{}}
	} else {
		h = BlockHeap{impl: &downHeap{}}
	}
	heap.Init(h.impl)
	return h
}

// Pop removes the block with lowest height from this heap and returns it
func (bh BlockHeap) Pop() *Block {
	return heap.Pop(bh.impl).(*Block)
}

// Push pushes the block onto the heap
func (bh BlockHeap) Push(block *Block) {
	heap.Push(bh.impl, block)
}

// Len returns the length of this heap
func (bh BlockHeap) Len() int {
	return bh.impl.Len()
}
