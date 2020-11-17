package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTraversalManager exposes methods for traversing blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error)
	LowestChainBlockAboveOrEqualToBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error)
	SelectedParentIterator(highHash *externalapi.DomainHash) BlockIterator
	SelectedChildIterator(highHash, lowHash *externalapi.DomainHash) (BlockIterator, error)
	AnticoneFromContext(context, lowHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	BlueWindow(highHash *externalapi.DomainHash, windowSize uint64) ([]*externalapi.DomainHash, error)
	NewDownHeap() BlockHeap
	NewUpHeap() BlockHeap
}
