package blockprocessor

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/datastructures"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/processes"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

// BlockProcessor is responsible for processing incoming blocks
// and creating blocks from the current state
type BlockProcessor struct {
	dagParams       *dagconfig.Params
	databaseContext *dbaccess.DatabaseContext

	consensusStateManager processes.ConsensusStateManager
	pruningManager        processes.PruningManager
	blockValidator        processes.BlockValidator
	dagTopologyManager    processes.DAGTopologyManager
	reachabilityTree      processes.ReachabilityTree
	acceptanceDataStore   datastructures.AcceptanceDataStore
	blockIndex            datastructures.BlockIndex
	blockMessageStore     datastructures.BlockMessageStore
	blockStatusStore      datastructures.BlockStatusStore
	consensusStateStore   datastructures.ConsensusStateStore
}

// New instantiates a new BlockProcessor
func New(
	dagParams *dagconfig.Params,
	databaseContext *dbaccess.DatabaseContext,
	consensusStateManager processes.ConsensusStateManager,
	pruningManager processes.PruningManager,
	blockValidator processes.BlockValidator,
	dagTopologyManager processes.DAGTopologyManager,
	reachabilityTree processes.ReachabilityTree,
	acceptanceDataStore datastructures.AcceptanceDataStore,
	blockIndex datastructures.BlockIndex,
	blockMessageStore datastructures.BlockMessageStore,
	blockStatusStore datastructures.BlockStatusStore,
	consensusStateStore datastructures.ConsensusStateStore) *BlockProcessor {

	return &BlockProcessor{
		dagParams:          dagParams,
		databaseContext:    databaseContext,
		pruningManager:     pruningManager,
		blockValidator:     blockValidator,
		dagTopologyManager: dagTopologyManager,
		reachabilityTree:   reachabilityTree,

		consensusStateManager: consensusStateManager,
		acceptanceDataStore:   acceptanceDataStore,
		blockIndex:            blockIndex,
		blockMessageStore:     blockMessageStore,
		blockStatusStore:      blockStatusStore,
		consensusStateStore:   consensusStateStore,
	}
}

// BuildBlock builds a block over the current state, with the transactions
// selected by the given transactionSelector
func (bp *BlockProcessor) BuildBlock(coinbaseScriptPublicKey []byte, coinbaseExtraData []byte,
	transactionSelector model.TransactionSelector) *appmessage.MsgBlock {

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

	return nil
}

// ValidateAndInsertBlock validates the given block and, if valid, applies it
// to the current state
func (bp *BlockProcessor) ValidateAndInsertBlock(block *appmessage.MsgBlock) error {
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
