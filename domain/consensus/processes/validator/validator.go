package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

// Validator exposes a set of validation classes, after which
// it's possible to determine whether either a block or a
// transaction is valid
type Validator struct {
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

// New instantiates a new Validator
func New(consensusStateManager model.ConsensusStateManager) *Validator {
	return &Validator{
		consensusStateManager: consensusStateManager,
	}
}
