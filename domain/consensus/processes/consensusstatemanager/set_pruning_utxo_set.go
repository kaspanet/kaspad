package consensusstatemanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensusserialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
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
	log.Tracef("setPruningPointUTXOSet start")
	defer log.Tracef("setPruningPointUTXOSet end")

	headerTipsPruningPoint, err := csm.HeaderTipsPruningPoint()
	if err != nil {
		return err
	}
	log.Tracef("The pruning point of the header tips is: %s", headerTipsPruningPoint)

	protoUTXOSet := &utxoserialization.ProtoUTXOSet{}
	err = proto.Unmarshal(serializedUTXOSet, protoUTXOSet)
	if err != nil {
		return err
	}

	utxoSetMultiSet, err := calcMultisetFromProtoUTXOSet(protoUTXOSet)
	if err != nil {
		return err
	}
	log.Tracef("Calculated multiset for given UTXO set: %s", utxoSetMultiSet.Hash())

	headerTipsPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, headerTipsPruningPoint)
	if err != nil {
		return err
	}
	log.Tracef("The multiset in the header of the header tip pruning point: %s",
		headerTipsPruningPointHeader.UTXOCommitment)

	if headerTipsPruningPointHeader.UTXOCommitment != *utxoSetMultiSet.Hash() {
		return errors.Wrapf(ruleerrors.ErrBadPruningPointUTXOSet, "the expected multiset hash of the pruning "+
			"point UTXO set is %s but got %s", headerTipsPruningPointHeader.UTXOCommitment, *utxoSetMultiSet.Hash())
	}
	log.Tracef("Header tip pruning point multiset validation passed")

	log.Tracef("Staging the parent hashes for the header tips pruning point as the DAG tips")
	csm.consensusStateStore.StageTips(headerTipsPruningPointHeader.ParentHashes)

	log.Tracef("Setting the parent hashes for the header tips pruning point as the virtual parents")
	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, headerTipsPruningPointHeader.ParentHashes)
	if err != nil {
		return err
	}

	log.Tracef("Staging the virtual UTXO set")
	err = csm.consensusStateStore.StageVirtualUTXOSet(protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet))
	if err != nil {
		return err
	}

	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Tracef("Updating the header tips pruning point diff parents with an empty UTXO diff")
	err = csm.updateVirtualDiffParents(utxo.NewUTXODiff())
	if err != nil {
		return err
	}

	log.Tracef("Staging the status of the header tips pruning point as %s", externalapi.StatusValid)
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
	log.Tracef("HeaderTipsPruningPoint start")
	defer log.Tracef("HeaderTipsPruningPoint end")

	headerTips, err := csm.headerTipsStore.Tips(csm.databaseContext)
	if err != nil {
		return nil, err
	}
	log.Tracef("The current header tips are: %s", headerTips)

	log.Tracef("Temporarily staging the parents of the virtual header to be the header tips: %s", headerTips)
	csm.blockRelationStore.StageBlockRelation(virtualHeaderHash, &model.BlockRelations{
		Parents: headerTips,
	})

	defer csm.blockRelationStore.Discard()

	err = csm.ghostdagManager.GHOSTDAG(virtualHeaderHash)
	if err != nil {
		return nil, err
	}
	defer csm.ghostdagDataStore.Discard()

	pruningPoint, err := csm.dagTraversalManager.BlockAtDepth(virtualHeaderHash, csm.pruningDepth)
	if err != nil {
		return nil, err
	}
	log.Tracef("The block at depth %d from %s is: %s", csm.pruningDepth, virtualHeaderHash, pruningPoint)
	return pruningPoint, nil
}
