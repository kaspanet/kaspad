package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
)

var (
	utxoBucket = database.MakeBucket([]byte("utxo"))
)

func utxoKey(outpointKey []byte) []byte {
	return utxoBucket.Key(outpointKey)
}

// AddToUTXOSet adds the given outpoint-utxoEntry pair to
// the database's UTXO set.
func AddToUTXOSet(context Context, outpointKey []byte, utxoEntry []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := utxoKey(outpointKey)
	return accessor.Put(key, utxoEntry)
}

// RemoveFromUTXOSet removes the given outpoint from the
// database's UTXO set.
func RemoveFromUTXOSet(context Context, outpointKey []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}

	key := utxoKey(outpointKey)
	return accessor.Delete(key)
}

// UTXOSetCursor opens a cursor over all the UTXO entries
// that have been previously added to the database.
func UTXOSetCursor(context Context) (database.Cursor, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}

	return accessor.Cursor(utxoBucket.Path())
}
