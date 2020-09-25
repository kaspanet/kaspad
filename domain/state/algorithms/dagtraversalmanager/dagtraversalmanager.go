package dagtraversalmanager

import "github.com/kaspanet/kaspad/util/daghash"

type DAGTraversalManager interface {
	BlockAtDepth(uint64) *daghash.Hash
	SelectedParentIterator(highHash *daghash.Hash) SelectedParentIterator
}
