// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"fmt"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
)

const (
	// txIndexName is the human-readable name for the index.
	txIndexName = "transaction index"

	includingBlocksIndexKeyEntrySize = 8 // 4 bytes for offset + 4 bytes for transaction length
)

var (
	includingBlocksIndexKey = []byte("includingblocksidx")

	acceptingBlocksIndexKey = []byte("acceptingblocksidx")
)

// txsAcceptedByVirtual is the in-memory index of txIDs that were accepted
// by the current virtual
var txsAcceptedByVirtual map[daghash.TxID]bool

// -----------------------------------------------------------------------------
// The transaction index consists of an entry for every transaction in the DAG.
//
// There are two buckets used in total. The first bucket maps the hash of
// each transaction to its location in each block it's included in. The second bucket
// contains all of the blocks that from their viewpoint the transaction has been
// accepted (i.e. the transaction is found in their blue set without double spends),
// and their blue block (or themselves) that included the transaction.
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
//   block id        uint64          8 bytes
//   start offset    uint32          4 bytes
//   tx length       uint32          4 bytes
//   -----
//   Total: 16 bytes
//
// The accepting blocks index contains a sub bucket for each transaction hash (32 byte each), that its serialized format is:
//
//   <accepting block id> = <including block id>
//
//   Field           Type              Size
//   accepting block id        uint64          8 bytes
//   including block id        uint64          8 bytes
//   -----
//   Total: 16 bytes
//
// -----------------------------------------------------------------------------

func putIncludingBlocksEntry(target []byte, txLoc wire.TxLoc) {
	byteOrder.PutUint32(target, uint32(txLoc.TxStart))
	byteOrder.PutUint32(target[4:], uint32(txLoc.TxLen))
}

func dbPutIncludingBlocksEntry(dbTx database.Tx, txID *daghash.TxID, blockID uint64, serializedData []byte) error {
	bucket, err := dbTx.Metadata().Bucket(includingBlocksIndexKey).CreateBucketIfNotExists(txID[:])
	if err != nil {
		return err
	}
	return bucket.Put(blockdag.SerializeBlockID(blockID), serializedData)
}

func dbPutAcceptingBlocksEntry(dbTx database.Tx, txID *daghash.TxID, blockID uint64, serializedData []byte) error {
	bucket, err := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).CreateBucketIfNotExists(txID[:])
	if err != nil {
		return err
	}
	return bucket.Put(blockdag.SerializeBlockID(blockID), serializedData)
}

// dbFetchFirstTxRegion uses an existing database transaction to fetch the block
// region for the provided transaction hash from the transaction index. When
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
			Description: fmt.Sprintf("No block region "+
				"was found for %s", txID),
		}
	}
	cursor := txBucket.Cursor()
	if ok := cursor.First(); !ok {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No block region "+
				"was found for %s", txID),
		}
	}
	serializedBlockID := cursor.Key()
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
	hash, err := blockdag.DBFetchBlockHashBySerializedID(dbTx, serializedBlockID)
	if err != nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("corrupt transaction index "+
				"entry for %s: %s", txID, err),
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
func dbAddTxIndexEntries(dbTx database.Tx, block *util.Block, blockID uint64, multiBlockTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	// The offset and length of the transactions within the serialized
	// block.
	txLocs, err := block.TxLoc()
	if err != nil {
		return err
	}

	// As an optimization, allocate a single slice big enough to hold all
	// of the serialized transaction index entries for the block and
	// serialize them directly into the slice. Then, pass the appropriate
	// subslice to the database to be written. This approach significantly
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

	for _, blockTxsAcceptanceData := range multiBlockTxsAcceptanceData {
		var includingBlockID uint64
		if blockTxsAcceptanceData.BlockHash.IsEqual(block.Hash()) {
			includingBlockID = blockID
		} else {
			includingBlockID, err = blockdag.DBFetchBlockIDByHash(dbTx, &blockTxsAcceptanceData.BlockHash)
			if err != nil {
				return err
			}
		}

		serializedIncludingBlockID := blockdag.SerializeBlockID(includingBlockID)

		for _, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
			err = dbPutAcceptingBlocksEntry(dbTx, txAcceptanceData.Tx.ID(), blockID, serializedIncludingBlockID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func updateTxsAcceptedByVirtual(virtualTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	// Initialize a new txsAcceptedByVirtual
	entries := 0
	for _, blockTxsAcceptanceData := range virtualTxsAcceptanceData {
		entries += len(blockTxsAcceptanceData.TxAcceptanceData)
	}
	txsAcceptedByVirtual = make(map[daghash.TxID]bool, entries)

	// Copy virtualTxsAcceptanceData to txsAcceptedByVirtual
	for _, blockTxsAcceptanceData := range virtualTxsAcceptanceData {
		for _, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
			txsAcceptedByVirtual[*txAcceptanceData.Tx.ID()] = true
		}
	}

	return nil
}

// TxIndex implements a transaction by hash index. That is to say, it supports
// querying all transactions by their hash.
type TxIndex struct {
	db database.DB
}

// Ensure the TxIndex type implements the Indexer interface.
var _ Indexer = (*TxIndex)(nil)

// Init initializes the hash-based transaction index. In particular, it finds
// the highest used block ID and stores it for later use when connecting or
// disconnecting blocks.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Init(db database.DB, dag *blockdag.BlockDAG) error {
	idx.db = db

	// Initialize the txsAcceptedByVirtual index
	virtualTxsAcceptanceData, err := dag.TxsAcceptedByVirtual()
	if err != nil {
		return err
	}
	err = updateTxsAcceptedByVirtual(virtualTxsAcceptanceData)
	if err != nil {
		return err
	}
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
// to be created for the first time. It creates the buckets for the hash-based
// transaction index and the internal block ID indexes.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Create(dbTx database.Tx) error {
	meta := dbTx.Metadata()
	if _, err := meta.CreateBucket(includingBlocksIndexKey); err != nil {
		return err
	}
	_, err := meta.CreateBucket(acceptingBlocksIndexKey)
	return err

}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG. This indexer adds a hash-to-transaction mapping
// for every transaction in the passed block.
//
// This is part of the Indexer interface.
func (idx *TxIndex) ConnectBlock(dbTx database.Tx, block *util.Block, blockID uint64, dag *blockdag.BlockDAG,
	acceptedTxsData blockdag.MultiBlockTxsAcceptanceData, virtualTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	if err := dbAddTxIndexEntries(dbTx, block, blockID, acceptedTxsData); err != nil {
		return err
	}

	err := updateTxsAcceptedByVirtual(virtualTxsAcceptanceData)
	if err != nil {
		return err
	}
	return nil
}

// TxFirstBlockRegion returns the first block region for the provided transaction hash
// from the transaction index. The block region can in turn be used to load the
// raw transaction bytes. When there is no entry for the provided hash, nil
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
func (idx *TxIndex) TxBlocks(txHash *daghash.Hash) ([]*daghash.Hash, error) {
	blockHashes := make([]*daghash.Hash, 0)
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

func dbFetchTxBlocks(dbTx database.Tx, txHash *daghash.Hash) ([]*daghash.Hash, error) {
	blockHashes := make([]*daghash.Hash, 0)
	bucket := dbTx.Metadata().Bucket(includingBlocksIndexKey).Bucket(txHash[:])
	if bucket == nil {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("No including blocks "+
				"were found for %s", txHash),
		}
	}
	err := bucket.ForEach(func(serializedBlockID, _ []byte) error {
		blockHash, err := blockdag.DBFetchBlockHashBySerializedID(dbTx, serializedBlockID)
		if err != nil {
			return err
		}
		blockHashes = append(blockHashes, blockHash)
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
	// If the transaction was accepted by the current virtual,
	// return the zeroHash immediately
	if _, ok := txsAcceptedByVirtual[*txID]; ok {
		return &daghash.ZeroHash, nil
	}

	bucket := dbTx.Metadata().Bucket(acceptingBlocksIndexKey).Bucket(txID[:])
	if bucket == nil {
		return nil, nil
	}
	cursor := bucket.Cursor()
	if !cursor.First() {
		return nil, database.Error{
			ErrorCode: database.ErrCorruption,
			Description: fmt.Sprintf("Accepting blocks bucket is "+
				"empty for %s", txID),
		}
	}
	for ; cursor.Key() != nil; cursor.Next() {
		blockHash, err := blockdag.DBFetchBlockHashBySerializedID(dbTx, cursor.Key())
		if err != nil {
			return nil, err
		}
		if dag.IsInSelectedParentChain(blockHash) {
			return blockHash, nil
		}
	}
	return nil, nil
}

// NewTxIndex returns a new instance of an indexer that is used to create a
// mapping of the hashes of all transactions in the blockDAG to the respective
// block, location within the block, and size of the transaction.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockdag package. This allows the index to be
// seamlessly maintained along with the DAG.
func NewTxIndex() *TxIndex {
	return &TxIndex{}
}

// DropTxIndex drops the transaction index from the provided database if it
// exists. Since the address index relies on it, the address index will also be
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

// Recover is invoked when the indexer wasn't turned on for several blocks
// and the indexer needs to close the gaps.
//
// This is part of the Indexer interface.
func (idx *TxIndex) Recover(dbTx database.Tx, currentBlockID, lastKnownBlockID uint64) error {
	return errors.Errorf("txindex was turned off for %d blocks and can't be recovered."+
		" To resume working drop the txindex with --droptxindex", lastKnownBlockID-currentBlockID)
}
