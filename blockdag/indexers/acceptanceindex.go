package indexers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
)

const (
	// acceptanceIndexName is the human-readable name for the index.
	acceptanceIndexName = "address index"
)

var (
	// acceptanceIndexKey is the key of the acceptance index and the db bucket used
	// to house it.
	acceptanceIndexKey = []byte("acceptanceidx")
)

type AcceptanceIndex struct {
	db database.DB
}

// NewAcceptanceIndex returns a new instance of an indexer that is used to create a
// mapping between block hashes and their txAcceptanceData.
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockchain package.  This allows the index to be
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
// to be created for the first time.  It creates the bucket for the address
// index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Create(dbTx database.Tx) error {
	_, err := dbTx.Metadata().CreateBucket(acceptanceIndexKey)
	return err
}

// Init initializes the hash-based acceptance index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Init(db database.DB, _ *blockdag.BlockDAG) error {
	idx.db = db
	return nil
}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) ConnectBlock(dbTx database.Tx, block *util.Block, _ *blockdag.BlockDAG,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData, _ blockdag.MultiBlockTxsAcceptanceData) error {
	return idx.dbPutTxsAcceptanceData(dbTx, block.Hash(), txsAcceptanceData)
}

func (idx *AcceptanceIndex) dbPutTxsAcceptanceData(dbTx database.Tx, hash *daghash.Hash,
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) error {
	serializedTxsAcceptanceData, err := idx.serializeMultiBlockTxsAcceptanceData(txsAcceptanceData)
	if err != nil {
		return err
	}

	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	return bucket.Put(hash[:], serializedTxsAcceptanceData)
}

// TxsAcceptanceData returns the acceptance data of all the transactions that
// were accepted by the block with hash blockHash.
func (idx *AcceptanceIndex) TxsAcceptanceData(blockHash *daghash.Hash) (blockdag.MultiBlockTxsAcceptanceData, error) {
	var txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		txsAcceptanceData, err = idx.dbFetchTxsAcceptanceData(dbTx, blockHash)
		return err
	})
	if err != nil {
		return nil, err
	}
	return txsAcceptanceData, nil
}

func (idx *AcceptanceIndex) dbFetchTxsAcceptanceData(dbTx database.Tx,
	hash *daghash.Hash) (blockdag.MultiBlockTxsAcceptanceData, error) {
	bucket := dbTx.Metadata().Bucket(acceptanceIndexKey)
	serializedTxsAcceptanceData := bucket.Get(hash[:])
	if serializedTxsAcceptanceData == nil {
		return nil, fmt.Errorf("no entry in the accpetance index for block with hash %s", hash)
	}

	return idx.deserializeMultiBlockTxsAcceptanceData(serializedTxsAcceptanceData)
}

func (idx *AcceptanceIndex) serializeMultiBlockTxsAcceptanceData(
	txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(txsAcceptanceData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (idx *AcceptanceIndex) deserializeMultiBlockTxsAcceptanceData(
	serializedTxsAcceptanceData []byte) (blockdag.MultiBlockTxsAcceptanceData, error) {
	buffer := bytes.NewBuffer(serializedTxsAcceptanceData)
	decoder := gob.NewDecoder(buffer)

	var txsAcceptanceData blockdag.MultiBlockTxsAcceptanceData
	err := decoder.Decode(&txsAcceptanceData)
	if err != nil {
		return nil, err
	}
	return txsAcceptanceData, nil
}
