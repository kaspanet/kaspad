package consensusstatemanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var virtualHeaderHash = &externalapi.DomainHash{
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
}

func (csm *consensusStateManager) SetPruningPointUTXOSet(serializedUTXOSet []byte) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "SetPruningPointUTXOSet")
	defer onEnd()

	err := csm.setPruningPointUTXOSet(serializedUTXOSet)
	if err != nil {
		csm.discardSetPruningPointUTXOSetChanges()
		return err
	}

	return csm.commitSetPruningPointUTXOSetAll()
}

func (csm *consensusStateManager) setPruningPointUTXOSet(serializedUTXOSet []byte) error {
	headerTipsPruningPoint, err := csm.headerTipsPruningPoint()
	if err != nil {
		return err
	}

	utxoSetIterator := deserializeUTXOSet(serializedUTXOSet)
	utxoSetMultiSet, err := calcMultisetFromUTXOSetIterator(utxoSetIterator)
	if err != nil {
		return err
	}

	headerTipsPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, headerTipsPruningPoint)
	if err != nil {
		return err
	}

	if headerTipsPruningPointHeader.UTXOCommitment != *utxoSetMultiSet.Hash() {
		return errors.Wrapf(ruleerrors.ErrBadPruningPointUTXOSet, "the expected multiset hash of the pruning "+
			"point UTXO set is %s but got %s", headerTipsPruningPointHeader.UTXOCommitment, *utxoSetMultiSet.Hash())
	}

	csm.consensusStateStore.StageTips(headerTipsPruningPointHeader.ParentHashes)
	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, headerTipsPruningPointHeader.ParentHashes)
	if err != nil {
		return err
	}

	csm.consensusStateStore.StageVirtualUTXOSet(utxoSetIterator)

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = csm.updateVirtualDiffParents(headerTipsPruningPoint, model.NewUTXODiff())
	if err != nil {
		return err
	}

	csm.blockStatusStore.Stage(headerTipsPruningPoint, externalapi.StatusValid)
	return nil
}

func (csm *consensusStateManager) discardSetPruningPointUTXOSetChanges() {
	for _, store := range csm.stores {
		store.Discard()
	}
}

func (csm *consensusStateManager) commitSetPruningPointUTXOSetAll() error {
	dbTx, err := csm.databaseContext.Begin()
	if err != nil {
		return err
	}

	for _, store := range csm.stores {
		err = store.Commit(dbTx)
		if err != nil {
			return err
		}
	}

	return dbTx.Commit()
}

func deserializeUTXOSet(serializedUTXOSet []byte) model.ReadOnlyUTXOSetIterator {
	panic("implement me")
}

func (csm *consensusStateManager) headerTipsPruningPoint() (*externalapi.DomainHash, error) {
	headerTips, err := csm.headerTipsStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	csm.blockRelationStore.StageBlockRelation(virtualHeaderHash, &model.BlockRelations{
		Parents: headerTips,
	})
	defer csm.blockRelationStore.Discard()

	err = csm.ghostdagManager.GHOSTDAG(virtualHeaderHash)
	if err != nil {
		return nil, err
	}
	defer csm.ghostdagDataStore.Discard()

	virtualHeaderGHOSTDAGData, err := csm.ghostdagDataStore.Get(csm.databaseContext, virtualHeaderHash)
	if err != nil {
		return nil, err
	}

	return csm.dagTraversalManager.HighestChainBlockBelowBlueScore(virtualHeaderHash, virtualHeaderGHOSTDAGData.BlueScore-pruningDepth())
}

func pruningDepth() uint64 {
	panic("unimplemented")
}
