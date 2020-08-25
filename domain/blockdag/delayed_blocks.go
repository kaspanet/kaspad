package blockdag

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"time"
)

// delayedBlock represents a block which has a delayed timestamp and will be processed at processTime
type delayedBlock struct {
	block       *util.Block
	processTime mstime.Time
	flags       BehaviorFlags
}

func (dag *BlockDAG) isKnownDelayedBlock(hash *daghash.Hash) bool {
	_, exists := dag.delayedBlocks[*hash]
	return exists
}

func (dag *BlockDAG) addDelayedBlock(block *util.Block, flags BehaviorFlags, delay time.Duration) error {
	processTime := dag.Now().Add(delay)
	log.Debugf("Adding block to delayed blocks queue (block hash: %s, process time: %s)", block.Hash().String(), processTime)
	delayedBlock := &delayedBlock{
		block:       block,
		processTime: processTime,
		flags:       flags,
	}

	dag.delayedBlocks[*block.Hash()] = delayedBlock
	dag.delayedBlocksQueue.Push(delayedBlock)
	return dag.processDelayedBlocks()
}

// processDelayedBlocks loops over all delayed blocks and processes blocks which are due.
// This method is invoked after processing a block (ProcessBlock method).
func (dag *BlockDAG) processDelayedBlocks() error {
	// Check if the delayed block with the earliest process time should be processed
	for dag.delayedBlocksQueue.Len() > 0 {
		earliestDelayedBlockProcessTime := dag.peekDelayedBlock().processTime
		if earliestDelayedBlockProcessTime.After(dag.Now()) {
			break
		}
		delayedBlock := dag.popDelayedBlock()
		_, _, err := dag.processBlockNoLock(delayedBlock.block, delayedBlock.flags|BFAfterDelay)
		if err != nil {
			log.Errorf("Error while processing delayed block (block %s): %s", delayedBlock.block.Hash().String(), err)
			// Rule errors should not be propagated as they refer only to the delayed block,
			// while this function runs in the context of another block
			if !errors.As(err, &RuleError{}) {
				return err
			}
		}
		log.Debugf("Processed delayed block (block %s)", delayedBlock.block.Hash().String())
	}

	return nil
}

// popDelayedBlock removes the topmost (delayed block with earliest process time) of the queue and returns it.
func (dag *BlockDAG) popDelayedBlock() *delayedBlock {
	delayedBlock := dag.delayedBlocksQueue.pop()
	delete(dag.delayedBlocks, *delayedBlock.block.Hash())
	return delayedBlock
}

func (dag *BlockDAG) peekDelayedBlock() *delayedBlock {
	return dag.delayedBlocksQueue.peek()
}

// maxDelayOfParents returns the maximum delay of the given block hashes.
// Note that delay could be 0, but isDelayed will return true. This is the case where the parent process time is due.
func (dag *BlockDAG) maxDelayOfParents(parentHashes []*daghash.Hash) (delay time.Duration, isDelayed bool) {
	for _, parentHash := range parentHashes {
		if delayedParent, exists := dag.delayedBlocks[*parentHash]; exists {
			isDelayed = true
			parentDelay := delayedParent.processTime.Sub(dag.Now())
			if parentDelay > delay {
				delay = parentDelay
			}
		}
	}

	return delay, isDelayed
}

func (dag *BlockDAG) shouldBlockBeDelayed(block *util.Block) (delay time.Duration, isDelayed bool) {
	header := &block.MsgBlock().Header

	maxTimestamp := dag.Now().Add(time.Duration(dag.TimestampDeviationTolerance) * dag.Params.TargetTimePerBlock)
	if header.Timestamp.After(maxTimestamp) {
		return header.Timestamp.Sub(maxTimestamp), true
	}
	return 0, false
}
