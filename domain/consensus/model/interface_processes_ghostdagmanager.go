package model

// GHOSTDAGManager resolves and manages GHOSTDAG block data
type GHOSTDAGManager interface {
	GHOSTDAG(blockParents []*DomainHash) *BlockGHOSTDAGData
	BlockData(blockHash *DomainHash) *BlockGHOSTDAGData
}
