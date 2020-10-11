package model

// SelectedParentIterator is an iterator over the selected parent
// chain of some block
type SelectedParentIterator interface {
	Next() bool
	Get() *DomainHash
}
