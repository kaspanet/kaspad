package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash)
	StagePruningPointUTXOSet(pruningPointUTXOSetIterator ReadOnlyUTXOSetIterator)
	StagePruningPointCandidate(candidate *externalapi.DomainHash)
	IsStaged() bool
	PruningPointCandidate(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader) (bool, error)
	PruningPoint(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader) (bool, error)
	ClearImportedPruningPointUTXOs(dbTx DBTransaction) error
	InsertImportedPruningPointUTXOs(dbTx DBTransaction, outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
	ImportedPruningPointUTXOIterator(dbContext DBReader) (ReadOnlyUTXOSetIterator, error)
	ClearImportedPruningPointMultiset(dbTx DBTransaction) error
	ImportedPruningPointMultiset(dbContext DBReader) (Multiset, error)
	UpdateImportedPruningPointMultiset(dbTx DBTransaction, multiset Multiset) error
	CommitImportedPruningPointUTXOSet() error
	PruningPointUTXOs(dbContext DBReader, offset int, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error)
}
