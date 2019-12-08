package indexers

import (
	"bytes"
	"encoding/gob"
	"github.com/kaspanet/kaspad/blockdag"
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
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
	db  database.DB
	dag *blockdag.BlockDAG
}

// Ensure the AcceptanceIndex type implements the Indexer interface.
var _ Indexer = (*AcceptanceIndex)(nil)

// NewAcceptanceIndex returns a new instance of an indexer that is used to create a
// mapping between block hashes and their txAcceptanceData.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockdag package. This allows the index to be
// seamlessly maintained along with the DAG.
func NewAcceptanceIndex() *AcceptanceIndex {
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
func (idx *AcceptanceIndex) ConnectBlock(dbTx database.Tx, _ *util.Block, blockID uint64, _ *blockdag.BlockDAG,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData, _ blockdag.MultiBlockTxsAcceptanceData) error {
	return dbPutTxsAcceptanceData(dbTx, blockID, txsAcceptanceData)
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

// Recover is invoked when the indexer wasn't turned on for several blocks
// and the indexer needs to close the gaps.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Recover(dbTx database.Tx, currentBlockID, lastKnownBlockID uint64) error {
	for blockID := currentBlockID + 1; blockID <= lastKnownBlockID; blockID++ {
		hash, err := blockdag.DBFetchBlockHashByID(dbTx, currentBlockID)
		if err != nil {
			return err
		}
		txAcceptanceData, err := idx.dag.TxsAcceptedByBlockHash(hash)
		if err != nil {
			return err
		}
		err = idx.ConnectBlock(dbTx, nil, blockID, nil, txAcceptanceData, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func dbPutTxsAcceptanceData(dbTx database.Tx, blockID uint64,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	serializedTxsAcceptanceData, err := serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		return err
	}

	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	return bucket.Put(blockdag.SerializeBlockID(blockID), serializedTxsAcceptanceData)
}

func dbFetchTxsAcceptanceDataByHash(dbTx database.Tx,
	hash *daghash.Hash) (blockdag.MultiBlockTxsAcceptanceData, error) {

	blockID, err := blockdag.DBFetchBlockIDByHash(dbTx, hash)
	if err != nil {
		return nil, err
	}

	return dbFetchTxsAcceptanceDataByID(dbTx, blockID)
}

func dbFetchTxsAcceptanceDataByID(dbTx database.Tx,
	blockID uint64) (blockdag.MultiBlockTxsAcceptanceData, error) {
	serializedBlockID := blockdag.SerializeBlockID(blockID)
	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	serializedTxsAcceptanceData := bucket.Get(serializedBlockID)
	if serializedTxsAcceptanceData == nil {
		return nil, errors.Errorf("no entry in the accpetance index for block id %d", blockID)
	}

	return deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
}

type serializableTxAcceptanceData struct {
	MsgTx      wire.MsgTx
	IsAccepted bool
}

type serializableBlockTxsAcceptanceData struct {
	BlockHash        daghash.Hash
	TxAcceptanceData []serializableTxAcceptanceData
}

type serializableMultiBlockTxsAcceptanceData []serializableBlockTxsAcceptanceData

func serializeMultiBlockTxsAcceptanceData(
	multiBlockTxsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) ([]byte, error) {
	// Convert MultiBlockTxsAcceptanceData to a serializable format
	serializableData := make(serializableMultiBlockTxsAcceptanceData, len(multiBlockTxsAcceptanceData))
	for i, blockTxsAcceptanceData := range multiBlockTxsAcceptanceData {
		serializableBlockData := serializableBlockTxsAcceptanceData{
			BlockHash:        blockTxsAcceptanceData.BlockHash,
			TxAcceptanceData: make([]serializableTxAcceptanceData, len(blockTxsAcceptanceData.TxAcceptanceData)),
		}
		for i, txAcceptanceData := range blockTxsAcceptanceData.TxAcceptanceData {
			serializableBlockData.TxAcceptanceData[i] = serializableTxAcceptanceData{
				MsgTx:      *txAcceptanceData.Tx.MsgTx(),
				IsAccepted: txAcceptanceData.IsAccepted,
			}
		}
		serializableData[i] = serializableBlockData
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
	multiBlockTxsAcceptanceData := make(blockdag.MultiBlockTxsAcceptanceData, len(serializedData))
	for i, serializableBlockData := range serializedData {
		blockTxsAcceptanceData := blockdag.BlockTxsAcceptanceData{
			BlockHash:        serializableBlockData.BlockHash,
			TxAcceptanceData: make([]blockdag.TxAcceptanceData, len(serializableBlockData.TxAcceptanceData)),
		}
		for i, txData := range serializableBlockData.TxAcceptanceData {
			msgTx := txData.MsgTx
			blockTxsAcceptanceData.TxAcceptanceData[i] = blockdag.TxAcceptanceData{
				Tx:         util.NewTx(&msgTx),
				IsAccepted: txData.IsAccepted,
			}
		}
		multiBlockTxsAcceptanceData[i] = blockTxsAcceptanceData
	}

	return multiBlockTxsAcceptanceData, nil
}
