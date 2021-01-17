package consensusstatemanager

import (
	"github.com/golang/protobuf/proto"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/domain/consensus/utils/serialization"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxoserialization"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/pkg/errors"
	"runtime"
)

func (csm *consensusStateManager) UpdatePruningPoint(newPruningPoint *externalapi.DomainBlock, serializedUTXOSet []byte) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "UpdatePruningPoint")
	defer onEnd()

	err := csm.updatePruningPoint(newPruningPoint, serializedUTXOSet)
	runtime.GC() // Clear out all the used memory in `updatePruningPoint`
	if err != nil {
		csm.discardSetPruningPointUTXOSetChanges()
		return err
	}

	return csm.commitSetPruningPointUTXOSetAll()
}

func (csm *consensusStateManager) updatePruningPoint(newPruningPoint *externalapi.DomainBlock, serializedUTXOSet []byte) error {
	log.Debugf("updatePruningPoint start")
	defer log.Debugf("updatePruningPoint end")

	newPruningPointHash := consensushashing.BlockHash(newPruningPoint)

	// We ignore the shouldSendNotification return value because we always want to send finality conflict notification
	// in case the new pruning point violates finality
	isViolatingFinality, _, err := csm.isViolatingFinality(newPruningPointHash)
	if err != nil {
		return err
	}

	if isViolatingFinality {
		log.Warnf("Finality Violation Detected! The suggest pruning point %s violates finality!", newPruningPointHash)
		return errors.Wrapf(ruleerrors.ErrSuggestedPruningViolatesFinality, "%s cannot be a pruning point because "+
			"it violates finality", newPruningPointHash)
	}

	before := runtime.MemStats{}
	runtime.ReadMemStats(&before)

	protoUTXOSet := &utxoserialization.ProtoUTXOSet{}
	err = proto.Unmarshal(serializedUTXOSet, protoUTXOSet)
	if err != nil {
		return err
	}

	var size int
	for _, utxo := range protoUTXOSet.Utxos {
		size += len(utxo.EntryOutpointPair)
	}
	after := runtime.MemStats{}
	runtime.ReadMemStats(&after)

	log.Debugf("updatePruningPoint: utxoSize before deserializing: %d bytes. after, at least %d bytes.", len(serializedUTXOSet), size)
	log.Debugf("updatePruningPoint: before: used memory: %d bytes. total memory: %d bytes", before.Alloc, before.HeapIdle-before.HeapReleased+before.HeapInuse)
	log.Debugf("updatePruningPoint: after: used memory: %d bytes. total memory: %d bytes", after.Alloc, after.HeapIdle-after.HeapReleased+after.HeapInuse)

	utxoSetMultiSet, err := utxoserialization.CalculateMultisetFromProtoUTXOSet(protoUTXOSet)
	if err != nil {
		return err
	}
	log.Debugf("Calculated multiset for given UTXO set: %s", utxoSetMultiSet.Hash())

	newPruningPointHeader, err := csm.blockHeaderStore.BlockHeader(csm.databaseContext, newPruningPointHash)
	if err != nil {
		return err
	}
	log.Debugf("The UTXO commitment of the pruning point: %s",
		newPruningPointHeader.UTXOCommitment())

	if !newPruningPointHeader.UTXOCommitment().Equal(utxoSetMultiSet.Hash()) {
		return errors.Wrapf(ruleerrors.ErrBadPruningPointUTXOSet, "the expected multiset hash of the pruning "+
			"point UTXO set is %s but got %s", newPruningPointHeader.UTXOCommitment(), *utxoSetMultiSet.Hash())
	}
	log.Debugf("The new pruning point UTXO commitment validation passed")

	newTips := []*externalapi.DomainHash{newPruningPointHash}

	log.Debugf("Staging the the pruning point as the only DAG tip")
	csm.consensusStateStore.StageTips(newTips)

	log.Debugf("Setting the pruning point as the only virtual parent")
	err = csm.dagTopologyManager.SetParents(model.VirtualBlockHash, newTips)
	if err != nil {
		return err
	}

	log.Debugf("Calculating GHOSTDAG for the new virtual")
	err = csm.ghostdagManager.GHOSTDAG(model.VirtualBlockHash)
	if err != nil {
		return err
	}

	log.Debugf("Staging the virtual UTXO set")
	err = csm.consensusStateStore.StageVirtualUTXOSet(protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet))
	if err != nil {
		return err
	}
	protoUTXOSet = nil

	log.Debugf("Deleting all the existing virtual diff parents")
	csm.consensusStateStore.StageVirtualDiffParents(nil)

	log.Debugf("Updating the new pruning point to be the new virtual diff parent with an empty diff")
	err = csm.stageDiff(newPruningPointHash, utxo.NewUTXODiff(), nil)
	if err != nil {
		return err
	}

	log.Debugf("Staging the new pruning point and its UTXO set")
	csm.pruningStore.StagePruningPoint(newPruningPointHash, serializedUTXOSet)

	// Before we manually mark the new pruning point as valid, we validate that all of its transactions are valid
	// against the provided UTXO set.
	log.Debugf("Validating that the pruning point is UTXO valid")

	// validateBlockTransactionsAgainstPastUTXO pre-fills the block's transactions inputs, which
	// are assumed to not be pre-filled during further validations.
	// Therefore - clone newPruningPoint before passing it to validateBlockTransactionsAgainstPastUTXO
	err = csm.validateBlockTransactionsAgainstPastUTXO(newPruningPoint.Clone(), utxo.NewUTXODiff())
	if err != nil {
		return err
	}

	log.Debugf("Staging the new pruning point as %s", externalapi.StatusUTXOValid)
	csm.blockStatusStore.Stage(newPruningPointHash, externalapi.StatusUTXOValid)

	log.Debugf("Staging the new pruning point multiset")
	csm.multisetStore.Stage(newPruningPointHash, utxoSetMultiSet)
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

	stats := runtime.MemStats{}
	for _, store := range csm.stores {
		runtime.ReadMemStats(&stats)
		log.Debugf("Committing store: %T, used memory: %d bytes, total: %d bytes\n", store, stats.Alloc, stats.HeapIdle-stats.HeapReleased+stats.HeapInuse)
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

func (p *protoUTXOSetIterator) Next() bool {
	p.index++
	return p.index < len(p.utxoSet.Utxos)
}

func (p *protoUTXOSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	entry, outpoint, err := utxo.DeserializeUTXO(p.utxoSet.Utxos[p.index].EntryOutpointPair)
	if err != nil {
		if serialization.IsMalformedError(err) {
			return nil, nil, errors.Wrapf(ruleerrors.ErrMalformedUTXO, "malformed utxo: %s", err)
		}
		return nil, nil, err
	}

	return outpoint, entry, nil
}

func protoUTXOSetToReadOnlyUTXOSetIterator(protoUTXOSet *utxoserialization.ProtoUTXOSet) model.ReadOnlyUTXOSetIterator {
	return &protoUTXOSetIterator{utxoSet: protoUTXOSet, index: -1}
}
