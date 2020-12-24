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

	dagTraversalManager    model.DAGTraversalManager
	dagTopologyManager     model.DAGTopologyManager
	consensusStateManager  model.ConsensusStateManager
	consensusStateStore    model.ConsensusStateStore
	ghostdagDataStore      model.GHOSTDAGDataStore
	pruningStore           model.PruningStore
	blockStatusStore       model.BlockStatusStore
	headerSelectedTipStore model.HeaderSelectedTipStore

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
	headerSelectedTipStore model.HeaderSelectedTipStore,

	multiSetStore model.MultisetStore,
	acceptanceDataStore model.AcceptanceDataStore,
	blocksStore model.BlockStore,
	utxoDiffStore model.UTXODiffStore,

	genesisHash *externalapi.DomainHash,
	finalityInterval uint64,
	pruningDepth uint64,
) model.PruningManager {

	return &pruningManager{
		databaseContext:        databaseContext,
		dagTraversalManager:    dagTraversalManager,
		dagTopologyManager:     dagTopologyManager,
		consensusStateManager:  consensusStateManager,
		consensusStateStore:    consensusStateStore,
		ghostdagDataStore:      ghostdagDataStore,
		pruningStore:           pruningStore,
		blockStatusStore:       blockStatusStore,
		multiSetStore:          multiSetStore,
		acceptanceDataStore:    acceptanceDataStore,
		blocksStore:            blocksStore,
		utxoDiffStore:          utxoDiffStore,
		headerSelectedTipStore: headerSelectedTipStore,
		genesisHash:            genesisHash,
		pruningDepth:           pruningDepth,
		finalityInterval:       finalityInterval,
	}
}

// FindNextPruningPoint finds the next pruning point from the
// given blockHash
func (pm *pruningManager) UpdatePruningPointByVirtual() error {
	hasPruningPoint, err := pm.pruningStore.HasPruningPoint(pm.databaseContext)
	if err != nil {
		return err
	}

	if !hasPruningPoint {
		err = pm.savePruningPoint(pm.genesisHash)
		if err != nil {
			return err
		}
	}

	currentCandidate, err := pm.pruningPointCandidate()
	if err != nil {
		return err
	}

	currentCandidateGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, currentCandidate)
	if err != nil {
		return err
	}

	virtual, err := pm.ghostdagDataStore.Get(pm.databaseContext, model.VirtualBlockHash)
	if err != nil {
		return err
	}

	virtualSelectedParent, err := pm.ghostdagDataStore.Get(pm.databaseContext, virtual.SelectedParent())
	if err != nil {
		return err
	}

	currentPruningPoint, err := pm.pruningStore.PruningPoint(pm.databaseContext)
	if err != nil {
		return err
	}

	currentPruningPointGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, currentPruningPoint)
	if err != nil {
		return err
	}

	iterator, err := pm.dagTraversalManager.SelectedChildIterator(virtual.SelectedParent(), currentCandidate)
	if err != nil {
		return err
	}

	// Finding the next pruning point candidate: look for the latest
	// selected child of the current candidate that is in depth of at
	// least pm.pruningDepth blocks from the virtual selected parent.
	//
	// Note: Sometimes the current candidate is less than pm.pruningDepth
	// from the virtual. This can happen only if the virtual blue score
	// got smaller, because virtual blue score is not guaranteed to always
	// increase (because sometimes a block with higher blue work can have
	// lower blue score).
	// In such cases we still keep the same candidate because it's guaranteed
	// that a block that was once in depth of pm.pruningDepth cannot be
	// reorged without causing a finality conflict first.
	newCandidate := currentCandidate
	newCandidateGHOSTDAGData := currentCandidateGHOSTDAGData

	newPruningPoint := currentPruningPoint
	newPruningPointGHOSTDAGData := currentPruningPointGHOSTDAGData
	for iterator.Next() {
		selectedChild := iterator.Get()
		selectedChildGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, selectedChild)
		if err != nil {
			return err
		}

		if virtualSelectedParent.BlueScore()-selectedChildGHOSTDAGData.BlueScore() < pm.pruningDepth {
			break
		}

		newCandidate = selectedChild
		newCandidateGHOSTDAGData = selectedChildGHOSTDAGData

		// We move the pruning point every time the candidate's finality score is
		// bigger than the current pruning point finality score.
		if pm.finalityScore(newCandidateGHOSTDAGData.BlueScore()) > pm.finalityScore(newPruningPointGHOSTDAGData.BlueScore()) {
			newPruningPoint = newCandidate
			newPruningPointGHOSTDAGData = newCandidateGHOSTDAGData
		}
	}

	if !newCandidate.Equal(currentCandidate) {
		pm.pruningStore.StagePruningPointCandidate(newCandidate)
	}

	// We move the pruning point every time the candidate's finality score is
	// bigger than the current pruning point finality score.
	if pm.finalityScore(newCandidateGHOSTDAGData.BlueScore()) <= pm.finalityScore(currentPruningPointGHOSTDAGData.BlueScore()) {
		return nil
	}

	if !newPruningPoint.Equal(currentPruningPoint) {
		err = pm.savePruningPoint(newPruningPoint)
		if err != nil {
			return err
		}
		return pm.deletePastBlocks(newPruningPoint)
	}

	return nil
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
			virtualParents, err := pm.dagTopologyManager.Parents(model.VirtualBlockHash)
			if err != nil {
				return err
			}

			// Because virtual doesn't have reachability data, we need to check reachability
			// using it parents.
			isInVirtualPast, err := pm.dagTopologyManager.IsAncestorOfAny(tip, virtualParents)
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
	pm.pruningStore.StagePruningPoint(blockHash, serializedUtxo)

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

func (pm *pruningManager) IsValidPruningPoint(block *externalapi.DomainHash) (bool, error) {
	if *pm.genesisHash == *block {
		return true, nil
	}

	headersSelectedTip, err := pm.headerSelectedTipStore.HeadersSelectedTip(pm.databaseContext)
	if err != nil {
		return false, err
	}

	// A pruning point has to be in the selected chain of the headers selected tip.
	headersSelectedTipGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, headersSelectedTip)
	if err != nil {
		return false, err
	}

	isInSelectedParentChainOfHeadersSelectedTip, err := pm.dagTopologyManager.IsInSelectedParentChainOf(block,
		headersSelectedTip)
	if err != nil {
		return false, err
	}

	if !isInSelectedParentChainOfHeadersSelectedTip {
		return false, nil
	}

	ghostdagData, err := pm.ghostdagDataStore.Get(pm.databaseContext, block)
	if err != nil {
		return false, err
	}

	// A pruning point has to be at depth of at least pm.pruningDepth
	if headersSelectedTipGHOSTDAGData.BlueScore()-ghostdagData.BlueScore() < pm.pruningDepth {
		return false, nil
	}

	selectedParentGHOSTDAGData, err := pm.ghostdagDataStore.Get(pm.databaseContext, ghostdagData.SelectedParent())
	if err != nil {
		return false, err
	}

	// A pruning point has to be the lowest chain block with a certain finality score, so
	// if the block selected parent has the same finality score it means it cannot be a
	// pruning point.
	if pm.finalityScore(ghostdagData.BlueScore()) == pm.finalityScore(selectedParentGHOSTDAGData.BlueScore()) {
		return false, nil
	}

	return true, nil
}

func (pm *pruningManager) pruningPointCandidate() (*externalapi.DomainHash, error) {
	hasPruningPointCandidate, err := pm.pruningStore.HasPruningPointCandidate(pm.databaseContext)
	if err != nil {
		return nil, err
	}

	if !hasPruningPointCandidate {
		return pm.genesisHash, nil
	}

	return pm.pruningStore.PruningPointCandidate(pm.databaseContext)
}

func serializeUTXOSetIterator(iter model.ReadOnlyUTXOSetIterator) ([]byte, error) {
	serializedUtxo, err := utxoserialization.ReadOnlyUTXOSetToProtoUTXOSet(iter)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(serializedUtxo)
}

// finalityScore is the number of finality intervals passed since
// the given block.
func (pm *pruningManager) finalityScore(blueScore uint64) uint64 {
	return blueScore / pm.finalityInterval
}
