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

func (c consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff *model.UTXODiff) error {
	if c.stagedVirtualUTXOSet != nil {
		return errors.New("cannot stage virtual UTXO diff while virtual UTXO set is staged")
	}

	c.stagedVirtualUTXODiff = virtualUTXODiff
	return nil
}

func (c consensusStateStore) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
	if c.stagedVirtualUTXOSet != nil {
		return errors.New("cannot commit virtual UTXO diff while virtual UTXO set is staged")
	}

	for toRemoveOutpoint := range c.stagedVirtualUTXODiff.ToRemove {
		dbKey, err := utxoKey(&toRemoveOutpoint)
		if err != nil {
			return err
		}
		err = dbTx.Delete(dbKey)
		if err != nil {
			return err
		}
	}

	for toAddOutpoint, toAddEntry := range c.stagedVirtualUTXODiff.ToAdd {
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

	return nil
}

func (c consensusStateStore) commitVirtualUTXOSet(dbTx model.DBTransaction) error {
	if c.stagedVirtualUTXODiff != nil {
		return errors.New("cannot commit virtual UTXO set while virtual UTXO diff is staged")
	}

	for outpoint, utxoEntry := range c.stagedVirtualUTXOSet {
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

	return nil
}

func (c consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {

	if c.stagedVirtualUTXOSet != nil {
		return c.utxoByOutpointFromStagedVirtualUTXOSet(outpoint)
	}

	return c.utxoByOutpointFromStagedVirtualUTXODiff(dbContext, outpoint)
}

func (c consensusStateStore) utxoByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {

	if c.stagedVirtualUTXODiff != nil {
		if _, ok := c.stagedVirtualUTXODiff.ToRemove[*outpoint]; ok {
			return nil, errors.Errorf("outpoint was not found")
		}
		if utxoEntry, ok := c.stagedVirtualUTXODiff.ToAdd[*outpoint]; ok {
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

func (c consensusStateStore) utxoByOutpointFromStagedVirtualUTXOSet(outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {
	if utxoEntry, ok := c.stagedVirtualUTXOSet[*outpoint]; ok {
		return utxoEntry, nil
	}

	return nil, errors.Errorf("outpoint was not found")
}

func (c consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	if c.stagedVirtualUTXOSet != nil {
		return c.hasUTXOByOutpointFromStagedVirtualUTXOSet(outpoint), nil
	}

	return c.hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext, outpoint)
}

func (c consensusStateStore) hasUTXOByOutpointFromStagedVirtualUTXODiff(dbContext model.DBReader,
	outpoint *externalapi.DomainOutpoint) (bool, error) {
	if _, ok := c.stagedVirtualUTXODiff.ToRemove[*outpoint]; ok {
		return false, nil
	}
	if _, ok := c.stagedVirtualUTXODiff.ToAdd[*outpoint]; ok {
		return true, nil
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return false, err
	}

	return dbContext.Has(key)
}

func (c consensusStateStore) hasUTXOByOutpointFromStagedVirtualUTXOSet(outpoint *externalapi.DomainOutpoint) bool {
	_, ok := c.stagedVirtualUTXOSet[*outpoint]
	return ok
}

func (c consensusStateStore) VirtualUTXOSetIterator(dbContext model.DBReader) (model.ReadOnlyUTXOSetIterator, error) {
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

func (c consensusStateStore) StageVirtualUTXOSet(virtualUTXOSetIterator model.ReadOnlyUTXOSetIterator) error {
	if c.stagedVirtualUTXODiff != nil {
		return errors.New("cannot stage virtual UTXO set while virtual UTXO diff is staged")
	}

	c.stagedVirtualUTXOSet = make(model.UTXOCollection)
	for virtualUTXOSetIterator.Next() {
		outpoint, entry, err := virtualUTXOSetIterator.Get()
		if err != nil {
			return err
		}

		if _, exists := c.stagedVirtualUTXOSet[*outpoint]; exists {
			return errors.Errorf("outpoint %s is found more than once in the given iterator", outpoint)
		}
		c.stagedVirtualUTXOSet[*outpoint] = entry
	}

	return nil
}
