package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(pruningPointBlockHash *externalapi.DomainHash)
	StagePruningPointCandidate(candidate *externalapi.DomainHash)
	IsStaged() bool
	PruningPointCandidate(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader) (bool, error)
	PruningPoint(dbContext DBReader) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader) (bool, error)

	StageStartUpdatingPruningPointUTXOSet()
	HadStartedUpdatingPruningPointUTXOSet(dbContext DBWriter) (bool, error)
	FinishUpdatingPruningPointUTXOSet(dbContext DBWriter) error
	UpdatePruningPointUTXOSet(dbContext DBWriter, utxoSetIterator ReadOnlyUTXOSetIterator) error

	ClearImportedPruningPointUTXOs(dbContext DBWriter) error
	AppendImportedPruningPointUTXOs(dbTx DBTransaction, outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
	ImportedPruningPointUTXOIterator(dbContext DBReader) (ReadOnlyUTXOSetIterator, error)
	ClearImportedPruningPointMultiset(dbContext DBWriter) error
	ImportedPruningPointMultiset(dbContext DBReader) (Multiset, error)
	UpdateImportedPruningPointMultiset(dbTx DBTransaction, multiset Multiset) error
	CommitImportedPruningPointUTXOSet(dbContext DBWriter) error
	PruningPointUTXOs(dbContext DBReader, offset int, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error)
}
