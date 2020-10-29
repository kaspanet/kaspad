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

// baseHeap  is an implementation for heap.Interface that sorts blocks by their height
type baseHeap struct {
	slice           []*blockHeapNode
	ghostdagManager model.GHOSTDAGManager
}

func (h baseHeap) Len() int      { return len(h.slice) }
func (h baseHeap) Swap(i, j int) { h.slice[i], h.slice[j] = h.slice[j], h.slice[i] }

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

// upHeap extends baseHeap to include Less operation that traverses from bottom to top
type upHeap struct{ baseHeap }

func (h upHeap) Less(i, j int) bool {
	heapNodeA := h.slice[i]
	heapNodeB := h.slice[j]
	return h.ghostdagManager.Less(heapNodeA.hash, heapNodeA.ghostdagData, heapNodeB.hash, heapNodeB.ghostdagData)
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h downHeap) Less(i, j int) bool {
	heapNodeA := h.slice[i]
	heapNodeB := h.slice[j]
	return !h.ghostdagManager.Less(heapNodeA.hash, heapNodeA.ghostdagData, heapNodeB.hash, heapNodeB.ghostdagData)
}

// BlockHeap represents a mutable heap of Blocks, sorted by their height
type BlockHeap struct {
	impl          heap.Interface
	ghostdagStore model.GHOSTDAGDataStore
	dbContext     model.DBReader
}

// NewDownHeap initializes and returns a new BlockHeap
func (dtm dagTraversalManager) NewDownHeap() model.BlockHeap {
	h := BlockHeap{
		impl:          &downHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
	}
	heap.Init(h.impl)
	return h
}

// NewUpHeap initializes and returns a new BlockHeap
func (dtm dagTraversalManager) NewUpHeap() model.BlockHeap {
	h := BlockHeap{
		impl:          &upHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
	}
	heap.Init(h.impl)
	return h
}

// pop removes the block with lowest height from this heap and returns it
func (bh BlockHeap) Pop() *externalapi.DomainHash {
	return heap.Pop(bh.impl).(*blockHeapNode).hash
}

// Push pushes the block onto the heap
func (bh BlockHeap) Push(blockHash *externalapi.DomainHash) error {
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
func (bh BlockHeap) Len() int {
	return bh.impl.Len()
}
