package model

import "github.com/kaspanet/kaspad/util/daghash"

// SelectedParentIterator ...
type SelectedParentIterator interface {
	Next() bool
	Get() *daghash.Hash
}
