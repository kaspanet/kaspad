package pruningmanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxoserialization"
)

// pruningManager resolves and manages the current pruning point
type pruningManager struct {
	databaseContext model.DBReader

	dagTraversalManager   model.DAGTraversalManager
	dagTopologyManager    model.DAGTopologyManager
	consensusStateManager model.ConsensusStateManager
	consensusStateStore   model.ConsensusStateStore
	ghostdagDataStore     model.GHOSTDAGDataStore
	pruningStore          model.PruningStore
	blockStatusStore      model.BlockStatusStore

	multiSetStore       model.MultisetStore
	acceptanceDataStore model.AcceptanceDataStore
	blocksStore         model.BlockStore
	utxoDiffStore       model.UTXODiffStore

	genesisHash      *externalapi.DomainHash
	finalityInterval uint64
	pruningDepth     uint64
}

// New instantiates a new PruningManager
func New(
	databaseContext model.DBReader,

	dagTraversalManager model.DAGTraversalManager,
	dagTopologyManager model.DAGTopologyManager,
	consensusStateManager model.ConsensusStateManager,
	consensusStateStore model.ConsensusStateStore,
	ghostdagDataStore model.GHOSTDAGDataStore,
	pruningStore model.PruningStore,
	blockStatusStore model.BlockStatusStore,

	multiSetStore model.MultisetStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blocksStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,

	genesisHash *externalapi.DomainHash,
	finalityInterval uint64,
	pruningDepth uint64,
) model.PruningManager {

	return &pruningManager{
		databaseContext:       databaseContext,
		dagTraversalManager:   dagTraversalManager,
		dagTopologyManager:    dagTopologyManager,
		consensusStateManager: consensusStateManager,
		consensusStateStore:   consensusStateStore,
		ghostdagDataStore:     ghostdagDataStore,
		pruningStore:          pruningStore,
		blockStatusStore:      blockStatusStore,
		multiSetStore:         multiSetStore,
		acceptanceDataStore:   acceptanceDataStore,
		blocksStore:           blocksStore,
		utxoDiffStore:         utxoDiffStore,
		genesisHash:           genesisHash,
		pruningDepth:          pruningDepth,
		finalityInterval:      finalityInterval,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash
func (pm *pruningManager) FindNextPruningPoint() error {
	hasPruningPoint, err := pm.pruningStore.HasPruningPoint(pm.databaseContext)
	if err != nil {
		return err
	}

	if !hasPruningPoint {
		err = pm.savePruningPoint(pm.genesisHash)
		if err != nil {
			return err
		}
		return nil
	}

	currentP, err := pm.pruningStore.PruningPoint(pm.databaseContext)
	if err != nil {
		return err
	}

	virtual, err := pm.ghostdagDataStore.Get(pm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	currentPGhost, err := pm.ghostdagDataStore.Get(pm.databaseContext, currentP)
	if err != nil {
		return err
	}
	currentPBlueScore := currentPGhost.BlueScore
	// Because the pruning point changes only once per finality, then there's no need to even check for that if a finality interval hasn't passed.
	if virtual.BlueScore <= currentPBlueScore+pm.finalityInterval {
		return nil
	}

	// This means the pruning point is still genesis.
	if virtual.BlueScore <= pm.pruningDepth+pm.finalityInterval {
		return nil
	}

	// get Virtual(pruningDepth)
	candidatePHash, err := pm.dagTraversalManager.BlockAtDepth(model.VirtualBlockHash, pm.pruningDepth)
	if err != nil {
		return err
	}
	candidatePGhost, err := pm.ghostdagDataStore.Get(pm.databaseContext, candidatePHash)
	if err != nil {
		return err
	}

	// Actually check if the pruning point changed
	if (currentPBlueScore / pm.finalityInterval) < (candidatePGhost.BlueScore / pm.finalityInterval) {
		err = pm.savePruningPoint(candidatePHash)
		if err != nil {
			return err
		}
		return pm.deletePastBlocks(candidatePHash)
	}
	return pm.deletePastBlocks(currentP)
}

func (pm *pruningManager) deletePastBlocks(pruningPoint *externalapi.DomainHash) error {
	// Go over all P.Past and P.AC that's not in V.Past
	queue := pm.dagTraversalManager.NewDownHeap()

	// Find P.AC that's not in V.Past
	dagTips, err := pm.consensusStateStore.Tips(pm.databaseContext)
	if err != nil {
		return err
	}
	for _, tip := range dagTips {
		hasPruningPointInPast, err := pm.dagTopologyManager.IsAncestorOf(pruningPoint, tip)
		if err != nil {
			return err
		}
		if !hasPruningPointInPast {
			isInVirtualPast, err := pm.dagTopologyManager.IsAncestorOf(model.VirtualBlockHash, tip)
			if err != nil {
				return err
			}
			if !isInVirtualPast {
				// Add them to the queue so they and their past will be pruned
				err := queue.Push(tip)
				if err != nil {
					return err
				}
			}
		}
	}

	// Add P.Parents
	parents, err := pm.dagTopologyManager.Parents(pruningPoint)
	if err != nil {
		return err
	}
	for _, parent := range parents {
		err = queue.Push(parent)
		if err != nil {
			return err
		}
	}

	visited := map[externalapi.DomainHash]struct{}{}
	// Prune everything in the queue including its past
	for queue.Len() > 0 {
		current := queue.Pop()
		if _, ok := visited[*current]; ok {
			continue
		}
		visited[*current] = struct{}{}

		alreadyPruned, err := pm.deleteBlock(current)
		if err != nil {
			return err
		}
		if !alreadyPruned {
			parents, err := pm.dagTopologyManager.Parents(current)
			if err != nil {
				return err
			}
			for _, parent := range parents {
				err = queue.Push(parent)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (pm *pruningManager) savePruningPoint(blockHash *externalapi.DomainHash) error {
	utxoIter, err := pm.consensusStateManager.RestorePastUTXOSetIterator(blockHash)
	if err != nil {
		return err
	}
	serializedUtxo, err := serializeUTXOSetIterator(utxoIter)
	if err != nil {
		return err
	}
	pm.pruningStore.Stage(blockHash, serializedUtxo)

	return nil
}

func (pm *pruningManager) deleteBlock(blockHash *externalapi.DomainHash) (alreadyPruned bool, err error) {
	status, err := pm.blockStatusStore.Get(pm.databaseContext, blockHash)
	if err != nil {
		return false, err
	}
	if status == externalapi.StatusHeaderOnly {
		return true, nil
	}

	pm.multiSetStore.Delete(blockHash)
	pm.acceptanceDataStore.Delete(blockHash)
	pm.blocksStore.Delete(blockHash)
	pm.utxoDiffStore.Delete(blockHash)

	pm.blockStatusStore.Stage(blockHash, externalapi.StatusHeaderOnly)
	return false, nil
}

func serializeUTXOSetIterator(iter model.ReadOnlyUTXOSetIterator) ([]byte, error) {
	serializedUtxo, err := utxoserialization.ReadOnlyUTXOSetToProtoUTXOSet(iter)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(serializedUtxo)
}
