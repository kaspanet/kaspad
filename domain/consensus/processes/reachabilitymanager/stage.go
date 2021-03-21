package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityManager) stageData(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, data model.ReachabilityData) {
	rt.reachabilityDataStore.StageReachabilityData(stagingArea, blockHash, data)
}

func (rt *reachabilityManager) stageFutureCoveringSet(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash, set model.FutureCoveringTreeNodeSet) error {
	data, err := rt.reachabilityDataForInsertion(stagingArea, blockHash)
	if err != nil {
		return err
	}

	data.SetFutureCoveringSet(set)

	rt.reachabilityDataStore.StageReachabilityData(stagingArea, blockHash, data)
	return nil
}

func (rt *reachabilityManager) stageReindexRoot(stagingArea *model.StagingArea, blockHash *externalapi.DomainHash) {
	rt.reachabilityDataStore.StageReachabilityReindexRoot(stagingArea, blockHash)
}

func (rt *reachabilityManager) stageAddChild(stagingArea *model.StagingArea, node, child *externalapi.DomainHash) error {
	nodeData, err := rt.reachabilityDataForInsertion(stagingArea, node)
	if err != nil {
		return err
	}

	nodeData.AddChild(child)
	rt.stageData(stagingArea, node, nodeData)

	return nil
}

func (rt *reachabilityManager) stageParent(stagingArea *model.StagingArea, node, parent *externalapi.DomainHash) error {
	nodeData, err := rt.reachabilityDataForInsertion(stagingArea, node)
	if err != nil {
		return err
	}
	nodeData.SetParent(parent)
	rt.stageData(stagingArea, node, nodeData)

	return nil
}

func (rt *reachabilityManager) stageInterval(stagingArea *model.StagingArea, node *externalapi.DomainHash, interval *model.ReachabilityInterval) error {
	nodeData, err := rt.reachabilityDataForInsertion(stagingArea, node)
	if err != nil {
		return err
	}
	nodeData.SetInterval(interval)
	rt.stageData(stagingArea, node, nodeData)

	return nil
}
