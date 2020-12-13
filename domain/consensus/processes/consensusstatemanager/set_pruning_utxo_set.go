package consensusstatemanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"
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

func (csm *consensusStateManager) UpdatePruningPoint(newPruningPoint *externalapi.DomainHash, serializedUTXOSet []byte) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UpdatePruningPoint")
	defer onEnd()

	err := csm.updatePruningPoint(newPruningPoint, serializedUTXOSet)
	if err != nil {
		csm.discardSetPruningPointUTXOSetChanges()
		return err
	}

	return csm.commitSetPruningPointUTXOSetAll()
}

func (csm *consensusStateManager) updatePruningPoint(newPruningPoint *externalapi.DomainHash, serializedUTXOSet []byte) error {
	log.Tracef("updatePruningPoint start")
	defer log.Tracef("updatePruningPoint end")

	// We ignore the shouldSendNotification return value because we always want to send finality conflict notificaiton
	// in case the new pruning point violates finality
	isViolatingFinality, _, err := csm.isViolatingFinality(newPruningPoint)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		log.Warnf("Finality Violation Detected! The suggest pruning point %s violates finality!", newPruningPoint)
		return nil
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
	log.Tracef("Calculated multiset for given UTXO set: %s", utxoSetMultiSet.Hash())

	newPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, newPruningPoint)
	if err != nil {
		return err
	}
	log.Tracef("The UTXO commitment of the pruning point: %s",
		newPruningPointHeader.UTXOCommitment)

	if newPruningPointHeader.UTXOCommitment != *utxoSetMultiSet.Hash() {
		return errors.Wrapf(ruleerrors.ErrBadPruningPointUTXOSet, "the expected multiset hash of the pruning "+
			"point UTXO set is %s but got %s", newPruningPointHeader.UTXOCommitment, *utxoSetMultiSet.Hash())
	}
	log.Tracef("The new pruning point UTXO commitment validation passed")

	log.Tracef("Staging the parent hashes for pruning point as the DAG tips")
	csm.consensusStateStore.StageTips(newPruningPointHeader.ParentHashes)

	log.Tracef("Setting the parent hashes for the header tips pruning point as the virtual parents")
	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, newPruningPointHeader.ParentHashes)
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

	log.Tracef("Staging the new pruning point and its UTXO set")
	csm.pruningStore.Stage(newPruningPoint, serializedUTXOSet)

	log.Tracef("Staging the new pruning point as %s", externalapi.StatusValid)
	csm.blockStatusStore.Stage(newPruningPoint, externalapi.StatusValid)
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

func (p protoUTXOSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	entry, outpoint, err := utxo.DeserializeUTXO(p.utxoSet.Utxos[p.index].EntryOutpointPair)
	if err != nil {
		if serialization.IsMalformedError(err) {
			return nil, nil, errors.Wrap(ruleerrors.ErrMalformedUTXO, "malformed utxo")
		}
		return nil, nil, err
	}

	return outpoint, entry, nil
}

func protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet *utxoserialization.ProtoUTXOSet) model.ReadOnlyUTXOSetIterator {
	return &protoUTXOSetIterator{utxoSet: protoUTXOSet}
}
