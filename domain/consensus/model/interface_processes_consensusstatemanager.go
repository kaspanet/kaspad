package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	AddBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, externalapi.UTXODiff, error)
	PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error
	ImportPruningPoint(newPruningPoint *externalapi.DomainBlock) error
	RestorePastUTXOSetIterator(blockHash *externalapi.DomainHash) (externalapi.ReadOnlyUTXOSetIterator, error)
	CalculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (externalapi.UTXODiff, externalapi.AcceptanceData, Multiset, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
	RecoverUTXOIfRequired() error
}
