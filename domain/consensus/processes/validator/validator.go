package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockAndTransactionValidator exposes a set of validation
// classes, after which it's possible to determine whether
// either a block or a transaction is valid
type BlockAndTransactionValidator interface {
	model.BlockValidator
	model.TransactionValidator
}

// validator exposes a set of validation classes, after which
// it's possible to determine whether either a block or a
// transaction is valid
type validator struct {
	consensusStateManager model.ConsensusStateManager
}

// New instantiates a new validator
func New(consensusStateManager model.ConsensusStateManager) BlockAndTransactionValidator {
	return &validator{
		consensusStateManager: consensusStateManager,
	}
}

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (v *validator) ValidateHeaderInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *validator) ValidateHeaderInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (v *validator) ValidateBodyInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *validator) ValidateBodyInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (v *validator) ValidateAgainstPastUTXO(block *model.DomainBlock) error {
	return nil
}

// ValidateFinality makes sure the block does not violate finality
func (v *validator) ValidateFinality(block *model.DomainBlock) error {
	return nil
}

// ValidateTransactionAndCalculateFee validates the given transaction using
// the given utxoEntries. It also returns the transaction's fee
func (v *validator) ValidateTransactionAndCalculateFee(transaction *model.DomainTransaction, utxoEntries []*model.UTXOEntry) (fee uint64, err error) {
	return 0, nil
}
