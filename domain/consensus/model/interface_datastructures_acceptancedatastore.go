package model

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Insert(dbTx TxContextProxy, blockHash *DomainHash, acceptanceData *BlockAcceptanceData)
	Get(dbContext ContextProxy, blockHash *DomainHash) *BlockAcceptanceData
}
