package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// Validator exposes a set of validation classes, after which
// it's possible to determine whether either a block or a
// transaction is valid
type Validator struct {
	consensusStateManager model.ConsensusStateManager
}

// New instantiates a new Validator
func New(consensusStateManager model.ConsensusStateManager) *Validator {
	return &Validator{
		consensusStateManager: consensusStateManager,
	}
}

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (bv *Validator) ValidateHeaderInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (bv *Validator) ValidateHeaderInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (bv *Validator) ValidateBodyInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (bv *Validator) ValidateBodyInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (bv *Validator) ValidateAgainstPastUTXO(block *model.DomainBlock) error {
	return nil
}

// ValidateFinality makes sure the block does not violate finality
func (bv *Validator) ValidateFinality(block *model.DomainBlock) error {
	return nil
}

// ValidateTransactionInIsolation validates transactions in isolation from the current
// consensus state
func (bv *Validator) ValidateTransactionInIsolation(transaction *model.DomainTransaction) error {
	return nil
}

// ValidateTransactionInContext validates transactions in the context of the current
// consensus state
func (bv *Validator) ValidateTransactionInContext(transaction *model.DomainTransaction) error {
	return nil
}

// ValidateTransactionAndCalculateFee validates the given transaction using
// the given utxoEntries. It also returns the transaction's fee
func (bv *Validator) ValidateTransactionAndCalculateFee(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) (fee uint64, err error) {
	return 0, nil
}

// ValidateTransactionAgainstUTXO validates transactions in the context of the current
// UTXO Set
func (bv *Validator) ValidateTransactionAgainstUTXO(transaction *model.DomainTransaction) error {
	return nil
}
