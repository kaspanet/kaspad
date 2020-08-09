package blockdag

import (
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
	"time"
)

// maxOrphanBlocks is the maximum number of orphan blocks that can be
// queued.
const maxOrphanBlocks = 100

// orphanBlock represents a block that we don't yet have the parent for. It
// is a normal block plus an expiration time to prevent caching the orphan
// forever.
type orphanBlock struct {
	block      *util.Block
	expiration mstime.Time
}

// IsKnownOrphan returns whether the passed hash is currently a known orphan.
// Keep in mind that only a limited number of orphans are held onto for a
// limited amount of time, so this function must not be used as an absolute
// way to test if a block is an orphan block. A full block (as opposed to just
// its hash) must be passed to ProcessBlock for that purpose. However, calling
// ProcessBlock with an orphan that already exists results in an error, so this
// function provides a mechanism for a caller to intelligently detect *recent*
// duplicate orphans and react accordingly.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) IsKnownOrphan(hash *daghash.Hash) bool {
	// Protect concurrent access. Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	defer dag.orphanLock.RUnlock()
	_, exists := dag.orphans[*hash]

	return exists
}

// GetOrphanMissingAncestorHashes returns all of the missing parents in the orphan's sub-DAG
//
// This function is safe for concurrent access.
func (dag *BlockDAG) GetOrphanMissingAncestorHashes(orphanHash *daghash.Hash) []*daghash.Hash {
	// Protect concurrent access. Using a read lock only so multiple
	// readers can query without blocking each other.
	dag.orphanLock.RLock()
	defer dag.orphanLock.RUnlock()

	missingAncestorsHashes := make([]*daghash.Hash, 0)

	visited := make(map[daghash.Hash]bool)
	queue := []*daghash.Hash{orphanHash}
	for len(queue) > 0 {
		var current *daghash.Hash
		current, queue = queue[0], queue[1:]
		if !visited[*current] {
			visited[*current] = true
			orphan, orphanExists := dag.orphans[*current]
			if orphanExists {
				queue = append(queue, orphan.block.MsgBlock().Header.ParentHashes...)
			} else {
				if !dag.IsInDAG(current) && current != orphanHash {
					missingAncestorsHashes = append(missingAncestorsHashes, current)
				}
			}
		}
	}
	return missingAncestorsHashes
}

// removeOrphanBlock removes the passed orphan block from the orphan pool and
// previous orphan index.
func (dag *BlockDAG) removeOrphanBlock(orphan *orphanBlock) {
	// Protect concurrent access.
	dag.orphanLock.Lock()
	defer dag.orphanLock.Unlock()

	// Remove the orphan block from the orphan pool.
	orphanHash := orphan.block.Hash()
	delete(dag.orphans, *orphanHash)

	// Remove the reference from the previous orphan index too.
	for _, parentHash := range orphan.block.MsgBlock().Header.ParentHashes {
		// An indexing for loop is intentionally used over a range here as range
		// does not reevaluate the slice on each iteration nor does it adjust the
		// index for the modified slice.
		orphans := dag.prevOrphans[*parentHash]
		for i := 0; i < len(orphans); i++ {
			hash := orphans[i].block.Hash()
			if hash.IsEqual(orphanHash) {
				orphans = append(orphans[:i], orphans[i+1:]...)
				i--
			}
		}

		// Remove the map entry altogether if there are no longer any orphans
		// which depend on the parent hash.
		if len(orphans) == 0 {
			delete(dag.prevOrphans, *parentHash)
			continue
		}

		dag.prevOrphans[*parentHash] = orphans
	}
}

// addOrphanBlock adds the passed block (which is already determined to be
// an orphan prior calling this function) to the orphan pool. It lazily cleans
// up any expired blocks so a separate cleanup poller doesn't need to be run.
// It also imposes a maximum limit on the number of outstanding orphan
// blocks and will remove the oldest received orphan block if the limit is
// exceeded.
func (dag *BlockDAG) addOrphanBlock(block *util.Block) {
	// Remove expired orphan blocks.
	for _, oBlock := range dag.orphans {
		if mstime.Now().After(oBlock.expiration) {
			dag.removeOrphanBlock(oBlock)
			continue
		}

		// Update the newest orphan block pointer so it can be discarded
		// in case the orphan pool fills up.
		if dag.newestOrphan == nil || oBlock.block.Timestamp().After(dag.newestOrphan.block.Timestamp()) {
			dag.newestOrphan = oBlock
		}
	}

	// Limit orphan blocks to prevent memory exhaustion.
	if len(dag.orphans)+1 > maxOrphanBlocks {
		// If the new orphan is newer than the newest orphan on the orphan
		// pool, don't add it.
		if block.Timestamp().After(dag.newestOrphan.block.Timestamp()) {
			return
		}
		// Remove the newest orphan to make room for the added one.
		dag.removeOrphanBlock(dag.newestOrphan)
		dag.newestOrphan = nil
	}

	// Protect concurrent access. This is intentionally done here instead
	// of near the top since removeOrphanBlock does its own locking and
	// the range iterator is not invalidated by removing map entries.
	dag.orphanLock.Lock()
	defer dag.orphanLock.Unlock()

	// Insert the block into the orphan map with an expiration time
	// 1 hour from now.
	expiration := mstime.Now().Add(time.Hour)
	oBlock := &orphanBlock{
		block:      block,
		expiration: expiration,
	}
	dag.orphans[*block.Hash()] = oBlock

	// Add to parent hash lookup index for faster dependency lookups.
	for _, parentHash := range block.MsgBlock().Header.ParentHashes {
		dag.prevOrphans[*parentHash] = append(dag.prevOrphans[*parentHash], oBlock)
	}
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
