package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(stagingArea *StagingArea, pruningPointBlockHash *externalapi.DomainHash)
	StagePreviousPruningPoint(stagingArea *StagingArea, oldPruningPointBlockHash *externalapi.DomainHash)
	StagePruningPointCandidate(stagingArea *StagingArea, candidate *externalapi.DomainHash)
	IsStaged(stagingArea *StagingArea) bool
	PruningPointCandidate(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader, stagingArea *StagingArea) (bool, error)
	PreviousPruningPoint(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	PruningPoint(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader, stagingArea *StagingArea) (bool, error)

	StageStartUpdatingPruningPointUTXOSet(stagingArea *StagingArea)
	HadStartedUpdatingPruningPointUTXOSet(dbContext DBWriter) (bool, error)
	FinishUpdatingPruningPointUTXOSet(dbContext DBWriter) error
	UpdatePruningPointUTXOSet(dbContext DBWriter, diff externalapi.UTXODiff) error

	ClearImportedPruningPointUTXOs(dbContext DBWriter) error
	AppendImportedPruningPointUTXOs(dbTx DBTransaction, outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
	ImportedPruningPointUTXOIterator(dbContext DBReader) (externalapi.ReadOnlyUTXOSetIterator, error)
	ClearImportedPruningPointMultiset(dbContext DBWriter) error
	ImportedPruningPointMultiset(dbContext DBReader) (Multiset, error)
	UpdateImportedPruningPointMultiset(dbTx DBTransaction, multiset Multiset) error
	CommitImportedPruningPointUTXOSet(dbContext DBWriter) error
	PruningPointUTXOs(dbContext DBReader, fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error)
	PruningPointUTXOIterator(dbContext DBReader) (externalapi.ReadOnlyUTXOSetIterator, error)
}
