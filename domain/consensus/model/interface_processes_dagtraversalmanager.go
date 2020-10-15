package model

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager interface {
	ChainBlockAtBlueScore(highHash *DomainHash, blueScore uint64) *DomainHash
	SelectedParentIterator(highHash *DomainHash) SelectedParentIterator
}
