package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/database"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"
)

var utxoSetBucket = database.MakeBucket([]byte("virtual-utxo-set"))

func utxoKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return utxoSetBucket.Key(serializedOutpoint), nil
}

func (css *consensusStateStore) StageVirtualUTXODiff(stagingArea *model.StagingArea, virtualUTXODiff externalapi.UTXODiff) {
	stagingShard := css.stagingShard(stagingArea)

	stagingShard.virtualUTXODiffStaging = virtualUTXODiff
}

func (csss *consensusStateStagingShard) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
	hadStartedImportingPruningPointUTXOSet, err := csss.store.HadStartedImportingPruningPointUTXOSet(dbTx)
	if err != nil {
		return err
	}
	if hadStartedImportingPruningPointUTXOSet {
		return errors.New("cannot commit virtual UTXO diff after starting to import the pruning point UTXO set")
	}

	if csss.virtualUTXODiffStaging == nil {
		return nil
	}

	toRemoveIterator := csss.virtualUTXODiffStaging.ToRemove().Iterator()
	defer toRemoveIterator.Close()
	for ok := toRemoveIterator.First(); ok; ok = toRemoveIterator.Next() {
		toRemoveOutpoint, _, err := toRemoveIterator.Get()
		if err != nil {
			return err
		}

		csss.store.virtualUTXOSetCache.Remove(toRemoveOutpoint)

		dbKey, err := utxoKey(toRemoveOutpoint)
		if err != nil {
			return err
		}
		err = dbTx.Delete(dbKey)
		if err != nil {
			return err
		}
	}

	toAddIterator := csss.virtualUTXODiffStaging.ToAdd().Iterator()
	defer toAddIterator.Close()
	for ok := toAddIterator.First(); ok; ok = toAddIterator.Next() {
		toAddOutpoint, toAddEntry, err := toAddIterator.Get()
		if err != nil {
			return err
		}

		csss.store.virtualUTXOSetCache.Add(toAddOutpoint, toAddEntry)

		dbKey, err := utxoKey(toAddOutpoint)
		if err != nil {
			return err
		}
		serializedEntry, err := serializeUTXOEntry(toAddEntry)
		if err != nil {
			return err
		}
		err = dbTx.Put(dbKey, serializedEntry)
		if err != nil {
			return err
		}
	}

	// Note: we don't discard the staging here since that's
	// being done at the end of Commit()
	return nil
}

func (css *consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, stagingArea *model.StagingArea,
	outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, error) {

	stagingShard := css.stagingShard(stagingArea)

	return css.utxoByOutpointFromStagedVirtualUTXODiff(dbContext, stagingShard, outpoint)
}

func (css *consensusStateStore) utxoByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	stagingShard *consensusStateStagingShard, outpoint *externalapi.DomainOutpoint) (externalapi.UTXOEntry, error) {

	if stagingShard.virtualUTXODiffStaging != nil {
		if stagingShard.virtualUTXODiffStaging.ToRemove().Contains(outpoint) {
			return nil, errors.Errorf("outpoint was not found")
		}
		if utxoEntry, ok := stagingShard.virtualUTXODiffStaging.ToAdd().Get(outpoint); ok {
			return utxoEntry, nil
		}
	}

	if entry, ok := css.virtualUTXOSetCache.Get(outpoint); ok {
		return entry, nil
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return nil, err
	}

	serializedUTXOEntry, err := dbContext.Get(key)
	if err != nil {
		return nil, err
	}

	entry, err := deserializeUTXOEntry(serializedUTXOEntry)
	if err != nil {
		return nil, err
	}

	css.virtualUTXOSetCache.Add(outpoint, entry)
	return entry, nil
}

func (css *consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, stagingArea *model.StagingArea,
	outpoint *externalapi.DomainOutpoint) (bool, error) {

	stagingShard := css.stagingShard(stagingArea)

	return css.hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext, stagingShard, outpoint)
}

func (css *consensusStateStore) hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	stagingShard *consensusStateStagingShard, outpoint *externalapi.DomainOutpoint) (bool, error) {

	if stagingShard.virtualUTXODiffStaging != nil {
		if stagingShard.virtualUTXODiffStaging.ToRemove().Contains(outpoint) {
			return false, nil
		}
		if _, ok := stagingShard.virtualUTXODiffStaging.ToAdd().Get(outpoint); ok {
			return true, nil
		}
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return false, err
	}

	return dbContext.Has(key)
}

func (css *consensusStateStore) VirtualUTXOs(dbContext model.DBReader, fromOutpoint *externalapi.DomainOutpoint, limit int) (
	[]*externalapi.OutpointAndUTXOEntryPair, error) {

	cursor, err := dbContext.Cursor(utxoSetBucket)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	if fromOutpoint != nil {
		serializedFromOutpoint, err := serializeOutpoint(fromOutpoint)
		if err != nil {
			return nil, err
		}
		seekKey := utxoSetBucket.Key(serializedFromOutpoint)
		err = cursor.Seek(seekKey)
		if err != nil {
			return nil, err
		}
	}

	iterator := newCursorUTXOSetIterator(cursor)
	defer iterator.Close()

	outpointAndUTXOEntryPairs := make([]*externalapi.OutpointAndUTXOEntryPair, 0, limit)
	for len(outpointAndUTXOEntryPairs) < limit && iterator.Next() {
		outpoint, utxoEntry, err := iterator.Get()
		if err != nil {
			return nil, err
		}
		outpointAndUTXOEntryPairs = append(outpointAndUTXOEntryPairs, &externalapi.OutpointAndUTXOEntryPair{
			Outpoint:  outpoint,
			UTXOEntry: utxoEntry,
		})
	}
	return outpointAndUTXOEntryPairs, nil
}

func (css *consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader, stagingArea *model.StagingArea) (
	externalapi.ReadOnlyUTXOSetIterator, error) {

	stagingShard := css.stagingShard(stagingArea)

	cursor, err := dbContext.Cursor(utxoSetBucket)
	if err != nil {
		return nil, err
	}

	mainIterator := newCursorUTXOSetIterator(cursor)
	if stagingShard.virtualUTXODiffStaging != nil {
		return utxo.IteratorWithDiff(mainIterator, stagingShard.virtualUTXODiffStaging)
	}

	return mainIterator, nil
}

type utxoSetIterator struct {
	cursor   model.DBCursor
	isClosed bool
}

func newCursorUTXOSetIterator(cursor model.DBCursor) externalapi.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u utxoSetIterator) First() bool {
	if u.isClosed {
		panic("Tried using a closed utxoSetIterator")
	}
	return u.cursor.First()
}

func (u utxoSetIterator) Next() bool {
	if u.isClosed {
		panic("Tried using a closed utxoSetIterator")
	}
	return u.cursor.Next()
}

func (u utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
	if u.isClosed {
		return nil, nil, errors.New("Tried using a closed utxoSetIterator")
	}
	key, err := u.cursor.Key()
	if err != nil {
		panic(err)
	}

	utxoEntryBytes, err := u.cursor.Value()
	if err != nil {
		return nil, nil, err
	}

	outpoint, err = deserializeOutpoint(key.Suffix())
	if err != nil {
		return nil, nil, err
	}

	utxoEntry, err = deserializeUTXOEntry(utxoEntryBytes)
	if err != nil {
		return nil, nil, err
	}

	return outpoint, utxoEntry, nil
}

func (u utxoSetIterator) Close() error {
	if u.isClosed {
		return errors.New("Tried using a closed utxoSetIterator")
	}
	u.isClosed = true
	err := u.cursor.Close()
	if err != nil {
		return err
	}
	u.cursor = nil
	return nil
}
