package blockdag

import (
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/pkg/errors"
)

var (
	// idByHashIndexBucketName is the name of the db bucket used to house
	// the block hash -> block id index.
	idByHashIndexBucketName = []byte("idbyhashidx")

	// hashByIDIndexBucketName is the name of the db bucket used to house
	// the block id -> block hash index.
	hashByIDIndexBucketName = []byte("hashbyididx")

	currentBlockIDKey = []byte("currentblockid")
)

// -----------------------------------------------------------------------------
// This is a mapping between block hashes and unique IDs. The ID
// is simply a sequentially incremented uint64 that is used instead of block hash
// for the indexers. This is useful because it is only 8 bytes versus 32 bytes
// hashes and thus saves a ton of space when a block is referenced in an index.
// It consists of three buckets: the first bucket maps the hash of each
// block to the unique ID and the second maps that ID back to the block hash.
// The third bucket contains the last received block ID, and is used
// when starting the node to check that the enabled indexes are up to date
// with the latest received block, and if not, initiate recovery process.
//
// The serialized format for keys and values in the block hash to ID bucket is:
//   <hash> = <ID>
//
//   Field           Type              Size
//   hash            daghash.Hash     32 bytes
//   ID              uint64            8 bytes
//   -----
//   Total: 40 bytes
//
// The serialized format for keys and values in the ID to block hash bucket is:
//   <ID> = <hash>
//
//   Field           Type              Size
//   ID              uint64            8 bytes
//   hash            daghash.Hash     32 bytes
//   -----
//   Total: 40 bytes
//
// -----------------------------------------------------------------------------

const blockIDSize = 8 // 8 bytes for block ID

// DBFetchBlockIDByHash uses an existing database transaction to retrieve the
// block id for the provided hash from the index.
func DBFetchBlockIDByHash(dbTx database.Tx, hash *daghash.Hash) (uint64, error) {
	hashIndex := dbTx.Metadata().Bucket(idByHashIndexBucketName)
	serializedID := hashIndex.Get(hash[:])
	if serializedID == nil {
		return 0, errors.Errorf("no entry in the block ID index for block with hash %s", hash)
	}

	return DeserializeBlockID(serializedID), nil
}

// DBFetchBlockHashBySerializedID uses an existing database transaction to
// retrieve the hash for the provided serialized block id from the index.
func DBFetchBlockHashBySerializedID(dbTx database.Tx, serializedID []byte) (*daghash.Hash, error) {
	idIndex := dbTx.Metadata().Bucket(hashByIDIndexBucketName)
	hashBytes := idIndex.Get(serializedID)
	if hashBytes == nil {
		return nil, errors.Errorf("no entry in the block ID index for block with id %d", byteOrder.Uint64(serializedID))
	}

	var hash daghash.Hash
	copy(hash[:], hashBytes)
	return &hash, nil
}

// dbPutBlockIDIndexEntry uses an existing database transaction to update or add
// the index entries for the hash to id and id to hash mappings for the provided
// values.
func dbPutBlockIDIndexEntry(dbTx database.Tx, hash *daghash.Hash, serializedID []byte) error {
	// Add the block hash to ID mapping to the index.
	meta := dbTx.Metadata()
	hashIndex := meta.Bucket(idByHashIndexBucketName)
	if err := hashIndex.Put(hash[:], serializedID[:]); err != nil {
		return err
	}

	// Add the block ID to hash mapping to the index.
	idIndex := meta.Bucket(hashByIDIndexBucketName)
	return idIndex.Put(serializedID[:], hash[:])
}

// DBFetchCurrentBlockID returns the last known block ID.
func DBFetchCurrentBlockID(dbTx database.Tx) uint64 {
	serializedID := dbTx.Metadata().Get(currentBlockIDKey)
	if serializedID == nil {
		return 0
	}
	return DeserializeBlockID(serializedID)
}

// DeserializeBlockID returns a deserialized block id
func DeserializeBlockID(serializedID []byte) uint64 {
	return byteOrder.Uint64(serializedID)
}

// SerializeBlockID returns a serialized block id
func SerializeBlockID(blockID uint64) []byte {
	serializedBlockID := make([]byte, blockIDSize)
	byteOrder.PutUint64(serializedBlockID, blockID)
	return serializedBlockID
}

// DBFetchBlockHashByID uses an existing database transaction to retrieve the
// hash for the provided block id from the index.
func DBFetchBlockHashByID(dbTx database.Tx, id uint64) (*daghash.Hash, error) {
	return DBFetchBlockHashBySerializedID(dbTx, SerializeBlockID(id))
}

func createBlockID(dbTx database.Tx, blockHash *daghash.Hash) (uint64, error) {
	currentBlockID := DBFetchCurrentBlockID(dbTx)
	newBlockID := currentBlockID + 1
	serializedNewBlockID := SerializeBlockID(newBlockID)
	err := dbTx.Metadata().Put(currentBlockIDKey, serializedNewBlockID)
	if err != nil {
		return 0, err
	}
	err = dbPutBlockIDIndexEntry(dbTx, blockHash, serializedNewBlockID)
	if err != nil {
		return 0, err
	}
	return newBlockID, nil
}
