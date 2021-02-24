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

func (css *consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff externalapi.UTXODiff) {
	css.virtualUTXODiffStaging = virtualUTXODiff
}

func (css *consensusStateStore) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
	hadStartedImportingPruningPointUTXOSet, err := css.HadStartedImportingPruningPointUTXOSet(dbTx)
	if err != nil {
		return err
	}
	if hadStartedImportingPruningPointUTXOSet {
		return errors.New("cannot commit virtual UTXO diff after starting to import the pruning point UTXO set")
	}

	if css.virtualUTXODiffStaging == nil {
		return nil
	}

	toRemoveIterator := css.virtualUTXODiffStaging.ToRemove().Iterator()
	for ok := toRemoveIterator.First(); ok; ok = toRemoveIterator.Next() {
		toRemoveOutpoint, _, err := toRemoveIterator.Get()
		if err != nil {
			return err
		}

		css.virtualUTXOSetCache.Remove(toRemoveOutpoint)

		dbKey, err := utxoKey(toRemoveOutpoint)
		if err != nil {
			return err
		}
		err = dbTx.Delete(dbKey)
		if err != nil {
			return err
		}
	}

	toAddIterator := css.virtualUTXODiffStaging.ToAdd().Iterator()
	for ok := toAddIterator.First(); ok; ok = toAddIterator.Next() {
		toAddOutpoint, toAddEntry, err := toAddIterator.Get()
		if err != nil {
			return err
		}

		css.virtualUTXOSetCache.Add(toAddOutpoint, toAddEntry)

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

func (css *consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (
	externalapi.UTXOEntry, error) {

	return css.utxoByOutpointFromStagedVirtualUTXODiff(dbContext, outpoint)
}

func (css *consensusStateStore) utxoByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	outpoint *externalapi.DomainOutpoint) (
	externalapi.UTXOEntry, error) {

	if css.virtualUTXODiffStaging != nil {
		if css.virtualUTXODiffStaging.ToRemove().Contains(outpoint) {
			return nil, errors.Errorf("outpoint was not found")
		}
		if utxoEntry, ok := css.virtualUTXODiffStaging.ToAdd().Get(outpoint); ok {
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

func (css *consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	return css.hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext, outpoint)
}

func (css *consensusStateStore) hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	outpoint *externalapi.DomainOutpoint) (bool, error) {

	if css.virtualUTXODiffStaging != nil {
		if css.virtualUTXODiffStaging.ToRemove().Contains(outpoint) {
			return false, nil
		}
		if _, ok := css.virtualUTXODiffStaging.ToAdd().Get(outpoint); ok {
			return true, nil
		}
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return false, err
	}

	return dbContext.Has(key)
}

func (css *consensusStateStore) VirtualUTXOs(dbContext model.DBReader,
	fromOutpoint *externalapi.DomainOutpoint, limit int) ([]*externalapi.OutpointAndUTXOEntryPair, error) {

	cursor, err := dbContext.Cursor(utxoSetBucket)
	if err != nil {
		return nil, err
	}

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

func (css *consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader) (externalapi.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(utxoSetBucket)
	if err != nil {
		return nil, err
	}

	mainIterator := newCursorUTXOSetIterator(cursor)
	if css.virtualUTXODiffStaging != nil {
		return utxo.IteratorWithDiff(mainIterator, css.virtualUTXODiffStaging)
	}

	return mainIterator, nil
}

type utxoSetIterator struct {
	cursor model.DBCursor
}

func newCursorUTXOSetIterator(cursor model.DBCursor) externalapi.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u utxoSetIterator) First() bool {
	return u.cursor.First()
}

func (u utxoSetIterator) Next() bool {
	return u.cursor.Next()
}

func (u utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry externalapi.UTXOEntry, err error) {
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
