package model

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// ConsensusStateManager manages the node's consensus state
type ConsensusStateManager interface {
	AddBlockToVirtual(blockHash *externalapi.DomainHash) (*externalapi.SelectedParentChainChanges, error)
	PopulateTransactionWithUTXOEntries(transaction *externalapi.DomainTransaction) error
	SetPruningPointUTXOSet(serializedUTXOSet []byte) error
	RestorePastUTXOSetIterator(blockHash *externalapi.DomainHash) (ReadOnlyUTXOSetIterator, error)
	HeaderTipsPruningPoint() (*externalapi.DomainHash, error)
	CalculatePastUTXOAndAcceptanceData(blockHash *externalapi.DomainHash) (UTXODiff, externalapi.AcceptanceData, Multiset, error)
}
