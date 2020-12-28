package flowcontext

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/pkg/errors"
)

// maxOrphans is the maximum amount of orphans allowed in the
// orphans collection. This number is an approximation of how
// many orphans there can possibly be on average. It is based
// on: 2^orphanResolutionRange * PHANTOM K.
const maxOrphans = 600

// UnorphaningResult is the result of unorphaning a block
type UnorphaningResult struct {
	block                *externalapi.DomainBlock
	blockInsertionResult *externalapi.BlockInsertionResult
}

// AddOrphan adds the block to the orphan set
func (f *FlowContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	orphanHash := consensushashing.BlockHash(orphanBlock)
	f.orphans[*orphanHash] = orphanBlock

	if len(f.orphans) > maxOrphans {
		log.Debugf("Orphan collection size exceeded. Evicting a random orphan")
		f.evictRandomOrphan()
	}

	log.Infof("Received a block with missing parents, adding to orphan pool: %s", orphanHash)
}

func (f *FlowContext) evictRandomOrphan() {
	var toEvict externalapi.DomainHash
	for hash := range f.orphans {
		toEvict = hash
		break
	}
	delete(f.orphans, toEvict)
	log.Debugf("Evicted %s from the orphan collection", toEvict)
}

// IsOrphan returns whether the given blockHash belongs to an orphan block
func (f *FlowContext) IsOrphan(blockHash *externalapi.DomainHash) bool {
	f.orphansMutex.RLock()
	defer f.orphansMutex.RUnlock()

	_, ok := f.orphans[*blockHash]
	return ok
}

// UnorphanBlocks removes the block from the orphan set, and remove all of the blocks that are not orphans anymore.
func (f *FlowContext) UnorphanBlocks(rootBlock *externalapi.DomainBlock) ([]*UnorphaningResult, error) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	// Find all the children of rootBlock among the orphans
	// and add them to the process queue
	rootBlockHash := consensushashing.BlockHash(rootBlock)
	processQueue := f.addChildOrphansToProcessQueue(rootBlockHash, []externalapi.DomainHash{})

	var unorphaningResults []*UnorphaningResult
	for len(processQueue) > 0 {
		var orphanHash externalapi.DomainHash
		orphanHash, processQueue = processQueue[0], processQueue[1:]
		orphanBlock := f.orphans[orphanHash]

		log.Tracef("Considering to unorphan block %s with parents %s",
			orphanHash, orphanBlock.Header.ParentHashes)

		canBeUnorphaned := true
		for _, orphanBlockParentHash := range orphanBlock.Header.ParentHashes {
			orphanBlockParentInfo, err := f.domain.Consensus().GetBlockInfo(orphanBlockParentHash)
			if err != nil {
				return nil, err
			}
			if !orphanBlockParentInfo.Exists || orphanBlockParentInfo.BlockStatus == externalapi.StatusHeaderOnly {
				log.Tracef("Cannot unorphan block %s. It's missing at "+
					"least the following parent: %s", orphanHash, orphanBlockParentHash)

				canBeUnorphaned = false
				break
			}
		}
		if canBeUnorphaned {
			blockInsertionResult, err := f.unorphanBlock(orphanHash)
			if err != nil {
				return nil, err
			}
			unorphaningResults = append(unorphaningResults, &UnorphaningResult{
				block:                orphanBlock,
				blockInsertionResult: blockInsertionResult,
			})
			processQueue = f.addChildOrphansToProcessQueue(&orphanHash, processQueue)
		}
	}

	return unorphaningResults, nil
}

// addChildOrphansToProcessQueue finds all child orphans of `blockHash`
// and adds them to the given `processQueue` if they don't already exist
// inside of it
// Note that this method does not modify the given `processQueue`
func (f *FlowContext) addChildOrphansToProcessQueue(blockHash *externalapi.DomainHash,
	processQueue []externalapi.DomainHash) []externalapi.DomainHash {

	blockChildren := f.findChildOrphansOfBlock(blockHash)
	for _, blockChild := range blockChildren {
		exists := false
		for _, queueOrphan := range processQueue {
			if queueOrphan == blockChild {
				exists = true
				break
			}
		}
		if !exists {
			processQueue = append(processQueue, blockChild)
		}
	}
	return processQueue
}

func (f *FlowContext) findChildOrphansOfBlock(blockHash *externalapi.DomainHash) []externalapi.DomainHash {
	var childOrphans []externalapi.DomainHash
	for orphanHash, orphanBlock := range f.orphans {
		for _, orphanBlockParentHash := range orphanBlock.Header.ParentHashes {
			if orphanBlockParentHash.Equal(blockHash) {
				childOrphans = append(childOrphans, orphanHash)
				break
			}
		}
	}
	return childOrphans
}

func (f *FlowContext) unorphanBlock(orphanHash externalapi.DomainHash) (*externalapi.BlockInsertionResult, error) {
	orphanBlock, ok := f.orphans[orphanHash]
	if !ok {
		return nil, errors.Errorf("attempted to unorphan a non-orphan block %s", orphanHash)
	}
	delete(f.orphans, orphanHash)

	blockInsertionResult, err := f.domain.Consensus().ValidateAndInsertBlock(orphanBlock)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Warnf("Validation failed for orphan block %s: %s", orphanHash, err)
			return nil, nil
		}
		return nil, err
	}
	f.updateRecentBlockAddedTimesWithLastBlock()

	log.Infof("Unorphaned block %s", orphanHash)
	return blockInsertionResult, nil
}
