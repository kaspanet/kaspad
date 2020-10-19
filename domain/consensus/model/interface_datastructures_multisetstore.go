package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Insert(dbTx DBTxProxy, blockHash *externalapi.DomainHash, multiset Multiset) error
	Get(dbContext DBContextProxy, blockHash *externalapi.DomainHash) (Multiset, error)
	Delete(dbTx DBTxProxy, blockHash *externalapi.DomainHash) error
}
