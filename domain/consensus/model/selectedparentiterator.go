package model

import "github.com/kaspanet/kaspad/util/daghash"

// SelectedParentIterator is an iterator over the selected parent
// chain of some block
type SelectedParentIterator interface {
	Next() bool
	Get() *daghash.Hash
}
