package finalitymanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

type finalityManager struct {
	dagTopologyManager  model.DAGTopologyManager
	dagTraversalManager model.DAGTraversalManager
	finalityStore       model.FinalityStore
	genesisHash         *externalapi.DomainHash
	finalityDepth       uint64
}

func New(dagTopologyManager model.DAGTopologyManager,
	dagTraversalManager model.DAGTraversalManager,
	finalityStore model.FinalityStore,
	genesisHash *externalapi.DomainHash,
	finalityDepth uint64) model.FinalityManager {

	return &finalityManager{
		genesisHash:         genesisHash,
		dagTopologyManager:  dagTopologyManager,
		finalityStore:       finalityStore,
		dagTraversalManager: dagTraversalManager,
		finalityDepth:       finalityDepth,
	}
}

func (fm *finalityManager) IsViolatingFinality(blockHash *externalapi.DomainHash) (bool, error) {
	if *blockHash == *fm.genesisHash {
		log.Tracef("Block %s is the genesis block, "+
			"and does not violate finality by definition", blockHash)
		return false, nil
	}
	log.Tracef("isViolatingFinality start for block %s", blockHash)
	defer log.Tracef("isViolatingFinality end for block %s", blockHash)

	virtualFinalityPoint, err := fm.VirtualFinalityPoint()
	if err != nil {
		return false, err
	}
	log.Tracef("The virtual finality point is: %s", virtualFinalityPoint)

	isInSelectedParentChain, err := fm.dagTopologyManager.IsInSelectedParentChainOf(virtualFinalityPoint, blockHash)
	if err != nil {
		return false, err
	}
	log.Tracef("Is the virtual finality point %s "+
		"in the selected parent chain of %s: %t", virtualFinalityPoint, blockHash, isInSelectedParentChain)

	return !isInSelectedParentChain, nil
}

func (fm *finalityManager) VirtualFinalityPoint() (*externalapi.DomainHash, error) {
	log.Tracef("virtualFinalityPoint start")
	defer log.Tracef("virtualFinalityPoint end")

	virtualFinalityPoint, err := fm.dagTraversalManager.BlockAtDepth(
		model.VirtualBlockHash, fm.finalityDepth)
	if err != nil {
		return nil, err
	}
	log.Tracef("The current virtual finality block is: %s", virtualFinalityPoint)

	return virtualFinalityPoint, nil
}

func (fm *finalityManager) FinalityPoint(blockHash *externalapi.DomainHash) (*externalapi.DomainHash, error) {
	return fm.dagTraversalManager.BlockAtDepth(blockHash, fm.finalityDepth)
}
