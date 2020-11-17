package flowcontext

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/pkg/errors"
)

func (f *FlowContext) AddOrphan(orphanBlock *externalapi.DomainBlock) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	orphanHash := consensusserialization.BlockHash(orphanBlock)
	f.orphans[*orphanHash] = orphanBlock
}

func (f *FlowContext) UnorphanBlocks(rootBlock *externalapi.DomainBlock) ([]*externalapi.DomainBlock, error) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	// Find all the children of rootBlock among the orphans
	// and add them to the process queue
	rootBlockHash := consensusserialization.BlockHash(rootBlock)
	processQueue := f.findOrphansOfParentBlock(*rootBlockHash)

	var unorphanedBlocks []*externalapi.DomainBlock
	for len(processQueue) > 0 {
		var orphanHash externalapi.DomainHash
		orphanHash, processQueue = processQueue[0], processQueue[1:]
		orphanBlock := f.orphans[orphanHash]

		log.Tracef("Considering to unorphan block %s with parents", orphanHash, orphanBlock.Header.ParentHashes)
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

			// Add the orphans of the block that had just been
			// unorphaned to the process queue unless they already
			// appear in it
			orphansOfUnorphanedBlock := f.findOrphansOfParentBlock(orphanHash)
			for _, candidateOrphan := range orphansOfUnorphanedBlock {
				exists := false
				for _, queueOrphan := range processQueue {
					if queueOrphan == candidateOrphan {
						exists = true
						break
					}
				}
				if !exists {
					processQueue = append(processQueue, candidateOrphan)
				}
			}
		}
	}

	return unorphanedBlocks, nil
}

func (f *FlowContext) findOrphansOfParentBlock(blockHash externalapi.DomainHash) []externalapi.DomainHash {
	var orphansOfParentBlock []externalapi.DomainHash
	for orphanHash, orphanBlock := range f.orphans {
		for _, orphanBlockParentHash := range orphanBlock.Header.ParentHashes {
			if *orphanBlockParentHash == blockHash {
				orphansOfParentBlock = append(orphansOfParentBlock, orphanHash)
				break
			}
		}
	}
	return orphansOfParentBlock
}

func (f *FlowContext) unorphanBlock(orphanHash externalapi.DomainHash) error {
	orphanBlock, ok := f.orphans[orphanHash]
	if !ok {
		return errors.Errorf("attempted to unorphan a non-orphan block %s", orphanHash)
	}
	err := f.domain.Consensus().ValidateAndInsertBlock(orphanBlock)
	if err != nil {
		return err
	}
	delete(f.orphans, orphanHash)

	log.Debugf("Unorphaned block %s", orphanHash)
	return nil
}
