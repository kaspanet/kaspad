package flowcontext

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/pkg/errors"
)

// AddOrphan adds the block to the orphan set
func (f *FlowContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	orphanHash := consensusserialization.BlockHash(orphanBlock)
	f.orphans[*orphanHash] = orphanBlock

	log.Infof("Received a block with missing parents, adding to orphan pool: %s", orphanHash)
}

// UnorphanBlocks removes the block from the orphan set, and remove all of the blocks that are not orphans anymore.
func (f *FlowContext) UnorphanBlocks(rootBlock *externalapi.DomainBlock) ([]*externalapi.DomainBlock, error) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	// Find all the children of rootBlock among the orphans
	// and add them to the process queue
	rootBlockHash := consensusserialization.BlockHash(rootBlock)
	processQueue := f.addChildOrphansToProcessQueue(rootBlockHash, []externalapi.DomainHash{})

	var unorphanedBlocks []*externalapi.DomainBlock
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
			if !orphanBlockParentInfo.Exists {
				log.Tracef("Cannot unorphan block %s. It's missing at "+
					"least the following parent: %s", orphanHash, orphanBlockParentHash)

				canBeUnorphaned = false
				break
			}
		}
		if canBeUnorphaned {
			err := f.unorphanBlock(orphanHash)
			if err != nil {
				return nil, err
			}
			unorphanedBlocks = append(unorphanedBlocks, orphanBlock)
			processQueue = f.addChildOrphansToProcessQueue(&orphanHash, processQueue)
		}
	}

	return unorphanedBlocks, nil
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
			if *orphanBlockParentHash == *blockHash {
				childOrphans = append(childOrphans, orphanHash)
				break
			}
		}
	}
	return childOrphans
}

func (f *FlowContext) unorphanBlock(orphanHash externalapi.DomainHash) error {
	orphanBlock, ok := f.orphans[orphanHash]
	if !ok {
		return errors.Errorf("attempted to unorphan a non-orphan block %s", orphanHash)
	}
	delete(f.orphans, orphanHash)

	err := f.domain.Consensus().ValidateAndInsertBlock(orphanBlock)
	if err != nil {
		if errors.As(err, &ruleerrors.RuleError{}) {
			log.Infof("Validation failed for orphan block %s: %s", orphanHash, err)
			return nil
		}
		return err
	}

	log.Infof("Unorphaned block %s", orphanHash)
	return nil
}
