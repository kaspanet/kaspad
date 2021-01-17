package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash, pruningPointUTXOSetIterator ReadOnlyUTXOSetIterator)
	StagePruningPointCandidate(candidate *externalapi.DomainHash)
	IsStaged() bool
	PruningPointCandidate(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader) (bool, error)
	PruningPoint(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader) (bool, error)
	PruningPointSerializedUTXOSet(dbContext DBReader) ([]byte, error)
	InsertPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
}
