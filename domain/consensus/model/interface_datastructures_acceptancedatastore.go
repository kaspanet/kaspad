package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Stage(blockHash *externalapi.DomainHash, acceptanceData []*BlockAcceptanceData)
	IsStaged() bool
	Discard()
	Commit(dbTx DBTxProxy) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) ([]*BlockAcceptanceData, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
