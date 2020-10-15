package model

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*DomainHash) (*BlockGHOSTDAGData, error)
	BlockData(blockHash *DomainHash) *BlockGHOSTDAGData
	ChooseSelectedParent(
		blockHashA *DomainHash, blockAGHOSTDAGData *BlockGHOSTDAGData,
		blockHashB *DomainHash, blockBGHOSTDAGData *BlockGHOSTDAGData) *DomainHash
}
