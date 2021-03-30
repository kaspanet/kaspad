package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateStore represents a store for the current consensus state
type ConsensusStateStore interface {
	Store
	IsStaged(stagingArea *StagingArea) bool

	StageVirtualUTXODiff(stagingArea *StagingArea, virtualUTXODiff externalapi.UTXODiff)
	UTXOByOutpoint(dbContext DBReader, stagingArea *StagingArea, outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, error)
	HasUTXOByOutpoint(dbContext DBReader, stagingArea *StagingArea, outpoint *externalapi.DomainOutpoint) (bool, error)
	VirtualUTXOSetIterator(dbContext DBReader, stagingArea *StagingArea) (externalapi.ReadOnlyUTXOSetIterator, error)
	VirtualUTXOs(dbContext DBReader, fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error)

	StageTips(stagingArea *StagingArea, tipHashes []*externalapi.DomainHash)
	Tips(stagingArea *StagingArea, dbContext DBReader) ([]*externalapi.DomainHash, error)

	StartImportingPruningPointUTXOSet(dbContext DBWriter) error
	HadStartedImportingPruningPointUTXOSet(dbContext DBWriter) (bool, error)
	ImportPruningPointUTXOSetIntoVirtualUTXOSet(dbContext DBWriter, pruningPointUTXOSetIterator externalapi.ReadOnlyUTXOSetIterator) error
	FinishImportingPruningPointUTXOSet(dbContext DBWriter) error
}
