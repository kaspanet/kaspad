package model

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(highHash *DomainHash, depth uint64) *DomainHash
	SelectedParentIterator(highHash *DomainHash) SelectedParentIterator
}
