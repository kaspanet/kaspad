package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	AddBlock(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, externalapi.UTXODiff, error)
	PopulateTransactionWithUTXOEntries(stagingArea *StagingArea, transaction *externalapi.DomainTransaction) error
	ImportPruningPoint(stagingArea *StagingArea, newPruningPoint *externalapi.DomainBlock) error
	RestorePastUTXOSetIterator(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.ReadOnlyUTXOSetIterator, error)
	CalculatePastUTXOAndAcceptanceData(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, externalapi.AcceptanceData, Multiset, error)
	GetVirtualSelectedParentChainFromBlock(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
	RecoverUTXOIfRequired(stagingArea *StagingArea) error
}
