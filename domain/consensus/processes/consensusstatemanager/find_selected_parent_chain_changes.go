package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (csm *consensusStateManager) GetVirtualSelectedParentChainFromBlock(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) (*externalapi.SelectedChainPath, error) {

	// Calculate chain changes between the given blockHash and the
	// virtual's selected parent. Note that we explicitly don't
	// do the calculation against the virtual itself so that we
	// won't later need to remove it from the result.
	virtualGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, stagingArea, model.VirtualBlockHash, false)
	if err != nil {
		return nil, err
	}
	virtualSelectedParent := virtualGHOSTDAGData.SelectedParent()

	return csm.dagTraversalManager.CalculateChainPath(stagingArea, blockHash, virtualSelectedParent)
}
