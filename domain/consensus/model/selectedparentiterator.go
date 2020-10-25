package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SelectedParentIterator is an iterator over the selected parent
// chain of some block
type SelectedParentIterator interface {
	Next() bool
	Get() *externalapi.DomainHash
}
