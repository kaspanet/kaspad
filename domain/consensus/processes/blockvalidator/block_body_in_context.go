package blockvalidator

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/virtual"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

// ValidateBodyInContext validates block bodies in the context of the current
// consensus state
func (v *blockValidator) ValidateBodyInContext(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, isBlockWithPrefilledData bool) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "ValidateBodyInContext")
	defer onEnd()

	err := v.checkBlockIsNotPruned(stagingArea, blockHash)
	if err != nil {
		return err
	}

	err = v.checkBlockTransactions(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if !isBlockWithPrefilledData {
		err := v.checkParentBlockBodiesExist(stagingArea, blockHash)
		if err != nil {
			return err
		}
	}
	return nil
}

// checkBlockIsNotPruned Checks we don't add block bodies to pruned blocks
func (v *blockValidator) checkBlockIsNotPruned(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {
	hasValidatedHeader, err := v.hasValidatedHeader(stagingArea, blockHash)
	if err != nil {
		return err
	}

	// If we don't add block body to a header only block it can't be in the past
	// of the tips, because it'll be a new tip.
	if !hasValidatedHeader {
		return nil
	}

	tips, err := v.consensusStateStore.Tips(stagingArea, v.databaseContext)
	if err != nil {
		return err
	}

	isAncestorOfSomeTips, err := v.dagTopologyManager.IsAncestorOfAny(stagingArea, blockHash, tips)
	if err != nil {
		return err
	}

	// A header only block in the past of one of the tips has to be pruned
	if isAncestorOfSomeTips {
		return errors.Wrapf(ruleerrors.ErrPrunedBlock, "cannot add block body to a pruned block %s", blockHash)
	}

	return nil
}

func (v *blockValidator) checkParentBlockBodiesExist(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {

	missingParentHashes := []*externalapi.DomainHash{}
	parents, err := v.dagTopologyManager.Parents(stagingArea, blockHash)
	if err != nil {
		return err
	}

	if virtual.ContainsOnlyVirtualGenesis(parents) {
		return nil
	}

	for _, parent := range parents {
		hasBlock, err := v.blockStore.HasBlock(v.databaseContext, stagingArea, parent)
		if err != nil {
			return err
		}

		if !hasBlock {
			pruningPoint, err := v.pruningStore.PruningPoint(v.databaseContext, stagingArea)
			if err != nil {
				return err
			}

			isInPastOfPruningPoint, err := v.dagTopologyManager.IsAncestorOf(stagingArea, parent, pruningPoint)
			if err != nil {
				return err
			}

			// If a block parent is in the past of the pruning point
			// it means its body will never be used, so it's ok if
			// it's missing.
			// This will usually happen during IBD when getting the blocks
			// in the pruning point anticone.
			if isInPastOfPruningPoint {
				log.Debugf("Block %s parent %s is missing a body, but is in the past of the pruning point",
					blockHash, parent)
				continue
			}

			log.Debugf("Block %s parent %s is missing a body", blockHash, parent)

			missingParentHashes = append(missingParentHashes, parent)
		}
	}

	if len(missingParentHashes) > 0 {
		return ruleerrors.NewErrMissingParents(missingParentHashes)
	}

	return nil
}

func (v *blockValidator) checkBlockTransactions(
	stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) error {

	block, err := v.blockStore.Block(v.databaseContext, stagingArea, blockHash)
	if err != nil {
		return err
	}

	// Ensure all transactions in the block are finalized.
	for _, tx := range block.Transactions {
		if err = v.transactionValidator.ValidateTransactionInContextIgnoringUTXO(stagingArea, tx, blockHash); err != nil {
			return err
		}
	}

	return nil
}
