package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTraversalManager exposes methods for traversing blocks
// in the DAG
type DAGTraversalManager interface {
	BlockAtDepth(stagingArea *StagingArea, highHash *externalapi.DomainHash, depth uint64) (*externalapi.DomainHash, error)
	LowestChainBlockAboveOrEqualToBlueScore(stagingArea *StagingArea, highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error)
	// SelectedChildIterator should return a BlockIterator that iterates
	// from lowHash (exclusive) to highHash (inclusive) over highHash's selected parent chain
	SelectedChildIterator(stagingArea *StagingArea, highHash, lowHash *externalapi.DomainHash) (BlockIterator, error)
	Anticone(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	AnticoneFromVirtual(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	BlockWindow(stagingArea *StagingArea, highHash *externalapi.DomainHash, windowSize int, isBlockWithPrefilledData bool) ([]*externalapi.DomainHash, error)
	NewDownHeap(stagingArea *StagingArea) BlockHeap
	NewUpHeap(stagingArea *StagingArea) BlockHeap
	CalculateChainPath(stagingArea *StagingArea, fromBlockHash, toBlockHash *externalapi.DomainHash) (
		*externalapi.SelectedChainPath, error)
}
