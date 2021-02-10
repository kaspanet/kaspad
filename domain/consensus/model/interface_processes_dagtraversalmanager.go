package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTraversalManager exposes methods for traversing blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error)
	LowestChainBlockAboveOrEqualToBlueScore(highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error)
	// SelectedChildIterator should return a BlockIterator that iterates
	// from lowHash (exclusive) to highHash (inclusive) over highHash's selected parent chain
	SelectedChildIterator(highHash, lowHash *externalapi.DomainHash) (BlockIterator, error)
	Anticone(blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	BlueWindow(highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.DomainHash, error)
	NewDownHeap() BlockHeap
	NewUpHeap() BlockHeap
	CalculateChainPath(
		fromBlockHash, toBlockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
}
