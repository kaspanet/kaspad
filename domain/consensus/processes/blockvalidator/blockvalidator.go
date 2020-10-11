package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator struct {
}

// New instantiates a new BlockValidator
func New() *BlockValidator {
	return &BlockValidator{}
}

// ValidateHeaderInIsolation validates block headers in isolation from the current
// consensus state
func (bv *BlockValidator) ValidateHeaderInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (bv *BlockValidator) ValidateHeaderInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInIsolation validates block bodies in isolation from the current
// consensus state
func (bv *BlockValidator) ValidateBodyInIsolation(block *model.DomainBlock) error {
	return nil
}

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (bv *BlockValidator) ValidateBodyInContext(block *model.DomainBlock) error {
	return nil
}

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (bv *BlockValidator) ValidateAgainstPastUTXO(block *model.DomainBlock) error {
	return nil
}

// ValidateFinality makes sure the block does not violate finality
func (bv *BlockValidator) ValidateFinality(block *model.DomainBlock) error {
	return nil
}
