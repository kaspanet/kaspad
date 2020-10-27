package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	Stage(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSet ReadOnlyUTXOSet)
	IsStaged() bool
	PruningPoint(dbContext DBContextProxy) (*externalapi.DomainHash, error)
	PruningPointSerializedUTXOSet(dbContext DBContextProxy) ([]byte, error)
}
