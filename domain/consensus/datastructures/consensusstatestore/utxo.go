package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/pkg/errors"
)

var utxoSetBucket = dbkeys.MakeBucket([]byte("virtual-utxo-set"))

func utxoKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := serializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return utxoSetBucket.Key(serializedOutpoint), nil
}

func (css *consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff model.UTXODiff) {
	css.virtualUTXODiffStaging = virtualUTXODiff
}

func (css *consensusStateStore) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
	if css.virtualUTXODiffStaging == nil {
		return nil
	}

	toRemoveIterator := css.virtualUTXODiffStaging.ToRemove().Iterator()
	for toRemoveIterator.Next() {
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
	for toAddIterator.Next() {
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

func (css *consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
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

func newCursorUTXOSetIterator(cursor model.DBCursor) model.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u utxoSetIterator) First() {
	u.cursor.First()
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

func (css *consensusStateStore) OverwriteVirtualUTXOSet(virtualUTXOSetIterator model.ReadOnlyUTXOSetIterator) error {
	if css.virtualUTXODiffStaging != nil {
		return errors.New("cannot overwrite virtual UTXO set while virtual UTXO diff is staged")
	}

	// Clear the cache
	css.virtualUTXOSetCache.Clear()

	// Delete all the old UTXOs from the database
	deleteCursor, err := css.databaseContext.Cursor(utxoSetBucket)
	if err != nil {
		return err
	}
	for deleteCursor.Next() {
		key, err := deleteCursor.Key()
		if err != nil {
			return err
		}
		err = css.databaseContext.Delete(key)
		if err != nil {
			return err
		}
	}

	// Insert all the new UTXOs into the database
	virtualUTXOSetIterator.First()
	for virtualUTXOSetIterator.Next() {
		outpoint, entry, err := virtualUTXOSetIterator.Get()
		if err != nil {
			return err
		}

		key, err := utxoKey(outpoint)
		if err != nil {
			return err
		}
		serializedUTXOEntry, err := serializeUTXOEntry(entry)
		if err != nil {
			return err
		}

		err = css.databaseContext.Put(key, serializedUTXOEntry)
		if err != nil {
			return err
		}
	}

	return nil
}
