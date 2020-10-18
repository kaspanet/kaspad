package model

// DAGTraversalManager exposes methods for travering blocks
// in the DAG
type DAGTraversalManager interface {
	HighestChainBlockBelowBlueScore(highHash *DomainHash, blueScore uint64) (*DomainHash, error)
	SelectedParentIterator(highHash *DomainHash) (SelectedParentIterator, error)
}
