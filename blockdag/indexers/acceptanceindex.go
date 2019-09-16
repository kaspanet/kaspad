package indexers

import (
	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
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
}

// NewAcceptanceIndex ...
//
// It implements the Indexer interface which plugs into the IndexManager that in
// turn is used by the blockchain package.  This allows the index to be
// seamlessly maintained along with the chain.
func NewAcceptanceIndex(dagParams *dagconfig.Params) *AcceptanceIndex {
	return nil
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

// Init is only provided to satisfy the Indexer interface as there is nothing to
// initialize for this index.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) Init(_ database.DB, _ *blockdag.BlockDAG) error {
	return nil
}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the DAG.
//
// This is part of the Indexer interface.
func (idx *AcceptanceIndex) ConnectBlock(_ database.Tx, _ *util.Block, _ *blockdag.BlockDAG,
	_ blockdag.MultiBlockTxsAcceptanceData, _ blockdag.MultiBlockTxsAcceptanceData) error {
	return nil
}
