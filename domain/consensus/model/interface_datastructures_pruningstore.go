package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash)
	StagePruningPointUTXOSetIterator(pruningPointUTXOSetIterator ReadOnlyUTXOSetIterator)
	StagePruningPointCandidate(candidate *externalapi.DomainHash)
	IsStaged() bool
	PruningPointCandidate(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader) (bool, error)
	PruningPoint(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader) (bool, error)
	ClearCandidatePruningPointUTXOs(dbTx DBTransaction) error
	InsertCandidatePruningPointUTXOs(dbTx DBTransaction, outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
	CandidatePruningPointUTXOIterator(dbContext DBReader) (ReadOnlyUTXOSetIterator, error)
	ClearCandidatePruningPointMultiset(dbTx DBTransaction) error
	CandidatePruningPointMultiset(dbContext DBReader) (Multiset, error)
	UpdateCandidatePruningPointMultiset(dbTx DBTransaction, multiset Multiset) error
	CommitCandidatePruningPointUTXOSet() error
	PruningPointUTXOs(dbContext DBReader, offset int, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error)
}
