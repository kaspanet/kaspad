package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, acceptanceData *AcceptanceData)
	IsStaged() bool
	Get(dbContext DBReader, blockHash *externalapi.DomainHash) (*AcceptanceData, error)
	Delete(dbTx DBTransaction, blockHash *externalapi.DomainHash) error
}
