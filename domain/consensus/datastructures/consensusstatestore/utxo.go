package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var utxoSetBucket = dbkeys.MakeBucket([]byte("virtual-utxo-set"))

func utxoKey(outpoint *externalapi.DomainOutpoint) (model.DBKey, error) {
	serializedOutpoint, err := hashserialization.SerializeOutpoint(outpoint)
	if err != nil {
		return nil, err
	}

	return utxoSetBucket.Key(serializedOutpoint), nil
}

func (c consensusStateStore) StageVirtualUTXODiff(virtualUTXODiff *model.UTXODiff) {
	c.stagedVirtualUTXODiff = virtualUTXODiff
}

func (c consensusStateStore) commitVirtualUTXODiff(dbTx model.DBTransaction) error {
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
		serializedEntry, err := hashserialization.SerializeUTXOEntry(toAddEntry)
		if err != nil {
			return err
		}
		err = dbTx.Put(dbKey, serializedEntry)
		if err != nil {
			return err
		}
	}

	c.stagedVirtualUTXODiff = nil

	return nil
}

func (c consensusStateStore) UTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (
	*externalapi.UTXOEntry, error) {

	if _, ok := c.stagedVirtualUTXODiff.ToRemove[*outpoint]; ok {
		return nil, database.ErrNotFound
	}
	if utxoEntry, ok := c.stagedVirtualUTXODiff.ToAdd[*outpoint]; ok {
		return utxoEntry, nil
	}

	key, err := utxoKey(outpoint)
	if err != nil {
		return nil, err
	}

	serializedUTXOEntry, err := dbContext.Get(key)
	if err != nil {
		return nil, err
	}

	return hashserialization.DeserializeUTXOEntry(serializedUTXOEntry)
}

func (c consensusStateStore) HasUTXOByOutpoint(dbContext model.DBReader, outpoint *externalapi.DomainOutpoint) (bool, error) {
	if _, ok := c.stagedVirtualUTXODiff.ToRemove[*outpoint]; ok {
		return false, database.ErrNotFound
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
