package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// MultisetStore represents a store of Multisets
type MultisetStore interface {
	Store
	Stage(blockHash *externalapi.DomainHash, multiset Multiset)
	IsStaged() bool
	Get(dbContext DBReader, blockHash *externalapi.DomainHash) (Multiset, error)
	Delete(blockHash *externalapi.DomainHash)
}
