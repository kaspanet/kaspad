package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

// BlockValidator exposes a set of validation classes, after which
// it's possible to determine whether a block is valid
type BlockValidator struct {
	powMax                         *big.Int
	skipPoW                        bool
	genesisHash                    *model.DomainHash
	enableNonNativeSubnetworks     bool
	disableDifficultyAdjustment    bool
	powMaxBits                     uint32
	difficultyAdjustmentWindowSize uint64

	dagTopologyManager    model.DAGTopologyManager
	ghostdagManager       model.GHOSTDAGManager
	consensusStateManager model.ConsensusStateManager
}

// New instantiates a new BlockValidator
func New(powMax *big.Int, skipPoW bool) *BlockValidator {
	return &BlockValidator{
		powMax:  powMax,
		skipPoW: skipPoW,
	}
}

// ValidateFinality makes sure the block does not violate finality
func (bv *BlockValidator) ValidateFinality(block *model.DomainBlock) error {
	return nil
}
