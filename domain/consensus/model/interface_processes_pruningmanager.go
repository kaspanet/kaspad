package model

import "github.com/kaspanet/kaspad/util/daghash"

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	FindNextPruningPoint(blockHash *daghash.Hash) (found bool, newPruningPoint *daghash.Hash, newPruningPointUTXOSet ReadOnlyUTXOSet)
	PruningPoint() *daghash.Hash
	SerializedUTXOSet() []byte
}
