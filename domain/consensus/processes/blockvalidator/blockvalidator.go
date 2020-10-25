package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"math/big"
)

// blockValidator exposes a set of validation classes, after which
// it's possible to determine whether either a block is valid
type blockValidator struct {
	powMax                         *big.Int
	skipPoW                        bool
	genesisHash                    *externalapi.DomainHash
	enableNonNativeSubnetworks     bool
	disableDifficultyAdjustment    bool
	powMaxBits                     uint32
	difficultyAdjustmentWindowSize uint64
	finalityDepth                  uint64

	databaseContext       model.DBContextProxy
	consensusStateManager model.ConsensusStateManager
	difficultyManager     model.DifficultyManager
	pastMedianTimeManager model.PastMedianTimeManager
	transactionValidator  model.TransactionValidator
	utxoDiffManager       model.UTXODiffManager
	acceptanceManager     model.AcceptanceManager
	ghostdagManager       model.GHOSTDAGManager
	dagTopologyManager    model.DAGTopologyManager
	dagTraversalManager   model.DAGTraversalManager

	blockStore        model.BlockStore
	ghostdagDataStore model.GHOSTDAGDataStore
}

// New instantiates a new BlockValidator
func New(powMax *big.Int,
	skipPoW bool,
	genesisHash *externalapi.DomainHash,
	enableNonNativeSubnetworks bool,
	disableDifficultyAdjustment bool,
	powMaxBits uint32,
	difficultyAdjustmentWindowSize uint64,
	finalityDepth uint64,
	databaseContext model.DBContextProxy,
	consensusStateManager model.ConsensusStateManager,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	transactionValidator model.TransactionValidator,
	utxoDiffManager model.UTXODiffManager,
	acceptanceManager model.AcceptanceManager,
	ghostdagManager model.GHOSTDAGManager,
	dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	blockStore model.BlockStore,
	ghostdagDataStore model.GHOSTDAGDataStore) *blockValidator {
	return &blockValidator{powMax: powMax, skipPoW: skipPoW, genesisHash: genesisHash, enableNonNativeSubnetworks: enableNonNativeSubnetworks, disableDifficultyAdjustment: disableDifficultyAdjustment, powMaxBits: powMaxBits, difficultyAdjustmentWindowSize: difficultyAdjustmentWindowSize, finalityDepth: finalityDepth, databaseContext: databaseContext, consensusStateManager: consensusStateManager, difficultyManager: difficultyManager, pastMedianTimeManager: pastMedianTimeManager, transactionValidator: transactionValidator, utxoDiffManager: utxoDiffManager, acceptanceManager: acceptanceManager, ghostdagManager: ghostdagManager, dagTopologyManager: dagTopologyManager, dagTraversalManager: dagTraversalManager, blockStore: blockStore, ghostdagDataStore: ghostdagDataStore}
}

// ValidateAgainstPastUTXO validates the block against the UTXO of its past
func (v *blockValidator) ValidateAgainstPastUTXO(blockHash *externalapi.DomainHash) error {
	return nil
}

// ValidateFinality makes sure the block does not violate finality
func (v *blockValidator) ValidateFinality(blockHash *externalapi.DomainHash) error {
	return nil
}
