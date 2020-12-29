package reachabilitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

func (rt *reachabilityManager) stageData(blockHash *externalapi.DomainHash, data model.ReadOnlyReachabilityData) {
	rt.reachabilityDataStore.StageReachabilityData(blockHash, data)
}

func (rt *reachabilityManager) stageFutureCoveringSet(blockHash *externalapi.DomainHash, set model.FutureCoveringTreeNodeSet) error {
	data, err := rt.reachabilityDataForInsertion(blockHash)
	if err != nil {
		return err
	}

	data.SetFutureCoveringSet(set)

	rt.reachabilityDataStore.StageReachabilityData(blockHash, data)
	return nil
}

func (rt *reachabilityManager) stageReindexRoot(blockHash *externalapi.DomainHash) {
	rt.reachabilityDataStore.StageReachabilityReindexRoot(blockHash)
}

func (rt *reachabilityManager) addChildAndStage(node, child *externalapi.DomainHash) error {
	nodeData, err := rt.reachabilityDataForInsertion(node)
	if err != nil {
		return err
	}

	nodeData.AddChild(child)
	rt.stageData(node, nodeData)

	return nil
}

func (rt *reachabilityManager) stageParent(node, parent *externalapi.DomainHash) error {
	nodeData, err := rt.reachabilityDataForInsertion(node)
	if err != nil {
		return err
	}
	nodeData.SetParent(parent)
	rt.stageData(node, nodeData)

	return nil
}

func (rt *reachabilityManager) stageInterval(node *externalapi.DomainHash, interval *model.ReachabilityInterval) error {
	nodeData, err := rt.reachabilityDataForInsertion(node)
	if err != nil {
		return err
	}
	nodeData.SetInterval(interval)
	rt.stageData(node, nodeData)

	return nil
}
