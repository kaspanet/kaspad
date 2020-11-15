package addressindex

import (
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

var (
	utxoMapBucket = database.MakeBucket([]byte("utxo-map"))
)

func utxoMapKey(addressKey []byte) *database.Key {
	return utxoMapBucket.Key(addressKey)
}

// addToUTXOMap adds the given address-utxoSet pair to
// the database's UTXOMap.
func addToUTXOMap(db database.Database, addressKey []byte, utxoSet []byte) error {
	key := utxoMapKey(addressKey)
	return db.Put(key, utxoSet)
}

// removeFromUTXOMap removes the given address from the
// database's UTXOMap.
func removeFromUTXOMap(db database.Database, addressKey []byte) error {
	key := utxoMapKey(addressKey)
	return db.Delete(key)
}

// getFromUTXOMap return the given address from the
// database's UTXOMap.
func getFromUTXOMap(db database.Database, addressKey []byte) ([]byte, error) {
	key := utxoMapKey(addressKey)
	return db.Get(key)
}
