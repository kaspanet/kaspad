package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// blockValidator exposes a set of validation classes, after which
// it's possible to determine whether either a block is valid
type blockValidator struct {
	consensusStateManager model.ConsensusStateManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	ghostdagManager       model.GHOSTDAGManager

	blockStatusStore model.BlockStatusStore
}

// New instantiates a new BlockValidator
func New(
	consensusStateManager model.ConsensusStateManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	ghostdagManager model.GHOSTDAGManager,
	blockStatusStore model.BlockStatusStore) model.BlockValidator {

	return &blockValidator{
		consensusStateManager: consensusStateManager,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		transactionValidator:  transactionValidator,
		ghostdagManager:       ghostdagManager,

		blockStatusStore: blockStatusStore,
	}
}

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (v *blockValidator) ValidateHeaderInIsolation(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (v *blockValidator) ValidateHeaderInContext(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (v *blockValidator) ValidateBodyInIsolation(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *blockValidator) ValidateBodyInContext(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (v *blockValidator) ValidateAgainstPastUTXO(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateFinality makes sure the block does not violate finality
func (v *blockValidator) ValidateFinality(blockHash *externalapi.DomainHash) error {
	return nil
}
