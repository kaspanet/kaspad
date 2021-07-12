package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	AddBlock(stagingArea *StagingArea, blockHash *externalapi.DomainHash, updateVirtual bool) (*externalapi.SelectedChainPath, externalapi.UTXODiff, *UTXODiffReversalData, error)
	PopulateTransactionWithUTXOEntries(stagingArea *StagingArea, transaction *externalapi.DomainTransaction) error
	ImportPruningPoint(stagingArea *StagingArea, newPruningPoint *externalapi.DomainHash) error
	RestorePastUTXOSetIterator(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.ReadOnlyUTXOSetIterator, error)
	CalculatePastUTXOAndAcceptanceData(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (externalapi.UTXODiff, externalapi.AcceptanceData, Multiset, error)
	GetVirtualSelectedParentChainFromBlock(stagingArea *StagingArea, blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
	RecoverUTXOIfRequired() error
	ReverseUTXODiffs(tipHash *externalapi.DomainHash, reversalData *UTXODiffReversalData) error
	ResolveVirtual() error
}
