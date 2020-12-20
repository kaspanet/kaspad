package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// AcceptanceDataStore represents a store of AcceptanceData
type AcceptanceDataStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, acceptanceData externalapi.AcceptanceData)
	IsStaged() bool
	Get(dbContext DBReader, blockHash *externalapi.DomainHash) (externalapi.AcceptanceData, error)
	Delete(blockHash *externalapi.DomainHash)
}
