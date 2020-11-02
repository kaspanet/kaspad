package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTraversalManager exposes methods for traversing blocks
// in the DAG
type DAGTraversalManager interface {
	HighestChainBlockBelowBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error)
	SelectedParentIterator(highHash *externalapi.DomainHash) SelectedParentIterator
	BlueWindow(highHash *externalapi.DomainHash, windowSize uint64) ([]*externalapi.DomainHash, error)
	NewDownHeap() BlockHeap
	NewUpHeap() BlockHeap
}
