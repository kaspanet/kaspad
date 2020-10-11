package model

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *DomainHash, acceptanceData *BlockAcceptanceData)
	Get(dbContext DBContextProxy, blockHash *DomainHash) *BlockAcceptanceData
}
