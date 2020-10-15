package model

// GHOSTDAGDataStore represents a store of BlockGHOSTDAGData
type GHOSTDAGDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, blockGHOSTDAGData *BlockGHOSTDAGData) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (*BlockGHOSTDAGData, error)
}
