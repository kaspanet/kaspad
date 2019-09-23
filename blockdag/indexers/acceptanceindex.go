package indexers

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

const (
	// acceptanceIndexName is the human-readable name for the index.
	acceptanceIndexName = "acceptance index"
)

var (
	// acceptanceIndexKey is the key of the acceptance index and the db bucket used
	// to house it.
	acceptanceIndexKey = []byte("acceptanceidx")
)

// AcceptanceIndex implements a txAcceptanceData by block hash index. That is to say,
// it stores a mapping between a block's hash and the set of transactions that the
// block accepts among its blue blocks.
type AcceptanceIndex struct {
	db database.DB
	dag *blockdag.BlockDAG
}

// Ensure the TxIndex type implements the Indexer interface.
var _ Indexer = (*AcceptanceIndex)(nil)

// NewAcceptanceIndex returns a new instance of an indexer that is used to create a
// mapping between block hashes and their txAcceptanceData.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockdag package. This allows the index to be
// seamlessly maintained along with the DAG.
func NewAcceptanceIndex(_ *dagconfig.Params) *AcceptanceIndex {
	return &AcceptanceIndex{}
}

// DropAcceptanceIndex drops the acceptance index from the provided database if it
// exists.
func DropAcceptanceIndex(db database.DB, interrupt <-chan struct{}) error {
	return dropIndex(db, acceptanceIndexKey, acceptanceIndexName, interrupt)
}

// Key returns the database key to use for the index as a byte slice.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Key() []byte {
	return acceptanceIndexKey
}

// Name returns the human-readable name of the index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Name() string {
	return acceptanceIndexName
}

// Create is invoked when the indexer manager determines the index needs
// to be created for the first time.  It creates the bucket for the
// acceptance index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Create(dbTx database.Tx) error {
	_, err := dbTx.Metadata().CreateBucket(acceptanceIndexKey)
	return err
}

// Init initializes the hash-based acceptance index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Init(db database.DB, dag *blockdag.BlockDAG) error {
	idx.db = db
	idx.dag = dag
	return nil
}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) ConnectBlock(dbTx database.Tx, _ *util.Block, newBlockID uint64, _ *blockdag.BlockDAG,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData, _ blockdag.MultiBlockTxsAcceptanceData) error {
	return dbPutTxsAcceptanceData(dbTx, serializeBlockID(newBlockID), txsAcceptanceData)
}

// TxsAcceptanceData returns the acceptance data of all the transactions that
// were accepted by the block with hash blockHash.
func (idx *AcceptanceIndex) TxsAcceptanceData(blockHash *daghash.Hash) (blockdag.MultiBlockTxsAcceptanceData, error) {
	var txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		txsAcceptanceData, err = dbFetchTxsAcceptanceDataByHash(dbTx, blockHash)
		return err
	})
	if err != nil {
		return nil, err
	}
	return txsAcceptanceData, nil
}

func (idx *AcceptanceIndex) Recover(dbTx database.Tx, currentBlockID, lastKnownBlockID uint64) error {
	for blockID := currentBlockID + 1; blockID <= lastKnownBlockID; blockID++{
		hash, err := dbFetchBlockHashByID(dbTx, currentBlockID)
		if err != nil{
			return err
		}
		txAcceptanceData, err := idx.dag.TxsAcceptedByBlockHash(hash)
		if err != nil{
			return err
		}
		err = idx.ConnectBlock(dbTx, nil, blockID, nil, txAcceptanceData, nil)
		if err != nil{
			return err
		}
	}
	return nil
}

func dbPutTxsAcceptanceData(dbTx database.Tx, newBlockID []byte,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		return err
	}

	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	return bucket.Put(newBlockID, serializedTxsAcceptanceData)
}

func dbFetchTxsAcceptanceDataByHash(dbTx database.Tx,
	hash *daghash.Hash) (blockdag.MultiBlockTxsAcceptanceData, error) {

	blockID, err := dbFetchBlockIDByHash(dbTx, hash)
	if err != nil{
		return nil, err
	}

	return dbFetchTxsAcceptanceDataByID(dbTx, blockID)
}

func dbFetchTxsAcceptanceDataByID(dbTx database.Tx,
	blockID uint64) (blockdag.MultiBlockTxsAcceptanceData, error) {
	serializedBlockID := serializeBlockID(blockID)
	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	serializedTxsAcceptanceData := bucket.Get(serializedBlockID)
	if serializedTxsAcceptanceData == nil {
		return nil, fmt.Errorf("no entry in the accpetance index for block with hash %s", hash)
	}

	return deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
}

type serializableTxAcceptanceData struct {
	MsgTx      wire.MsgTx
	IsAccepted bool
}

type serializableBlockTxsAcceptanceData []serializableTxAcceptanceData

type serializableMultiBlockTxsAcceptanceData map[daghash.Hash]serializableBlockTxsAcceptanceData

func serializeMultiBlockTxsAcceptanceData(
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) ([]byte, error) {
	// Convert MultiBlockTxsAcceptanceData to a serializable format
	serializableData := make(serializableMultiBlockTxsAcceptanceData, len(txsAcceptanceData))
	for hash, blockTxsAcceptanceData := range txsAcceptanceData {
		serializableBlockData := make(serializableBlockTxsAcceptanceData, len(blockTxsAcceptanceData))
		for i, txAcceptanceData := range blockTxsAcceptanceData {
			serializableBlockData[i] = serializableTxAcceptanceData{
				MsgTx:      *txAcceptanceData.Tx.MsgTx(),
				IsAccepted: txAcceptanceData.IsAccepted,
			}
		}
		serializableData[hash] = serializableBlockData
	}

	// Serialize
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(serializableData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func deserializeMultiBlockTxsAcceptanceData(
	serializedTxsAcceptanceData []byte) (blockdag.MultiBlockTxsAcceptanceData, error) {
	// Deserialize
	buffer := bytes.NewBuffer(serializedTxsAcceptanceData)
	decoder := gob.NewDecoder(buffer)
	var serializedData serializableMultiBlockTxsAcceptanceData
	err := decoder.Decode(&serializedData)
	if err != nil {
		return nil, err
	}

	// Convert serializable format to MultiBlockTxsAcceptanceData
	txsAcceptanceData := make(blockdag.MultiBlockTxsAcceptanceData, len(serializedData))
	for hash, serializableBlockData := range serializedData {
		blockTxsAcceptanceData := make(blockdag.BlockTxsAcceptanceData, len(serializableBlockData))
		for i, txData := range serializableBlockData {
			msgTx := txData.MsgTx
			blockTxsAcceptanceData[i] = blockdag.TxAcceptanceData{
				Tx:         util.NewTx(&msgTx),
				IsAccepted: txData.IsAccepted,
			}
		}
		txsAcceptanceData[hash] = blockTxsAcceptanceData
	}

	return txsAcceptanceData, nil
}
