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

func (f *FlowContext) UnorphanBlocks(blockHash *externalapi.DomainHash) error {
	f.orphansMutex.Lock()
	defer f.orphansMutex.Unlock()

	for orphanHash, orphanBlock := range f.orphans {
		canBeUnorphaned := true
		for _, orphanBlockParentHash := range orphanBlock.Header.ParentHashes {
			orphanBlockParentInfo, err := f.domain.Consensus().GetBlockInfo(orphanBlockParentHash)
			if err != nil {
				return err
			}
			if !orphanBlockParentInfo.Exists {
				canBeUnorphaned = false
				break
			}
		}
		if canBeUnorphaned {
			return f.unorphanBlock(orphanHash)
		}
	}
	return nil
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
	err = f.OnNewBlock(orphanBlock)
	if err != nil {
		return err
	}
	delete(f.orphans, orphanHash)

	log.Debugf("Unorphaned block %s", orphanHash)
	return nil
}
