package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// DAGTraversalManager exposes methods for traversing blocks
// in the DAG
type DAGTraversalManager interface {
	LowestChainBlockAboveOrEqualToBlueScore(stagingArea *StagingArea, highHash *externalapi.DomainHash, blueScore uint64) (*externalapi.DomainHash, error)
	// SelectedChildIterator should return a BlockIterator that iterates
	// from lowHash (exclusive) to highHash (inclusive) over highHash's selected parent chain
	SelectedChildIterator(stagingArea *StagingArea, highHash, lowHash *externalapi.DomainHash, includeLowHash bool) (BlockIterator, error)
	SelectedChild(stagingArea *StagingArea, highHash, lowHash *externalapi.DomainHash) (*externalapi.DomainHash, error)
	AnticoneFromBlocks(stagingArea *StagingArea, tips []*externalapi.DomainHash, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	AnticoneFromVirtualPOV(stagingArea *StagingArea, blockHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	BlockWindow(stagingArea *StagingArea, highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.DomainHash, error)
	BlockWindowWithGHOSTDAGData(stagingArea *StagingArea, highHash *externalapi.DomainHash, windowSize int) ([]*externalapi.BlockGHOSTDAGDataHashPair, error)
	DAABlockWindow(stagingArea *StagingArea, highHash *externalapi.DomainHash) ([]*externalapi.DomainHash, error)
	NewDownHeap(stagingArea *StagingArea) BlockHeap
	NewUpHeap(stagingArea *StagingArea) BlockHeap
	CalculateChainPath(stagingArea *StagingArea, fromBlockHash, toBlockHash *externalapi.DomainHash) (
		*externalapi.SelectedChainPath, error)
}
