package blockprocessor

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/dagconfig"
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

	start := mstime.Now()
	log.Debugf("BuildBlock start")

	parents := bp.selectParentsForNewBlock()
	transactions := transactionSelector(bp.consensusStateStore.FullUTXOSet())
	block := bp.buildBlock(coinbaseScriptPublicKey, coinbaseExtraData, parents, transactions)

	log.Debugf("BuildBlock end. Took: %s", mstime.Since(start))
	return block
}

func (bp *BlockProcessor) selectParentsForNewBlock() []*daghash.Hash {
	virtualParentHashes := bp.consensusStateStore.VirtualParents()
	if len(virtualParentHashes) < appmessage.MaxBlockParents {
		return virtualParentHashes
	}
	return virtualParentHashes[:appmessage.MaxBlockParents]
}

func (bp *BlockProcessor) buildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	parentHashes []*daghash.Hash, transactions []*util.Tx) *appmessage.MsgBlock {

	return nil, nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *blockProcessor) ValidateAndInsertBlock(block *externalapi.DomainBlock) error {
	start := mstime.Now()
	log.Debugf("ValidateAndInsertBlock start")

	shouldInsertBlock, blockStatus, err := bp.validateBlock(block)
	if err != nil {
		return err
	}
	if !shouldInsertBlock {
		return nil
	}
	if blockStatus != model.StatusValid {
		return bp.processNonValidBlock(block, blockStatus)
	}
	err = bp.insertBlock(block)
	if err != nil {
		return err
	}

	log.Debugf("ValidateAndInsertBlock end. Took: %s", mstime.Since(start))
	return nil
}

func (bp *BlockProcessor) validateBlock(block *appmessage.MsgBlock) (
	shouldInsertBlock bool, blockStatus model.BlockStatus, err error) {

	var validationErr *model.ValidationError

	err = bp.blockValidator.ValidateProofOfWork(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return false, model.StatusValidateFailed, nil
	}

	err = bp.blockValidator.ValidateHeaderInIsolation(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusValidateFailed, nil
	}

	err = bp.blockValidator.ValidateHeaderInContext(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusValidateFailed, nil
	}

	err = bp.blockValidator.ValidateBodyInIsolation(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusValidateFailed, nil
	}

	err = bp.blockValidator.ValidateBodyInContext(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusValidateFailed, nil
	}

	err = bp.blockValidator.ValidateAgainstPastUTXO(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusDisqualifiedFromChain, nil
	}

	err = bp.blockValidator.ValidateFinality(block)
	if err != nil {
		if !errors.As(err, validationErr) {
			return false, model.StatusValidateFailed, err
		}
		return true, model.StatusUTXOPendingVerification, nil
	}

	return true, model.StatusValid, nil
}

func (bp *BlockProcessor) insertBlock(block *appmessage.MsgBlock) error {
	return nil
}

func (bp *BlockProcessor) processNonValidBlock(block *appmessage.MsgBlock, blockStatus model.BlockStatus) error {
	return nil
}
