package consensusstatemanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxoserialization"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
)

var virtualHeaderHash = &externalapi.DomainHash{
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xfe,
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
	headerTipsPruningPoint, err := csm.HeaderTipsPruningPoint()
	if err != nil {
		return err
	}

	protoUTXOSet := &utxoserialization.ProtoUTXOSet{}
	err = proto.Unmarshal(serializedUTXOSet, protoUTXOSet)
	if err != nil {
		return err
	}

	utxoSetMultiSet, err := calcMultisetFromProtoUTXOSet(protoUTXOSet)
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

	err = csm.consensusStateStore.StageTips(headerTipsPruningPointHeader.ParentHashes)
	if err != nil {
		return err
	}

	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, headerTipsPruningPointHeader.ParentHashes)
	if err != nil {
		return err
	}

	err = csm.consensusStateStore.StageVirtualUTXOSet(protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet))
	if err != nil {
		return err
	}

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	err = csm.updateVirtualDiffParents(model.NewUTXODiff())
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

type protoUTXOSetIterator struct {
	utxoSet *utxoserialization.ProtoUTXOSet
	index   int
}

func (p protoUTXOSetIterator) Next() bool {
	p.index++
	return p.index < len(p.utxoSet.Utxos)
}

func (p protoUTXOSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry, err error) {
	entry, outpoint, err := consensusserialization.DeserializeUTXO(p.utxoSet.Utxos[p.index].EntryOutpointPair)
	if err != nil {
		if consensusserialization.IsMalformedError(err) {
			return nil, nil, errors.Wrap(ruleerrors.ErrMalformedUTXO, "malformed utxo")
		}
		return nil, nil, err
	}

	return outpoint, entry, nil
}

func protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet *utxoserialization.ProtoUTXOSet) model.ReadOnlyUTXOSetIterator {
	return &protoUTXOSetIterator{utxoSet: protoUTXOSet}
}

func (csm *consensusStateManager) HeaderTipsPruningPoint() (*externalapi.DomainHash, error) {
	headerTips, err := csm.headerTipsStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}

	err = csm.blockRelationStore.StageBlockRelation(virtualHeaderHash, &model.BlockRelations{
		Parents: headerTips,
	})
	if err != nil {
		return nil, err
	}

	defer csm.blockRelationStore.Discard()

	err = csm.ghostdagManager.GHOSTDAG(virtualHeaderHash)
	if err != nil {
		return nil, err
	}
	defer csm.ghostdagDataStore.Discard()

	return csm.dagTraversalManager.BlockAtDepth(virtualHeaderHash, csm.pruningDepth)
}
