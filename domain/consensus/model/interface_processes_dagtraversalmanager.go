package model

import "github.com/kaspanet/kaspad/util/daghash"

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(highHash *daghash.Hash, depth uint64) *daghash.Hash
	SelectedParentIterator(highHash *daghash.Hash) SelectedParentIterator
}
