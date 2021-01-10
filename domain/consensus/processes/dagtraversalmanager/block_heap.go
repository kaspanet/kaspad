package dagtraversalmanager

import (
	"container/heap"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type blockHeapNode struct {
	hash         *externalapi.DomainHash
	ghostdagData *model.BlockGHOSTDAGData
}

func (left *blockHeapNode) less(right *blockHeapNode, gm model.GHOSTDAGManager) bool {
	return gm.Less(left.hash, left.ghostdagData, right.hash, right.ghostdagData)
}

// baseHeap  is an implementation for heap.Interface that sorts blocks by their blueWork+hash
type baseHeap struct {
	slice           []*blockHeapNode
	ghostdagManager model.GHOSTDAGManager
}

func (h *baseHeap) Len() int      { return len(h.slice) }
func (h *baseHeap) Swap(i, j int) { h.slice[i], h.slice[j] = h.slice[j], h.slice[i] }

func (h *baseHeap) Push(x interface{}) {
	h.slice = append(h.slice, x.(*blockHeapNode))
}

func (h *baseHeap) Pop() interface{} {
	oldSlice := h.slice
	oldLength := len(oldSlice)
	popped := oldSlice[oldLength-1]
	h.slice = oldSlice[0 : oldLength-1]
	return popped
}

// peek returns the block with lowest blueWork+hash from this heap without removing it
func (h *baseHeap) peek() *blockHeapNode {
	return h.slice[0]
}

// upHeap extends baseHeap to include Less operation that traverses from bottom to top
type upHeap struct{ baseHeap }

func (h *upHeap) Less(i, j int) bool {
	heapNodeI := h.slice[i]
	heapNodeJ := h.slice[j]
	return heapNodeI.less(heapNodeJ, h.ghostdagManager)
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h *downHeap) Less(i, j int) bool {
	heapNodeI := h.slice[i]
	heapNodeJ := h.slice[j]
	return !heapNodeI.less(heapNodeJ, h.ghostdagManager)
}

// blockHeap represents a mutable heap of blocks, sorted by their blueWork+hash
type blockHeap struct {
	impl          heap.Interface
	ghostdagStore model.GHOSTDAGDataStore
	dbContext     model.DBReader
}

// NewDownHeap initializes and returns a new blockHeap
func (dtm *dagTraversalManager) NewDownHeap() model.BlockHeap {
	h := blockHeap{
		impl:          &downHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
	}
	heap.Init(h.impl)
	return &h
}

// NewUpHeap initializes and returns a new blockHeap
func (dtm *dagTraversalManager) NewUpHeap() model.BlockHeap {
	h := blockHeap{
		impl:          &upHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
	}
	heap.Init(h.impl)
	return &h
}

// Pop removes the block with lowest blueWork+hash from this heap and returns it
func (bh *blockHeap) Pop() *externalapi.DomainHash {
	return heap.Pop(bh.impl).(*blockHeapNode).hash
}

// Push pushes the block onto the heap
func (bh *blockHeap) Push(blockHash *externalapi.DomainHash) error {
	ghostdagData, err := bh.ghostdagStore.Get(bh.dbContext, blockHash)
	if err != nil {
		return err
	}

	heap.Push(bh.impl, &blockHeapNode{
		hash:         blockHash,
		ghostdagData: ghostdagData,
	})

	return nil
}

// Len returns the length of this heap
func (bh *blockHeap) Len() int {
	return bh.impl.Len()
}

// ToSlice copies this heap to a slice
func (bh *blockHeap) ToSlice() []*externalapi.DomainHash {
	length := bh.Len()
	hashes := make([]*externalapi.DomainHash, length)
	for i := 0; i < length; i++ {
		hashes[i] = bh.Pop()
	}
	return hashes
}

// sizedUpBlockHeap represents a mutable heap of Blocks, sorted by their blueWork+hash, capped by a specific size.
type sizedUpBlockHeap struct {
	impl          upHeap
	ghostdagStore model.GHOSTDAGDataStore
	dbContext     model.DBReader
}

// newSizedUpHeap initializes and returns a new sizedUpBlockHeap
func (dtm *dagTraversalManager) newSizedUpHeap(cap int) *sizedUpBlockHeap {
	h := sizedUpBlockHeap{
		impl:          upHeap{baseHeap{slice: make([]*blockHeapNode, 0, cap), ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
	}
	heap.Init(&h.impl)
	return &h
}

// len returns the length of this heap
func (sbh *sizedUpBlockHeap) len() int {
	return sbh.impl.Len()
}

// pop removes the block with lowest blueWork+hash from this heap and returns it
func (sbh *sizedUpBlockHeap) pop() *externalapi.DomainHash {
	return heap.Pop(&sbh.impl).(*blockHeapNode).hash
}

// tryPushWithGHOSTDAGData is just like tryPush but the caller provides the ghostdagData of the block.
func (sbh *sizedUpBlockHeap) tryPushWithGHOSTDAGData(blockHash *externalapi.DomainHash, ghostdagData *model.BlockGHOSTDAGData) (bool, error) {
	node := &blockHeapNode{
		hash:         blockHash,
		ghostdagData: ghostdagData,
	}
	if len(sbh.impl.slice) == cap(sbh.impl.slice) {
		min := sbh.impl.peek()
		// if the heap is full, and the new block is less than the minimum, return false
		if node.less(min, sbh.impl.ghostdagManager) {
			return false, nil
		}
		sbh.pop()
	}
	heap.Push(&sbh.impl, node)
	return true, nil
}

// tryPush tries to push the block onto the heap, if the heap is full and it's less than the minimum it rejects it
func (sbh *sizedUpBlockHeap) tryPush(blockHash *externalapi.DomainHash) (bool, error) {
	ghostdagData, err := sbh.ghostdagStore.Get(sbh.dbContext, blockHash)
	if err != nil {
		return false, err
	}
	return sbh.tryPushWithGHOSTDAGData(blockHash, ghostdagData)
}
