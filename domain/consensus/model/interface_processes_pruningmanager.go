package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// PruningManager resolves and manages the current pruning point
type PruningManager interface {
	UpdatePruningPointByVirtual(stagingArea *StagingArea) error
	IsValidPruningPoint(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (bool, error)
	ClearImportedPruningPointData() error
	AppendImportedPruningPointUTXOs(outpointAndUTXOEntryPairs []*externalapi.OutpointAndUTXOEntryPair) error
	UpdatePruningPointUTXOSetIfRequired() error
	PruneAllBlocksBelow(stagingArea *StagingArea, pruningPointHash *externalapi.DomainHash) error
}
