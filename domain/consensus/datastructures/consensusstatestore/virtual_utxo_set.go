package consensusstatestore

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/utils/dbkeys"
	"github.com/pkg/errors"
)

var overwritingVirtualUTXOSetKey = dbkeys.MakeBucket().Key([]byte("overwriting-virtual-utxo-set"))

func (css *consensusStateStore) BeginOverwritingVirtualUTXOSet() error {
	return css.databaseContext.Put(overwritingVirtualUTXOSetKey, []byte{0})
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

func (css *consensusStateStore) FinishOverwritingVirtualUTXOSet() error {
	return css.databaseContext.Delete(overwritingVirtualUTXOSetKey)
}
