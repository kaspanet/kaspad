package model

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData)
	Get(dbContext ContextProxy, blockHash *DomainHash) *BlockGHOSTDAGData
}
