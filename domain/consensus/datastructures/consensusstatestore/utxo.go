package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxo"
	"github.com/kaspanet/kaspad/domain/consensus/utils/utxolrucache"
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

func (css *consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff model.UTXODiff) error {
	if css.virtualUTXOSetStaging != nil {
		return errors.New("cannot stage virtual UTXO diff while virtual UTXO set is staged")
	}

	css.virtualUTXODiffStaging = virtualUTXODiff
	return nil
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

func (css *consensusStateStore) commitVirtualUTXOSet(dbTx model.DBTransaction) error {
	if css.virtualUTXOSetStaging == nil {
		return nil
	}

	css.virtualUTXOSetCache = utxolrucache.New(css.utxoSetCacheSize)
	iterator := css.virtualUTXOSetStaging.Iterator()
	for iterator.Next() {
		outpoint, utxoEntry, err := iterator.Get()
		if err != nil {
			return err
		}

		css.virtualUTXOSetCache.Add(outpoint, utxoEntry)
		dbKey, err := utxoKey(outpoint)
		if err != nil {
			return err
		}
		serializedEntry, err := serializeUTXOEntry(utxoEntry)
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

	if css.virtualUTXOSetStaging != nil {
		return css.utxoByOutpointFromStagedVirtualUTXOSet(outpoint)
	}

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

func (css *consensusStateStore) utxoByOutpointFromStagedVirtualUTXOSet(outpoint *externalapi.DomainOutpoint) (
	externalapi.UTXOEntry, error) {
	if utxoEntry, ok := css.virtualUTXOSetStaging.Get(outpoint); ok {
		return utxoEntry, nil
	}

	return nil, errors.Errorf("outpoint was not found")
}

func (css *consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	if css.virtualUTXOSetStaging != nil {
		return css.hasUTXOByOutpointFromStagedVirtualUTXOSet(outpoint), nil
	}

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

func (css *consensusStateStore) hasUTXOByOutpointFromStagedVirtualUTXOSet(outpoint *externalapi.DomainOutpoint) bool {
	return css.virtualUTXOSetStaging.Contains(outpoint)
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

func (css *consensusStateStore) StageVirtualUTXOSet(virtualUTXOSetIterator model.ReadOnlyUTXOSetIterator) error {
	if css.virtualUTXODiffStaging != nil {
		return errors.New("cannot stage virtual UTXO set while virtual UTXO diff is staged")
	}

	utxoMap := make(map[externalapi.DomainOutpoint]externalapi.UTXOEntry)
	for virtualUTXOSetIterator.Next() {
		outpoint, entry, err := virtualUTXOSetIterator.Get()
		if err != nil {
			return err
		}

		if _, exists := utxoMap[*outpoint]; exists {
			return errors.Errorf("outpoint %s is found more than once in the given iterator", outpoint)
		}
		utxoMap[*outpoint] = entry
	}
	css.virtualUTXOSetStaging = utxo.NewUTXOCollection(utxoMap)

	return nil
}
