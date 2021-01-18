package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	AddBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
	PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error
	UpdatePruningPoint(newPruningPoint *externalapi.DomainBlock) error
	RestorePastUTXOSetIterator(blockHash *externalapi.DomainHash) (ReadOnlyUTXOSetIterator, error)
	CalculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (UTXODiff, externalapi.AcceptanceData, Multiset, error)
	GetVirtualSelectedParentChainFromBlock(blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error)
}
