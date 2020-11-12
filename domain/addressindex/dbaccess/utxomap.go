package dbaccess

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var (
	utxoMapBucket = database.MakeBucket([]byte("utxo-map"))
)

func utxoMapKey(addressKey []byte) *database.Key {
	return utxoMapBucket.Key(addressKey)
}

// AddToUTXOMap adds the given address-utxoSet pair to
// the database's UTXOMap.
func AddToUTXOMap(context database.Database, addressKey []byte, utxoSet []byte) error {
	key := utxoMapKey(addressKey)
	return context.Put(key, utxoSet)
}

// RemoveFromUTXOMap removes the given address from the
// database's UTXOMap.
func RemoveFromUTXOMap(context database.Database, addressKey []byte) error {
	key := utxoMapKey(addressKey)
	return context.Delete(key)
}

// GetFromUTXOMap return the given address from the
// database's UTXOMap.
func GetFromUTXOMap(context database.Database, addressKey []byte) ([]byte, error) {
	key := utxoMapKey(addressKey)
	return context.Get(key)
}

// UTXOMapCursor opens a cursor over all the UTXOMap entries
// that have been previously added to the database.
func UTXOMapCursor(context database.Database) (database.Cursor, error) {
	return context.Cursor(utxoMapBucket)
}
