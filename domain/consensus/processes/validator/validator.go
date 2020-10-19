package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// validator exposes a set of validation classes, after which
// it's possible to determine whether either a block or a
// transaction is valid
type validator struct {
	consensusStateManager model.ConsensusStateManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
}

// New instantiates a new BlockAndTransactionValidator
func New(
	consensusStateManager model.ConsensusStateManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager) interface {
	model.BlockValidator
	model.TransactionValidator
} {

	return &validator{
		consensusStateManager: consensusStateManager,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
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

// ValidateTransactionAndPopulateWithConsensusData validates the given transaction
// and populates it with any missing consensus data
func (v *validator) ValidateTransactionAndPopulateWithConsensusData(transaction *model.DomainTransaction) error {
	return nil
}
