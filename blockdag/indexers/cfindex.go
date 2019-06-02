// Copyright (c) 2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package indexers

import (
	"errors"

	"github.com/daglabs/btcd/blockdag"
	"github.com/daglabs/btcd/dagconfig"
	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/gcs"
	"github.com/daglabs/btcd/util/gcs/builder"
	"github.com/daglabs/btcd/wire"
)

const (
	// cfIndexName is the human-readable name for the index.
	cfIndexName = "committed filter index"
)

// Committed filters come in two flavours: basic and extended. They are
// generated and dropped in pairs, and both are indexed by a block's hash.
// Besides holding different content, they also live in different buckets.
var (
	// cfIndexParentBucketKey is the name of the parent bucket used to house
	// the index. The rest of the buckets live below this bucket.
	cfIndexParentBucketKey = []byte("cfindexparentbucket")

	// cfIndexKeys is an array of db bucket names used to house indexes of
	// block hashes to cfilters.
	cfIndexKeys = [][]byte{
		[]byte("cf0byhashidx"),
		[]byte("cf1byhashidx"),
	}

	// cfHeaderKeys is an array of db bucket names used to house indexes of
	// block hashes to cf headers.
	cfHeaderKeys = [][]byte{
		[]byte("cf0headerbyhashidx"),
		[]byte("cf1headerbyhashidx"),
	}

	// cfHashKeys is an array of db bucket names used to house indexes of
	// block hashes to cf hashes.
	cfHashKeys = [][]byte{
		[]byte("cf0hashbyhashidx"),
		[]byte("cf1hashbyhashidx"),
	}

	maxFilterType = uint8(len(cfHeaderKeys) - 1)
)

// dbFetchFilterIdxEntry retrieves a data blob from the filter index database.
// An entry's absence is not considered an error.
func dbFetchFilterIdxEntry(dbTx database.Tx, key []byte, h *daghash.Hash) ([]byte, error) {
	idx := dbTx.Metadata().Bucket(cfIndexParentBucketKey).Bucket(key)
	return idx.Get(h[:]), nil
}

// dbStoreFilterIdxEntry stores a data blob in the filter index database.
func dbStoreFilterIdxEntry(dbTx database.Tx, key []byte, h *daghash.Hash, f []byte) error {
	idx := dbTx.Metadata().Bucket(cfIndexParentBucketKey).Bucket(key)
	return idx.Put(h[:], f)
}

// dbDeleteFilterIdxEntry deletes a data blob from the filter index database.
func dbDeleteFilterIdxEntry(dbTx database.Tx, key []byte, h *daghash.Hash) error {
	idx := dbTx.Metadata().Bucket(cfIndexParentBucketKey).Bucket(key)
	return idx.Delete(h[:])
}

// CfIndex implements a committed filter (cf) by hash index.
type CfIndex struct {
	db        database.DB
	dagParams *dagconfig.Params
}

// Ensure the CfIndex type implements the Indexer interface.
var _ Indexer = (*CfIndex)(nil)

// Init initializes the hash-based cf index. This is part of the Indexer
// interface.
func (idx *CfIndex) Init(db database.DB, _ *blockdag.BlockDAG) error {
	idx.db = db
	return nil
}

// Key returns the database key to use for the index as a byte slice. This is
// part of the Indexer interface.
func (idx *CfIndex) Key() []byte {
	return cfIndexParentBucketKey
}

// Name returns the human-readable name of the index. This is part of the
// Indexer interface.
func (idx *CfIndex) Name() string {
	return cfIndexName
}

// Create is invoked when the indexer manager determines the index needs to
// be created for the first time. It creates buckets for the two hash-based cf
// indexes (simple, extended).
func (idx *CfIndex) Create(dbTx database.Tx) error {
	meta := dbTx.Metadata()

	cfIndexParentBucket, err := meta.CreateBucket(cfIndexParentBucketKey)
	if err != nil {
		return err
	}

	for _, bucketName := range cfIndexKeys {
		_, err = cfIndexParentBucket.CreateBucket(bucketName)
		if err != nil {
			return err
		}
	}

	for _, bucketName := range cfHeaderKeys {
		_, err = cfIndexParentBucket.CreateBucket(bucketName)
		if err != nil {
			return err
		}
	}

	for _, bucketName := range cfHashKeys {
		_, err = cfIndexParentBucket.CreateBucket(bucketName)
		if err != nil {
			return err
		}
	}

	return nil
}

// storeFilter stores a given filter, and performs the steps needed to
// generate the filter's header.
func storeFilter(dbTx database.Tx, block *util.Block, f *gcs.Filter,
	filterType wire.FilterType) error {
	if uint8(filterType) > maxFilterType {
		return errors.New("unsupported filter type")
	}

	// Figure out which buckets to use.
	fkey := cfIndexKeys[filterType]
	hkey := cfHeaderKeys[filterType]
	hashkey := cfHashKeys[filterType]

	// Start by storing the filter.
	h := block.Hash()
	filterBytes, err := f.NBytes()
	if err != nil {
		return err
	}
	err = dbStoreFilterIdxEntry(dbTx, fkey, h, filterBytes)
	if err != nil {
		return err
	}

	// Next store the filter hash.
	filterHash, err := builder.GetFilterHash(f)
	if err != nil {
		return err
	}
	err = dbStoreFilterIdxEntry(dbTx, hashkey, h, filterHash[:])
	if err != nil {
		return err
	}

	// Then fetch the previous block's filter header.
	var prevHeader *daghash.Hash
	header := block.MsgBlock().Header
	if header.IsGenesis() {
		prevHeader = &daghash.ZeroHash
	} else {
		// TODO(Evgeny): Current implementation of GCS filter inherited from chain
		// (single parent) and must be ported to DAG (multiple parents)
		var parentHash *daghash.Hash
		if header.NumParentBlocks() != 0 {
			parentHash = header.ParentHashes[0]
		}
		prevFilterHashBytes, err := dbFetchFilterIdxEntry(dbTx, hkey, parentHash)
		if err != nil {
			return err
		}

		// Construct the new block's filter header, and store it.
		prevHeader, err = daghash.NewHash(prevFilterHashBytes)
		if err != nil {
			return err
		}
	}

	fh, err := builder.MakeHeaderForFilter(f, prevHeader)
	if err != nil {
		return err
	}
	return dbStoreFilterIdxEntry(dbTx, hkey, h, fh[:])
}

// ConnectBlock is invoked by the index manager when a new block has been
// connected to the main chain. This indexer adds a hash-to-cf mapping for
// every passed block. This is part of the Indexer interface.
func (idx *CfIndex) ConnectBlock(dbTx database.Tx, block *util.Block, _ *blockdag.BlockDAG, _ *util.Tx,
	_ blockdag.MultiBlockTxsAcceptanceData, _ blockdag.MultiBlockTxsAcceptanceData) error {

	f, err := builder.BuildBasicFilter(block.MsgBlock())
	if err != nil {
		return err
	}

	err = storeFilter(dbTx, block, f, wire.GCSFilterRegular)
	if err != nil {
		return err
	}

	f, err = builder.BuildExtFilter(block.MsgBlock())
	if err != nil {
		return err
	}

	return storeFilter(dbTx, block, f, wire.GCSFilterExtended)
}

// entryByBlockHash fetches a filter index entry of a particular type
// (eg. filter, filter header, etc) for a filter type and block hash.
func (idx *CfIndex) entryByBlockHash(filterTypeKeys [][]byte,
	filterType wire.FilterType, h *daghash.Hash) ([]byte, error) {

	if uint8(filterType) > maxFilterType {
		return nil, errors.New("unsupported filter type")
	}
	key := filterTypeKeys[filterType]

	var entry []byte
	err := idx.db.View(func(dbTx database.Tx) error {
		var err error
		entry, err = dbFetchFilterIdxEntry(dbTx, key, h)
		return err
	})
	return entry, err
}

// entriesByBlockHashes batch fetches a filter index entry of a particular type
// (eg. filter, filter header, etc) for a filter type and slice of block hashes.
func (idx *CfIndex) entriesByBlockHashes(filterTypeKeys [][]byte,
	filterType wire.FilterType, blockHashes []*daghash.Hash) ([][]byte, error) {

	if uint8(filterType) > maxFilterType {
		return nil, errors.New("unsupported filter type")
	}
	key := filterTypeKeys[filterType]

	entries := make([][]byte, 0, len(blockHashes))
	err := idx.db.View(func(dbTx database.Tx) error {
		for _, blockHash := range blockHashes {
			entry, err := dbFetchFilterIdxEntry(dbTx, key, blockHash)
			if err != nil {
				return err
			}
			entries = append(entries, entry)
		}
		return nil
	})
	return entries, err
}

// FilterByBlockHash returns the serialized contents of a block's basic or
// extended committed filter.
func (idx *CfIndex) FilterByBlockHash(h *daghash.Hash,
	filterType wire.FilterType) ([]byte, error) {
	return idx.entryByBlockHash(cfIndexKeys, filterType, h)
}

// FiltersByBlockHashes returns the serialized contents of a block's basic or
// extended committed filter for a set of blocks by hash.
func (idx *CfIndex) FiltersByBlockHashes(blockHashes []*daghash.Hash,
	filterType wire.FilterType) ([][]byte, error) {
	return idx.entriesByBlockHashes(cfIndexKeys, filterType, blockHashes)
}

// FilterHeaderByBlockHash returns the serialized contents of a block's basic
// or extended committed filter header.
func (idx *CfIndex) FilterHeaderByBlockHash(h *daghash.Hash,
	filterType wire.FilterType) ([]byte, error) {
	return idx.entryByBlockHash(cfHeaderKeys, filterType, h)
}

// FilterHeadersByBlockHashes returns the serialized contents of a block's basic
// or extended committed filter header for a set of blocks by hash.
func (idx *CfIndex) FilterHeadersByBlockHashes(blockHashes []*daghash.Hash,
	filterType wire.FilterType) ([][]byte, error) {
	return idx.entriesByBlockHashes(cfHeaderKeys, filterType, blockHashes)
}

// FilterHashByBlockHash returns the serialized contents of a block's basic
// or extended committed filter hash.
func (idx *CfIndex) FilterHashByBlockHash(h *daghash.Hash,
	filterType wire.FilterType) ([]byte, error) {
	return idx.entryByBlockHash(cfHashKeys, filterType, h)
}

// FilterHashesByBlockHashes returns the serialized contents of a block's basic
// or extended committed filter hash for a set of blocks by hash.
func (idx *CfIndex) FilterHashesByBlockHashes(blockHashes []*daghash.Hash,
	filterType wire.FilterType) ([][]byte, error) {
	return idx.entriesByBlockHashes(cfHashKeys, filterType, blockHashes)
}

// NewCfIndex returns a new instance of an indexer that is used to create a
// mapping of the hashes of all blocks in the blockchain to their respective
// committed filters.
//
// It implements the Indexer interface which plugs into the IndexManager that
// in turn is used by the blockchain package. This allows the index to be
// seamlessly maintained along with the chain.
func NewCfIndex(dagParams *dagconfig.Params) *CfIndex {
	return &CfIndex{dagParams: dagParams}
}

// DropCfIndex drops the CF index from the provided database if exists.
func DropCfIndex(db database.DB, interrupt <-chan struct{}) error {
	return dropIndex(db, cfIndexParentBucketKey, cfIndexName, interrupt)
}
