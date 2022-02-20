package dagtraversalmanager

import (
	"container/heap"

	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func blockGHOSTDAGDataHashPairLess(left, right *externalapi.BlockGHOSTDAGDataHashPair, gm model.GHOSTDAGManager) bool {
	return gm.Less(left.Hash, left.GHOSTDAGData, right.Hash, right.GHOSTDAGData)
}

// baseHeap  is an implementation for heap.Interface that sorts blocks by their blueWork+hash
type baseHeap struct {
	slice           []*externalapi.BlockGHOSTDAGDataHashPair
	ghostdagManager model.GHOSTDAGManager
}

func (h *baseHeap) Len() int      { return len(h.slice) }
func (h *baseHeap) Swap(i, j int) { h.slice[i], h.slice[j] = h.slice[j], h.slice[i] }

func (h *baseHeap) Push(x interface{}) {
	h.slice = append(h.slice, x.(*externalapi.BlockGHOSTDAGDataHashPair))
}

func (h *baseHeap) Pop() interface{} {
	oldSlice := h.slice
	oldLength := len(oldSlice)
	popped := oldSlice[oldLength-1]
	h.slice = oldSlice[0 : oldLength-1]
	return popped
}

// peek returns the block with lowest blueWork+hash from this heap without removing it
func (h *baseHeap) peek() *externalapi.BlockGHOSTDAGDataHashPair {
	return h.slice[0]
}

// upHeap extends baseHeap to include Less operation that traverses from bottom to top
type upHeap struct{ baseHeap }

func (h *upHeap) Less(i, j int) bool {
	heapNodeI := h.slice[i]
	heapNodeJ := h.slice[j]
	return blockGHOSTDAGDataHashPairLess(heapNodeI, heapNodeJ, h.ghostdagManager)
}

// downHeap extends baseHeap to include Less operation that traverses from top to bottom
type downHeap struct{ baseHeap }

func (h *downHeap) Less(i, j int) bool {
	heapNodeI := h.slice[i]
	heapNodeJ := h.slice[j]
	return !blockGHOSTDAGDataHashPairLess(heapNodeI, heapNodeJ, h.ghostdagManager)
}

// blockHeap represents a mutable heap of blocks, sorted by their blueWork+hash
type blockHeap struct {
	impl          heap.Interface
	ghostdagStore model.GHOSTDAGDataStore
	dbContext     model.DBReader
	stagingArea   *model.StagingArea
}

// NewDownHeap initializes and returns a new blockHeap
func (dtm *dagTraversalManager) NewDownHeap(stagingArea *model.StagingArea) model.BlockHeap {
	h := blockHeap{
		impl:          &downHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
		stagingArea:   stagingArea,
	}
	heap.Init(h.impl)
	return &h
}

// NewUpHeap initializes and returns a new blockHeap
func (dtm *dagTraversalManager) NewUpHeap(stagingArea *model.StagingArea) model.BlockHeap {
	h := blockHeap{
		impl:          &upHeap{baseHeap{ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
		stagingArea:   stagingArea,
	}
	heap.Init(h.impl)
	return &h
}

// Pop removes the block with lowest blueWork+hash from this heap and returns it
func (bh *blockHeap) Pop() *externalapi.DomainHash {
	return heap.Pop(bh.impl).(*externalapi.BlockGHOSTDAGDataHashPair).Hash
}

// Push pushes the block onto the heap
func (bh *blockHeap) Push(blockHash *externalapi.DomainHash) error {
	ghostdagData, err := bh.ghostdagStore.Get(bh.dbContext, bh.stagingArea, blockHash, false)
	if err != nil {
		return err
	}

	heap.Push(bh.impl, &externalapi.BlockGHOSTDAGDataHashPair{
		Hash:         blockHash,
		GHOSTDAGData: ghostdagData,
	})

	return nil
}

func (bh *blockHeap) PushSlice(blockHashes []*externalapi.DomainHash) error {
	for _, blockHash := range blockHashes {
		err := bh.Push(blockHash)
		if err != nil {
			return err
		}
	}
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
	stagingArea   *model.StagingArea
}

// newSizedUpHeap initializes and returns a new sizedUpBlockHeap
func (dtm *dagTraversalManager) newSizedUpHeap(stagingArea *model.StagingArea, cap int) *sizedUpBlockHeap {
	h := sizedUpBlockHeap{
		impl:          upHeap{baseHeap{slice: make([]*externalapi.BlockGHOSTDAGDataHashPair, 0, cap), ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
		stagingArea:   stagingArea,
	}
	heap.Init(&h.impl)
	return &h
}

func (dtm *dagTraversalManager) newSizedUpHeapFromSlice(stagingArea *model.StagingArea, slice []*externalapi.BlockGHOSTDAGDataHashPair) *sizedUpBlockHeap {
	h := sizedUpBlockHeap{
		impl:          upHeap{baseHeap{slice: slice, ghostdagManager: dtm.ghostdagManager}},
		ghostdagStore: dtm.ghostdagDataStore,
		dbContext:     dtm.databaseContext,
		stagingArea:   stagingArea,
	}
	return &h
}

// len returns the length of this heap
func (sbh *sizedUpBlockHeap) len() int {
	return sbh.impl.Len()
}

// pop removes the block with lowest blueWork+hash from this heap and returns it
func (sbh *sizedUpBlockHeap) pop() *externalapi.DomainHash {
	return heap.Pop(&sbh.impl).(*externalapi.BlockGHOSTDAGDataHashPair).Hash
}

// tryPushWithGHOSTDAGData is just like tryPush but the caller provides the ghostdagData of the block.
func (sbh *sizedUpBlockHeap) tryPushWithGHOSTDAGData(blockHash *externalapi.DomainHash,
	ghostdagData *externalapi.BlockGHOSTDAGData) (bool, error) {

	node := &externalapi.BlockGHOSTDAGDataHashPair{
		Hash:         blockHash,
		GHOSTDAGData: ghostdagData,
	}
	if len(sbh.impl.slice) == cap(sbh.impl.slice) {
		min := sbh.impl.peek()
		// if the heap is full, and the new block is less than the minimum, return false
		if blockGHOSTDAGDataHashPairLess(node, min, sbh.impl.ghostdagManager) {
			return false, nil
		}
		sbh.pop()
	}
	heap.Push(&sbh.impl, node)
	return true, nil
}

// tryPush tries to push the block onto the heap, if the heap is full and it's less than the minimum it rejects it
func (sbh *sizedUpBlockHeap) tryPush(blockHash *externalapi.DomainHash) (bool, error) {
	ghostdagData, err := sbh.ghostdagStore.Get(sbh.dbContext, sbh.stagingArea, blockHash, false)
	if err != nil {
		return false, err
	}
	return sbh.tryPushWithGHOSTDAGData(blockHash, ghostdagData)
}
