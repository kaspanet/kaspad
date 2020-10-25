package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util/mstime"
	"time"
)

// blockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type blockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *database.DomainDBContext

	consensusStateManager model.ConsensusStateManager
	pruningManager        model.PruningManager
	blockValidator        model.BlockValidator
	dagTopologyManager    model.DAGTopologyManager
	reachabilityTree      model.ReachabilityTree
	difficultyManager     model.DifficultyManager
	ghostdagManager       model.GHOSTDAGManager
	pastMedianTimeManager model.PastMedianTimeManager
	acceptanceDataStore   model.AcceptanceDataStore
	blockStore            model.BlockStore
	blockStatusStore      model.BlockStatusStore
	blockRelationStore    model.BlockRelationStore
}

// New instantiates a new BlockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *database.DomainDBContext,
	consensusStateManager model.ConsensusStateManager,
	pruningManager model.PruningManager,
	blockValidator model.BlockValidator,
	dagTopologyManager model.DAGTopologyManager,
	reachabilityTree model.ReachabilityTree,
	difficultyManager model.DifficultyManager,
	pastMedianTimeManager model.PastMedianTimeManager,
	ghostdagManager model.GHOSTDAGManager,
	acceptanceDataStore model.AcceptanceDataStore,
	blockStore model.BlockStore,
	blockStatusStore model.BlockStatusStore,
	blockRelationStore model.BlockRelationStore) model.BlockProcessor {

	return &blockProcessor{
		dagParams:             dagParams,
		databaseContext:       databaseContext,
		pruningManager:        pruningManager,
		blockValidator:        blockValidator,
		dagTopologyManager:    dagTopologyManager,
		reachabilityTree:      reachabilityTree,
		difficultyManager:     difficultyManager,
		pastMedianTimeManager: pastMedianTimeManager,
		ghostdagManager:       ghostdagManager,

		consensusStateManager: consensusStateManager,
		acceptanceDataStore:   acceptanceDataStore,
		blockStore:            blockStore,
		blockStatusStore:      blockStatusStore,
		blockRelationStore:    blockRelationStore,
	}
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (bp *blockProcessor) BuildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	start := time.Now()
	log.Debugf("BuildBlock start")
	block, err := bp.buildBlock(coinbaseData, transactions)
	log.Debugf("BuildBlock end. Took: %s", time.Since(start))
	return block, err
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *externalapi.DomainBlock) error {
	start := mstime.Now()
	log.Debugf("ValidateAndInsertBlock start")
	err := bp.validateAndInsertBlock(block)
	log.Debugf("ValidateAndInsertBlock end. Took: %s", mstime.Since(start))
	return err
}

func (bp *blockProcessor) buildBlock(coinbaseData *externalapi.DomainCoinbaseData,
	transactions []*externalapi.DomainTransaction) (*externalapi.DomainBlock, error) {

	return nil, nil
}

func (bp *blockProcessor) validateAndInsertBlock(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateProofOfWork(block)
	if err != nil {
		// If the validation failed:
		//   Write in blockStatusStore that the block is invalid
		// return err
	}

	err = bp.validateBlockInIsolationAndInContext(block)
	if err != nil {
		return err
	}

	return nil
}

func (bp *blockProcessor) validateBlockInIsolationAndInContext(block *externalapi.DomainBlock) error {
	err := bp.blockValidator.ValidateHeaderInIsolation(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateHeaderInContext(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateBodyInIsolation(block.Hash)
	if err != nil {
		return err
	}

	err = bp.blockValidator.ValidateBodyInContext(block.Hash)
	if err != nil {
		return err
	}

	return nil
}

func (bp *blockProcessor) insertBlock(block *externalapi.DomainBlock) error {
	return nil
}

func (bp *blockProcessor) processNonValidBlock(block *externalapi.DomainBlock, blockStatus model.BlockStatus) error {
	return nil
}
