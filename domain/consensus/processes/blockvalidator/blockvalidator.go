package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator struct {
	powMax                     *big.Int
	skipPoW                    bool
	genesisHash                *model.DomainHash
	enableNonNativeSubnetworks bool
}

// New instantiates a new BlockValidator
func New(powMax *big.Int, skipPoW bool) *BlockValidator {
	return &BlockValidator{
		powMax:  powMax,
		skipPoW: skipPoW,
	}
}

// ValidateHeaderInContext validates block headers in the context of the current
// consensus state
func (bv *BlockValidator) ValidateHeaderInContext(block *model.DomainBlock) error {
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
