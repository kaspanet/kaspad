package validator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

// validator exposes a set of validation classes, after which
// it's possible to determine whether either a block or a
// transaction is valid
type validator struct {
	powMax                         *big.Int
	skipPoW                        bool
	genesisHash                    *model.DomainHash
	enableNonNativeSubnetworks     bool
	disableDifficultyAdjustment    bool
	powMaxBits                     uint32
	difficultyAdjustmentWindowSize uint64
	blockCoinbaseMaturity          uint64
	finalityDepth                  uint64

	dagTopologyManager    model.DAGTopologyManager
	ghostdagManager       model.GHOSTDAGManager
	consensusStateManager model.ConsensusStateManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	dagTraversalManager   model.DAGTraversalManager
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
