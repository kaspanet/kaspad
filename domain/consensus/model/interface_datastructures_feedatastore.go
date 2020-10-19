package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// FeeDataStore represents a store of fee data
type FeeDataStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, fee uint64) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (uint64, error)
}
