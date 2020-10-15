package model

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, acceptanceData *BlockAcceptanceData) error
	Get(dbContext DBContextProxy, blockHash *DomainHash) (*BlockAcceptanceData, error)
	Delete(dbTx DBTxProxy, blockHash *DomainHash) error
}
