package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningStore represents a store for the current pruning state
type PruningStore interface {
	Store
	StagePruningPoint(dbContext DBWriter, stagingArea *StagingArea, pruningPointBlockHash *externalapi.DomainHash) error
	StagePruningPointByIndex(dbContext DBReader, stagingArea *StagingArea,
		pruningPointBlockHash *externalapi.DomainHash, index uint64) error
	StagePruningPointCandidate(stagingArea *StagingArea, candidate *externalapi.DomainHash)
	IsStaged(stagingArea *StagingArea) bool
	PruningPointCandidate(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	HasPruningPointCandidate(dbContext DBReader, stagingArea *StagingArea) (bool, error)
	PruningPoint(dbContext DBReader, stagingArea *StagingArea) (*externalapi.DomainHash, error)
	HasPruningPoint(dbContext DBReader, stagingArea *StagingArea) (bool, error)
	CurrentPruningPointIndex(dbContext DBReader, stagingArea *StagingArea) (uint64, error)
	PruningPointByIndex(dbContext DBReader, stagingArea *StagingArea, index uint64) (*externalapi.DomainHash, error)

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
