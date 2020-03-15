// Copyright (c) 2013-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"fmt"
	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/pkg/errors"
	"time"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
)

// BehaviorFlags is a bitmask defining tweaks to the normal behavior when
// performing DAG processing and consensus rules checks.
type BehaviorFlags uint32

const (
	// BFFastAdd may be set to indicate that several checks can be avoided
	// for the block since it is already known to fit into the DAG due to
	// already proving it correct links into the DAG.
	BFFastAdd BehaviorFlags = 1 << iota

	// BFNoPoWCheck may be set to indicate the proof of work check which
	// ensures a block hashes to a value less than the required target will
	// not be performed.
	BFNoPoWCheck

	// BFWasUnorphaned may be set to indicate that a block was just now
	// unorphaned
	BFWasUnorphaned

	// BFAfterDelay may be set to indicate that a block had timestamp too far
	// in the future, just finished the delay
	BFAfterDelay

	// BFIsSync may be set to indicate that the block was sent as part of the
	// netsync process
	BFIsSync

	// BFWasStored is set to indicate that the block was previously stored
	// in the block index but was never fully processed
	BFWasStored

	// BFDisallowDelay is set to indicate that a delayed block should be rejected.
	// This is used for the case where a block is submitted through RPC.
	BFDisallowDelay

	// BFNone is a convenience value to specifically indicate no flags.
	BFNone BehaviorFlags = 0
)

// IsInDAG determines whether a block with the given hash exists in
// the DAG.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsInDAG(hash *daghash.Hash) bool {
	return dag.index.HaveBlock(hash)
}

// processOrphans determines if there are any orphans which depend on the passed
// block hash (they are no longer orphans if true) and potentially accepts them.
// It repeats the process for the newly accepted blocks (to detect further
// orphans which may no longer be orphans) until there are no more.
//
// The flags do not modify the behavior of this function directly, however they
// are needed to pass along to maybeAcceptBlock.
//
// This function MUST be called with the DAG state lock held (for writes).
func (dag *BlockDAG) processOrphans(hash *daghash.Hash, flags BehaviorFlags) error {
	// Start with processing at least the passed hash. Leave a little room
	// for additional orphan blocks that need to be processed without
	// needing to grow the array in the common case.
	processHashes := make([]*daghash.Hash, 0, 10)
	processHashes = append(processHashes, hash)
	for len(processHashes) > 0 {
		// Pop the first hash to process from the slice.
		processHash := processHashes[0]
		processHashes[0] = nil // Prevent GC leak.
		processHashes = processHashes[1:]

		// Look up all orphans that are parented by the block we just
		// accepted.  An indexing for loop is
		// intentionally used over a range here as range does not
		// reevaluate the slice on each iteration nor does it adjust the
		// index for the modified slice.
		for i := 0; i < len(dag.prevOrphans[*processHash]); i++ {
			orphan := dag.prevOrphans[*processHash][i]
			if orphan == nil {
				log.Warnf("Found a nil entry at index %d in the "+
					"orphan dependency list for block %s", i,
					processHash)
				continue
			}

			// Skip this orphan if one or more of its parents are
			// still missing.
			_, err := lookupParentNodes(orphan.block, dag)
			if err != nil {
				var ruleErr RuleError
				if ok := errors.As(err, &ruleErr); ok && ruleErr.ErrorCode == ErrParentBlockUnknown {
					continue
				}
				return err
			}

			// Remove the orphan from the orphan pool.
			orphanHash := orphan.block.Hash()
			dag.removeOrphanBlock(orphan)
			i--

			// Potentially accept the block into the block DAG.
			err = dag.maybeAcceptBlock(orphan.block, flags|BFWasUnorphaned)
			if err != nil {
				// Since we don't want to reject the original block because of
				// a bad unorphaned child, only return an error if it's not a RuleError.
				if !errors.As(err, &RuleError{}) {
					return err
				}
				log.Warnf("Verification failed for orphan block %s: %s", orphanHash, err)
			}

			// Add this block to the list of blocks to process so
			// any orphan blocks that depend on this block are
			// handled too.
			processHashes = append(processHashes, orphanHash)
		}
	}
	return nil
}

// ProcessBlock is the main workhorse for handling insertion of new blocks into
// the block DAG. It includes functionality such as rejecting duplicate
// blocks, ensuring blocks follow all rules, orphan handling, and insertion into
// the block DAG.
//
// When no errors occurred during processing, the first return value indicates
// whether or not the block is an orphan.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) ProcessBlock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	dag.dagLock.Lock()
	defer dag.dagLock.Unlock()
	return dag.processBlockNoLock(block, flags)
}

func (dag *BlockDAG) processBlockNoLock(block *util.Block, flags BehaviorFlags) (isOrphan bool, isDelayed bool, err error) {
	isAfterDelay := flags&BFAfterDelay == BFAfterDelay
	wasBlockStored := flags&BFWasStored == BFWasStored
	disallowDelay := flags&BFDisallowDelay == BFDisallowDelay

	blockHash := block.Hash()
	log.Tracef("Processing block %s", blockHash)

	// The block must not already exist in the DAG.
	if dag.IsInDAG(blockHash) && !wasBlockStored {
		str := fmt.Sprintf("already have block %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	// The block must not already exist as an orphan.
	if _, exists := dag.orphans[*blockHash]; exists {
		str := fmt.Sprintf("already have block (orphan) %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	if dag.isKnownDelayedBlock(blockHash) {
		str := fmt.Sprintf("already have block (delayed) %s", blockHash)
		return false, false, ruleError(ErrDuplicateBlock, str)
	}

	if !isAfterDelay {
		// Perform preliminary sanity checks on the block and its transactions.
		delay, err := dag.checkBlockSanity(block, flags)
		if err != nil {
			return false, false, err
		}

		if delay != 0 && disallowDelay {
			str := fmt.Sprintf("Cannot process blocks beyond the allowed time offset while the BFDisallowDelay flag is raised %s", blockHash)
			return false, true, ruleError(ErrDelayedBlockIsNotAllowed, str)
		}

		if delay != 0 {
			err = dag.addDelayedBlock(block, delay)
			if err != nil {
				return false, false, err
			}
			return false, true, nil
		}
	}

	var missingParents []*daghash.Hash
	for _, parentHash := range block.MsgBlock().Header.ParentHashes {
		if !dag.IsInDAG(parentHash) {
			missingParents = append(missingParents, parentHash)
		}
	}

	// Handle the case of a block with a valid timestamp(non-delayed) which points to a delayed block.
	delay, isParentDelayed := dag.maxDelayOfParents(missingParents)
	if isParentDelayed {
		// Add Nanosecond to ensure that parent process time will be after its child.
		delay += time.Nanosecond
		err := dag.addDelayedBlock(block, delay)
		if err != nil {
			return false, false, err
		}
		return false, true, err
	}

	// Handle orphan blocks.
	if len(missingParents) > 0 {
		// Some orphans during netsync are a normal part of the process, since the anticone
		// of the chain-split is never explicitly requested.
		// Therefore, if we are during netsync - don't report orphans to default logs.
		//
		// The number K*2 was chosen since in peace times anticone is limited to K blocks,
		// while some red block can make it a bit bigger, but much more than that indicates
		// there might be some problem with the netsync process.
		if flags&BFIsSync == BFIsSync && dagconfig.KType(len(dag.orphans)) < dag.dagParams.K*2 {
			log.Debugf("Adding orphan block %s. This is normal part of netsync process", blockHash)
		} else {
			log.Infof("Adding orphan block %s", blockHash)
		}
		dag.addOrphanBlock(block)

		return true, false, nil
	}

	// The block has passed all context independent checks and appears sane
	// enough to potentially accept it into the block DAG.
	err = dag.maybeAcceptBlock(block, flags)
	if err != nil {
		return false, false, err
	}

	// Accept any orphan blocks that depend on this block (they are
	// no longer orphans) and repeat for those accepted blocks until
	// there are no more.
	err = dag.processOrphans(blockHash, flags)
	if err != nil {
		return false, false, err
	}

	if !isAfterDelay {
		err = dag.processDelayedBlocks()
		if err != nil {
			return false, false, err
		}
	}

	log.Debugf("Accepted block %s", blockHash)

	return false, false, nil
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
