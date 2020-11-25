package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
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

func (css *consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff *model.UTXODiff) error {
	if css.virtualUTXOSetStaging != nil {
		return errors.New("cannot stage virtual UTXO diff while virtual UTXO set is staged")
	}

	css.virtualUTXODiffStaging = virtualUTXODiff.Clone()
	return nil
}

func (css *consensusStateStore) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
	if css.virtualUTXODiffStaging == nil {
		return nil
	}

	for toRemoveOutpoint := range css.virtualUTXODiffStaging.ToRemove {
		dbKey, err := utxoKey(&toRemoveOutpoint)
		if err != nil {
			return err
		}
		err = dbTx.Delete(dbKey)
		if err != nil {
			return err
		}
	}

	for toAddOutpoint, toAddEntry := range css.virtualUTXODiffStaging.ToAdd {
		dbKey, err := utxoKey(&toAddOutpoint)
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

	for outpoint, utxoEntry := range css.virtualUTXOSetStaging {
		dbKey, err := utxoKey(&outpoint)
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
	*externalapi.UTXOEntry, error) {

	if css.virtualUTXOSetStaging != nil {
		return css.utxoByOutpointFromStagedVirtualUTXOSet(outpoint)
	}

	return css.utxoByOutpointFromStagedVirtualUTXODiff(dbContext, outpoint)
}

func (css *consensusStateStore) utxoByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {

	if css.virtualUTXODiffStaging != nil {
		if _, ok := css.virtualUTXODiffStaging.ToRemove[*outpoint]; ok {
			return nil, errors.Errorf("outpoint was not found")
		}
		if utxoEntry, ok := css.virtualUTXODiffStaging.ToAdd[*outpoint]; ok {
			return utxoEntry, nil
		}
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return nil, err
	}

	serializedUTXOEntry, err := dbContext.Get(key)
	if err != nil {
		return nil, err
	}

	return deserializeUTXOEntry(serializedUTXOEntry)
}

func (css *consensusStateStore) utxoByOutpointFromStagedVirtualUTXOSet(outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {
	if utxoEntry, ok := css.virtualUTXOSetStaging[*outpoint]; ok {
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
		if _, ok := css.virtualUTXODiffStaging.ToRemove[*outpoint]; ok {
			return false, nil
		}
		if _, ok := css.virtualUTXODiffStaging.ToAdd[*outpoint]; ok {
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
	_, ok := css.virtualUTXOSetStaging[*outpoint]
	return ok
}

func (css *consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
	cursor, err := dbContext.Cursor(utxoSetBucket)
	if err != nil {
		return nil, err
	}

	return newUTXOSetIterator(cursor), nil
}

type utxoSetIterator struct {
	cursor model.DBCursor
}

func newUTXOSetIterator(cursor model.DBCursor) model.ReadOnlyUTXOSetIterator {
	return &utxoSetIterator{cursor: cursor}
}

func (u utxoSetIterator) Next() bool {
	return u.cursor.Next()
}

func (u utxoSetIterator) Get() (outpoint *externalapi.DomainOutpoint, utxoEntry *externalapi.UTXOEntry, err error) {
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

	css.virtualUTXOSetStaging = make(model.UTXOCollection)
	for virtualUTXOSetIterator.Next() {
		outpoint, entry, err := virtualUTXOSetIterator.Get()
		if err != nil {
			return err
		}

		if _, exists := css.virtualUTXOSetStaging[*outpoint]; exists {
			return errors.Errorf("outpoint %s is found more than once in the given iterator", outpoint)
		}
		css.virtualUTXOSetStaging[*outpoint] = entry
	}

	return nil
}
