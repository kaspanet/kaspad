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

func (f *FlowContext) UnorphanBlocks() ([]*externalapi.DomainBlock, error) {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	unorphanedBlocks := make([]*externalapi.DomainBlock, 0)
	for orphanHash, orphanBlock := range f.orphans {
		canBeUnorphaned := true
		for _, orphanBlockParentHash := range orphanBlock.Header.ParentHashes {
			orphanBlockParentInfo, err := f.domain.Consensus().GetBlockInfo(orphanBlockParentHash)
			if err != nil {
				return nil, err
			}
			if !orphanBlockParentInfo.Exists {
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
		}
	}
	return unorphanedBlocks, nil
}

func (f *FlowContext) unorphanBlock(orphanHash externalapi.DomainHash) error {
	orphanBlock, ok := f.orphans[orphanHash]
	if !ok {
		return errors.Errorf("attempted to unorphan a non-orphan block")
	}
	err := f.domain.Consensus().ValidateAndInsertBlock(orphanBlock)
	if err != nil {
		return err
	}
	delete(f.orphans, orphanHash)

	log.Debugf("Unorphaned block %s", orphanHash)
	return nil
}
