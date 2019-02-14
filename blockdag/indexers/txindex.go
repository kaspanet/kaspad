// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"errors"
	"fmt"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig/daghash"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/wire"
)

const (
	// txIndexName is the human-readable name for the index.
	txIndexName = "transaction index"
)

var (
	includingBlocksIndexKey = []byte("includingblocksidx")

	includingBlocksIndexKeyEntrySize = 8 // 4 bytes for offset + 4 bytes for transaction length

	acceptingBlocksIndexKey = []byte("acceptingblocksidx")

	acceptingBlocksIndexKeyEntrySize = 4 // 4 bytes for including block ID

	// idByHashIndexBucketName is the name of the db bucket used to house
	// the block id -> block hash index.
	idByHashIndexBucketName = []byte("idbyhashidx")

	// hashByIDIndexBucketName is the name of the db bucket used to house
	// the block hash -> block id index.
	hashByIDIndexBucketName = []byte("hashbyididx")

	// errNoBlockIDEntry is an error that indicates a requested entry does
	// not exist in the block ID index.
	errNoBlockIDEntry = errors.New("no entry in the block ID index")
)

// -----------------------------------------------------------------------------
// The transaction index consists of an entry for every transaction in the DAG.
// In order to significantly optimize the space requirements a separate
// index which provides an internal mapping between each block that has been
// indexed and a unique ID for use within the hash to location mappings.  The ID
// is simply a sequentially incremented uint32.  This is useful because it is
// only 4 bytes versus 32 bytes hashes and thus saves a ton of space in the
// index.
//
// There are four buckets used in total. The first bucket maps the hash of
// each transaction to its location in each block it's included in. The second bucket
// contains all of the blocks that from their viewpoint the transaction has been
// accepted (i.e. the transaction is found in their blue set without double spends),
// and their blue block (or themselves) that included the transaction. The third
// bucket maps the hash of each block to the unique ID and the fourth maps
// that ID back to the block hash.
//
// NOTE: Although it is technically possible for multiple transactions to have
// the same hash as long as the previous transaction with the same hash is fully
// spent, this code only stores the most recent one because doing otherwise
// would add a non-trivial amount of space and overhead for something that will
// realistically never happen per the probability and even if it did, the old
// one must be fully spent and so the most likely transaction a caller would
// want for a given hash is the most recent one anyways.
//
// The including blocks index contains a sub bucket for each transaction hash (32 byte each), that its serialized format is:
//
//   <block id> = <start offset><tx length>
//
//   Field           Type              Size
//   block id        uint32          4 bytes
//   start offset    uint32          4 bytes
//   tx length       uint32          4 bytes
//   -----
//   Total: 12 bytes
//
// The accepting blocks index contains a sub bucket for each transaction hash (32 byte each), that its serialized format is:
//
//   <accepting block id> = <including block id>
//
//   Field           Type              Size
//   accepting block id        uint32          4 bytes
//   including block id        uint32          4 bytes
//   -----
//   Total: 8 bytes
//
// The serialized format for keys and values in the block hash to ID bucket is:
//   <hash> = <ID>
//
//   Field           Type              Size
//   hash            daghash.Hash    32 bytes
//   ID              uint32            4 bytes
//   -----
//   Total: 36 bytes
//
// The serialized format for keys and values in the ID to block hash bucket is:
//   <ID> = <hash>
//
//   Field           Type              Size
//   ID              uint32            4 bytes
//   hash            daghash.Hash    32 bytes
//   -----
//   Total: 36 bytes
//
// -----------------------------------------------------------------------------

// dbPutBlockIDIndexEntry uses an existing database transaction to update or add
// the index entries for the hash to id and id to hash mappings for the provided
// values.
func dbPutBlockIDIndexEntry(dbTx database.Tx, hash *daghash.Hash, id uint32) error {
	// Serialize the height for use in the index entries.
	var serializedID [4]byte
	byteOrder.PutUint32(serializedID[:], id)

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

// dbFetchBlockIDByHash uses an existing database transaction to retrieve the
// block id for the provided hash from the index.
func dbFetchBlockIDByHash(dbTx database.Tx, hash *daghash.Hash) (uint32, error) {
	hashIndex := dbTx.Metadata().Bucket(idByHashIndexBucketName)
	serializedID := hashIndex.Get(hash[:])
	if serializedID == nil {
		return 0, errNoBlockIDEntry
	}

	return byteOrder.Uint32(serializedID), nil
}

// dbFetchBlockHashBySerializedID uses an existing database transaction to
// retrieve the hash for the provided serialized block id from the index.
func dbFetchBlockHashBySerializedID(dbTx database.Tx, serializedID []byte) (*daghash.Hash, error) {
	idIndex := dbTx.Metadata().Bucket(hashByIDIndexBucketName)
	hashBytes := idIndex.Get(serializedID)
	if hashBytes == nil {
		return nil, errNoBlockIDEntry
	}

	var hash daghash.Hash
	copy(hash[:], hashBytes)
	return &hash, nil
}

// dbFetchBlockHashByID uses an existing database transaction to retrieve the
// hash for the provided block id from the index.
func dbFetchBlockHashByID(dbTx database.Tx, id uint32) (*daghash.Hash, error) {
	var serializedID [4]byte
	byteOrder.PutUint32(serializedID[:], id)
	return dbFetchBlockHashBySerializedID(dbTx, serializedID[:])
}

func putIncludingBlocksEntry(target []byte, txLoc wire.TxLoc) {
	byteOrder.PutUint32(target, uint32(txLoc.TxStart))
	byteOrder.PutUint32(target[4:], uint32(txLoc.TxLen))
}

func putAcceptingBlocksEntry(target []byte, includingBlockID uint32) {
	byteOrder.PutUint32(target, includingBlockID)
}

func dbPutIncludingBlocksEntry(dbTx database.Tx, txID *daghash.TxID, blockID uint32, serializedData []byte) error {
	bucket, err := dbTx.Metadata().Bucket(includingBlocksIndexKey).CreateBucketIfNotExists(txID[:])
	if err != nil {
		return err
	}
	blockIDBytes := make([]byte, 4)
	byteOrder.PutUint32(blockIDBytes, uint32(blockID))
	return bucket.Put(blockIDBytes, serializedData)
}

func dbPutAcceptingBlocksEntry(dbTx database.Tx, txID *daghash.TxID, blockID uint32, serializedData []byte) error {
	bucket, err := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).CreateBucketIfNotExists(txID[:])
	if err != nil {
		return err
	}
	blockIDBytes := make([]byte, 4)
	byteOrder.PutUint32(blockIDBytes, uint32(blockID))
	return bucket.Put(blockIDBytes, serializedData)
}

// dbFetchFirstTxRegion uses an existing database transaction to fetch the block
// region for the provided transaction hash from the transaction index.  When
// there is no entry for the provided hash, nil will be returned for the both
// the region and the error.
//
// P.S Because the transaction can be found in multiple blocks, this function arbitarily
// returns the first block region that is stored in the txindex.
func dbFetchFirstTxRegion(dbTx database.Tx, txID *daghash.TxID) (*database.BlockRegion, error) {
	// Load the record from the database and return now if it doesn't exist.
	txBucket := dbTx.Metadata().Bucket(includingBlocksIndexKey).Bucket(txID[:])
	if txBucket == nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No block region"+
				"was found for %s", txID),
		}
	}
	cursor := txBucket.Cursor()
	if ok := cursor.First(); !ok {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No block region"+
				"was found for %s", txID),
		}
	}
	blockIDBytes := cursor.Key()
	serializedData := cursor.Value()
	if len(serializedData) == 0 {
		return nil, nil
	}

	// Ensure the serialized data has enough bytes to properly deserialize.
	if len(serializedData) < includingBlocksIndexKeyEntrySize {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt transaction index "+
				"entry for %s", txID),
		}
	}

	// Load the block hash associated with the block ID.
	hash, err := dbFetchBlockHashBySerializedID(dbTx, blockIDBytes)
	if err != nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt transaction index "+
				"entry for %s: %v", txID, err),
		}
	}

	// Deserialize the final entry.
	region := database.BlockRegion{Hash: &daghash.Hash{}}
	copy(region.Hash[:], hash[:])
	region.Offset = byteOrder.Uint32(serializedData[:4])
	region.Len = byteOrder.Uint32(serializedData[4:])

	return &region, nil
}

// dbAddTxIndexEntries uses an existing database transaction to add a
// transaction index entry for every transaction in the passed block.
func dbAddTxIndexEntries(dbTx database.Tx, block *util.Block, blockID uint32, acceptedTxData []*blockdag.TxWithBlockHash) error {
	// The offset and length of the transactions within the serialized
	// block.
	txLocs, err := block.TxLoc()
	if err != nil {
		return err
	}

	// As an optimization, allocate a single slice big enough to hold all
	// of the serialized transaction index entries for the block and
	// serialize them directly into the slice.  Then, pass the appropriate
	// subslice to the database to be written.  This approach significantly
	// cuts down on the number of required allocations.
	includingBlocksOffset := 0
	serializedIncludingBlocksValues := make([]byte, len(block.Transactions())*includingBlocksIndexKeyEntrySize)
	for i, tx := range block.Transactions() {
		putIncludingBlocksEntry(serializedIncludingBlocksValues[includingBlocksOffset:], txLocs[i])
		endOffset := includingBlocksOffset + includingBlocksIndexKeyEntrySize
		err := dbPutIncludingBlocksEntry(dbTx, tx.ID(), blockID,
			serializedIncludingBlocksValues[includingBlocksOffset:endOffset:endOffset])
		if err != nil {
			return err
		}
		includingBlocksOffset += includingBlocksIndexKeyEntrySize
	}

	blockHashToID := make(map[daghash.Hash]uint32)
	blockHashToID[*block.Hash()] = blockID

	acceptingBlocksOffset := 0

	serializedAcceptingBlocksValues := make([]byte, len(acceptedTxData)*acceptingBlocksIndexKeyEntrySize)
	for _, tx := range acceptedTxData {
		var includingBlockID uint32
		var err error
		var ok bool

		if includingBlockID, ok = blockHashToID[*tx.InBlock]; !ok {
			includingBlockID, err = dbFetchBlockIDByHash(dbTx, tx.InBlock)
			if err != nil {
				return err
			}
			blockHashToID[*tx.InBlock] = includingBlockID
		}

		putAcceptingBlocksEntry(serializedAcceptingBlocksValues[acceptingBlocksOffset:], includingBlockID)
		endOffset := acceptingBlocksOffset + acceptingBlocksIndexKeyEntrySize
		err = dbPutAcceptingBlocksEntry(dbTx, tx.Tx.ID(), blockID,
			serializedAcceptingBlocksValues[acceptingBlocksOffset:endOffset:endOffset])
		if err != nil {
			return err
		}
		acceptingBlocksOffset += acceptingBlocksIndexKeyEntrySize
	}

	return nil
}

// TxIndex implements a transaction by hash index.  That is to say, it supports
// querying all transactions by their hash.
type TxIndex struct {
	db         database.DB
	curBlockID uint32
}

// Ensure the TxIndex type implements the Indexer interface.
var _ Indexer = (*TxIndex)(nil)

// Init initializes the hash-based transaction index.  In particular, it finds
// the highest used block ID and stores it for later use when connecting or
// disconnecting blocks.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Init(db database.DB) error {
	idx.db = db

	// Find the latest known block id field for the internal block id
	// index and initialize it.  This is done because it's a lot more
	// efficient to do a single search at initialize time than it is to
	// write another value to the database on every update.
	err := idx.db.View(func(dbTx database.Tx) error {
		// Scan forward in large gaps to find a block id that doesn't
		// exist yet to serve as an upper bound for the binary search
		// below.
		var highestKnown, nextUnknown uint32
		testBlockID := uint32(1)
		increment := uint32(100000)
		for {
			_, err := dbFetchBlockHashByID(dbTx, testBlockID)
			if err != nil {
				nextUnknown = testBlockID
				break
			}

			highestKnown = testBlockID
			testBlockID += increment
		}
		log.Tracef("Forward scan (highest known %d, next unknown %d)",
			highestKnown, nextUnknown)

		// No used block IDs due to new database.
		if nextUnknown == 1 {
			return nil
		}

		// Use a binary search to find the final highest used block id.
		// This will take at most ceil(log_2(increment)) attempts.
		for {
			testBlockID = (highestKnown + nextUnknown) / 2
			_, err := dbFetchBlockHashByID(dbTx, testBlockID)
			if err != nil {
				nextUnknown = testBlockID
			} else {
				highestKnown = testBlockID
			}
			log.Tracef("Binary scan (highest known %d, next "+
				"unknown %d)", highestKnown, nextUnknown)
			if highestKnown+1 == nextUnknown {
				break
			}
		}

		idx.curBlockID = highestKnown
		return nil
	})
	if err != nil {
		return err
	}

	log.Debugf("Current internal block ID: %d", idx.curBlockID)
	return nil
}

// Key returns the database key to use for the index as a byte slice.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Key() []byte {
	return includingBlocksIndexKey
}

// Name returns the human-readable name of the index.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Name() string {
	return txIndexName
}

// Create is invoked when the indexer manager determines the index needs
// to be created for the first time.  It creates the buckets for the hash-based
// transaction index and the internal block ID indexes.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Create(dbTx database.Tx) error {
	meta := dbTx.Metadata()
	if _, err := meta.CreateBucket(idByHashIndexBucketName); err != nil {
		return err
	}
	if _, err := meta.CreateBucket(hashByIDIndexBucketName); err != nil {
		return err
	}
	if _, err := meta.CreateBucket(includingBlocksIndexKey); err != nil {
		return err
	}
	_, err := meta.CreateBucket(acceptingBlocksIndexKey)
	return err

}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG.  This indexer adds a hash-to-transaction mapping
// for every transaction in the passed block.
//
// This is part of the Indexer interface.
func (idx *TxIndex) ConnectBlock(dbTx database.Tx, block *util.Block, _ *blockdag.BlockDAG, acceptedTxsData []*blockdag.TxWithBlockHash) error {
	// Increment the internal block ID to use for the block being connected
	// and add all of the transactions in the block to the index.
	newBlockID := idx.curBlockID + 1
	if block.MsgBlock().Header.IsGenesis() {
		newBlockID = 0
	}
	if err := dbAddTxIndexEntries(dbTx, block, newBlockID, acceptedTxsData); err != nil {
		return err
	}

	// Add the new block ID index entry for the block being connected and
	// update the current internal block ID accordingly.
	err := dbPutBlockIDIndexEntry(dbTx, block.Hash(), newBlockID)
	if err != nil {
		return err
	}
	idx.curBlockID = newBlockID
	return nil
}

// TxFirstBlockRegion returns the first block region for the provided transaction hash
// from the transaction index.  The block region can in turn be used to load the
// raw transaction bytes.  When there is no entry for the provided hash, nil
// will be returned for the both the entry and the error.
//
// This function is safe for concurrent access.
func (idx *TxIndex) TxFirstBlockRegion(txID *daghash.TxID) (*database.BlockRegion, error) {
	var region *database.BlockRegion
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		region, err = dbFetchFirstTxRegion(dbTx, txID)
		return err
	})
	return region, err
}

// TxBlocks returns the hashes of the blocks where the transaction exists
func (idx *TxIndex) TxBlocks(txHash *daghash.Hash) ([]daghash.Hash, error) {
	blockHashes := make([]daghash.Hash, 0)
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		blockHashes, err = dbFetchTxBlocks(dbTx, txHash)
		if err != nil {
			return err
		}
		return nil
	})
	return blockHashes, err
}

func dbFetchTxBlocks(dbTx database.Tx, txHash *daghash.Hash) ([]daghash.Hash, error) {
	blockHashes := make([]daghash.Hash, 0)
	bucket := dbTx.Metadata().Bucket(includingBlocksIndexKey).Bucket(txHash[:])
	if bucket == nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No including blocks "+
				"were found for %s", txHash),
		}
	}
	err := bucket.ForEach(func(blockIDBytes, _ []byte) error {
		blockID := byteOrder.Uint32(blockIDBytes)
		blockHash, err := dbFetchBlockHashByID(dbTx, blockID)
		if err != nil {
			return err
		}
		blockHashes = append(blockHashes, *blockHash)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return blockHashes, nil
}

// BlockThatAcceptedTx returns the hash of the block where the transaction got accepted (from the virtual block point of view)
func (idx *TxIndex) BlockThatAcceptedTx(dag *blockdag.BlockDAG, txID *daghash.TxID) (*daghash.Hash, error) {
	var acceptingBlock *daghash.Hash
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		acceptingBlock, err = dbFetchTxAcceptingBlock(dbTx, txID, dag)
		return err
	})
	return acceptingBlock, err
}

func dbFetchTxAcceptingBlock(dbTx database.Tx, txID *daghash.TxID, dag *blockdag.BlockDAG) (*daghash.Hash, error) {
	bucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(txID[:])
	if bucket == nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No accepting blocks "+
				"were found for %s", txID),
		}
	}
	cursor := bucket.Cursor()
	if !cursor.First() {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No accepting blocks "+
				"were found for %s", txID),
		}
	}
	for ; cursor.Key() != nil; cursor.Next() {
		blockID := byteOrder.Uint32(cursor.Key())
		blockHash, err := dbFetchBlockHashByID(dbTx, blockID)
		if err != nil {
			return nil, err
		}
		if dag.IsInSelectedPathChain(blockHash) {
			return blockHash, nil
		}
	}
	return nil, nil
}

// NewTxIndex returns a new instance of an indexer that is used to create a
// mapping of the hashes of all transactions in the blockchain to the respective
// block, location within the block, and size of the transaction.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockchain package.  This allows the index to be
// seamlessly maintained along with the chain.
func NewTxIndex() *TxIndex {
	return &TxIndex{}
}

// dropBlockIDIndex drops the internal block id index.
func dropBlockIDIndex(db database.DB) error {
	return db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()
		err := meta.DeleteBucket(idByHashIndexBucketName)
		if err != nil {
			return err
		}

		return meta.DeleteBucket(hashByIDIndexBucketName)
	})
}

// DropTxIndex drops the transaction index from the provided database if it
// exists.  Since the address index relies on it, the address index will also be
// dropped when it exists.
func DropTxIndex(db database.DB, interrupt <-chan struct{}) error {
	err := dropIndex(db, addrIndexKey, addrIndexName, interrupt)
	if err != nil {
		return err
	}

	err = dropIndex(db, includingBlocksIndexKey, addrIndexName, interrupt)
	if err != nil {
		return err
	}

	return dropIndex(db, acceptingBlocksIndexKey, txIndexName, interrupt)
}
