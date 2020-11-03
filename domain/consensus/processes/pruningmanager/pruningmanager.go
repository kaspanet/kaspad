package pruningmanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/constants"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
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

	pruningDepth     uint64
	finalityInterval uint64
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

	finalityInterval uint64,
	k model.KType,
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
		pruningDepth:          pruningDepth(uint64(k), finalityInterval, constants.MergeSetSizeLimit),
		finalityInterval:      finalityInterval,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash
func (pm *pruningManager) FindNextPruningPoint() error {
	virtual, err := pm.ghostdagDataStore.Get(pm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	currentP, err := pm.PruningPoint()
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
	candidatePHash, err := pm.dagTraversalManager.HighestChainBlockBelowBlueScore(model.VirtualBlockHash, pm.pruningDepth)
	if err != nil {
		return err
	}
	candidatePGhost, err := pm.ghostdagDataStore.Get(pm.databaseContext, candidatePHash)
	if err != nil {
		return err
	}

	// Actually check if the pruning point changed
	if (currentPBlueScore / pm.finalityInterval) < (candidatePGhost.BlueScore / pm.finalityInterval) {
		utxoIter, err := pm.consensusStateManager.RestorePastUTXOSetIterator(candidatePHash)
		if err != nil {
			return err
		}
		serializedUtxo, err := serializeUTXOSetIterator(utxoIter)
		if err != nil {
			return err
		}
		pm.pruningStore.Stage(candidatePHash, serializedUtxo)
		currentP = candidatePHash
	}
	return pm.deletePastBlocks(currentP)
}

// PruningPoint returns the hash of the current pruning point
func (pm *pruningManager) PruningPoint() (*externalapi.DomainHash, error) {
	pruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext)
	if err != nil {
		if !pm.pruningStore.HasPruningPoint(pm.databaseContext) {
			genesis, err := pm.dagTraversalManager.HighestChainBlockBelowBlueScore(model.VirtualBlockHash, pm.pruningDepth+pm.finalityInterval)
			if err != nil {
				return nil, err
			}
			utxoGenesisIter, err := pm.consensusStateManager.RestorePastUTXOSetIterator(genesis)
			if err != nil {
				return nil, err
			}
			serializedUtxo, err := serializeUTXOSetIterator(utxoGenesisIter)
			if err != nil {
				return nil, err
			}
			pm.pruningStore.Stage(genesis, serializedUtxo)
			pruningPoint = genesis
		} else {
			return nil, err
		}

	}
	return pruningPoint, nil
}

// SerializedUTXOSet returns the serialized UTXO set of the
// current pruning point
func (pm *pruningManager) SerializedUTXOSet() ([]byte, error) {
	return pm.pruningStore.PruningPointSerializedUTXOSet(pm.databaseContext)
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

func pruningDepth(k, finalityInterval, mergeSetSizeLimit uint64) uint64 {
	return 2*finalityInterval + 4*mergeSetSizeLimit*k + 2*k + 2
}

func serializeUTXOSetIterator(iter model.ReadOnlyUTXOSetIterator) ([]byte, error) {
	serializedUtxo, err := hashserialization.ReadOnlyUTXOSetToProtoUTXOSet(iter)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(serializedUtxo)
}
