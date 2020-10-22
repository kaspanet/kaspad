package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ReachabilityDataStore represents a store of ReachabilityData
type ReachabilityDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, reachabilityData *ReachabilityData) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (*ReachabilityData, error)
	ReindexRoot(dbContext DBContextProxy) (*externalapi.DomainHash, error)
}
