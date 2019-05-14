// Copyright (c) 2015-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockdag

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/daglabs/btcd/database"
	"github.com/daglabs/btcd/util"
	"github.com/daglabs/btcd/util/binaryserializer"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/util/subnetworkid"
	"github.com/daglabs/btcd/wire"
)

const (
	// blockHdrSize is the size of a block header.  This is simply the
	// constant from wire and is only provided here for convenience since
	// wire.MaxBlockHeaderPayload is quite long.
	blockHdrSize = wire.MaxBlockHeaderPayload

	// latestUTXOSetBucketVersion is the current version of the UTXO set
	// bucket that is used to track all unspent outputs.
	latestUTXOSetBucketVersion = 1
)

var (
	// blockIndexBucketName is the name of the db bucket used to house to the
	// block headers and contextual information.
	blockIndexBucketName = []byte("blockheaderidx")

	// dagStateKeyName is the name of the db key used to store the DAG
	// tip hashes.
	dagStateKeyName = []byte("dagstate")

	// utxoSetVersionKeyName is the name of the db key used to store the
	// version of the utxo set currently in the database.
	utxoSetVersionKeyName = []byte("utxosetversion")

	// utxoSetBucketName is the name of the db bucket used to house the
	// unspent transaction output set.
	utxoSetBucketName = []byte("utxoset")

	// utxoDiffsBucketName is the name of the db bucket used to house the
	// diffs and diff children of blocks.
	utxoDiffsBucketName = []byte("utxodiffs")

	// subnetworksBucketName is the name of the db bucket used to store the
	// subnetwork registry.
	subnetworksBucketName = []byte("subnetworks")

	// localSubnetworkKeyName is the name of the db key used to store the
	// node's local subnetwork ID.
	localSubnetworkKeyName = []byte("localsubnetworkidkey")

	// byteOrder is the preferred byte order used for serializing numeric
	// fields for storage in the database.
	byteOrder = binary.LittleEndian
)

// errNotInDAG signifies that a block hash or height that is not in the
// DAG was requested.
type errNotInDAG string

// Error implements the error interface.
func (e errNotInDAG) Error() string {
	return string(e)
}

// isNotInDAGErr returns whether or not the passed error is an
// errNotInDAG error.
func isNotInDAGErr(err error) bool {
	_, ok := err.(errNotInDAG)
	return ok
}

// errDeserialize signifies that a problem was encountered when deserializing
// data.
type errDeserialize string

// Error implements the error interface.
func (e errDeserialize) Error() string {
	return string(e)
}

// isDeserializeErr returns whether or not the passed error is an errDeserialize
// error.
func isDeserializeErr(err error) bool {
	_, ok := err.(errDeserialize)
	return ok
}

// dbPutVersion uses an existing database transaction to update the provided
// key in the metadata bucket to the given version.  It is primarily used to
// track versions on entities such as buckets.
func dbPutVersion(dbTx database.Tx, key []byte, version uint32) error {
	var serialized [4]byte
	byteOrder.PutUint32(serialized[:], version)
	return dbTx.Metadata().Put(key, serialized[:])
}

// -----------------------------------------------------------------------------
// The unspent transaction output (UTXO) set consists of an entry for each
// unspent output using a format that is optimized to reduce space using domain
// specific compression algorithms.  This format is a slightly modified version
// of the format used in Bitcoin Core.
//
// Each entry is keyed by an outpoint as specified below.  It is important to
// note that the key encoding uses a VLQ, which employs an MSB encoding so
// iteration of UTXOs when doing byte-wise comparisons will produce them in
// order.
//
// The serialized key format is:
//   <hash><output index>
//
//   Field                Type             Size
//   hash                 daghash.Hash   daghash.HashSize
//   output index         VLQ              variable
//
// The serialized value format is:
//
//   <header code><compressed txout>
//
//   Field                Type     Size
//   header code          VLQ      variable
//   compressed txout
//     compressed amount  VLQ      variable
//     compressed script  []byte   variable
//
// The serialized header code format is:
//   bit 0 - containing transaction is a block reward
//   bits 1-x - height of the block that contains the unspent txout
//
// Example 1:
// From tx in main blockchain:
// Blk 1, b7c3332bc138e2c9429818f5fed500bcc1746544218772389054dc8047d7cd3f:0
//
//    03320496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52
//    <><------------------------------------------------------------------>
//     |                                          |
//   header code                         compressed txout
//
//  - header code: 0x03 (coinbase, height 1)
//  - compressed txout:
//    - 0x32: VLQ-encoded compressed amount for 5000000000 (50 BTC)
//    - 0x04: special script type pay-to-pubkey
//    - 0x96...52: x-coordinate of the pubkey
//
// Example 2:
// From tx in main blockchain:
// Blk 113931, 4a16969aa4764dd7507fc1de7f0baa4850a246de90c45e59a3207f9a26b5036f:2
//
//    8cf316800900b8025be1b3efc63b0ad48e7f9f10e87544528d58
//    <----><------------------------------------------>
//      |                             |
//   header code             compressed txout
//
//  - header code: 0x8cf316 (not coinbase, height 113931)
//  - compressed txout:
//    - 0x8009: VLQ-encoded compressed amount for 15000000 (0.15 BTC)
//    - 0x00: special script type pay-to-pubkey-hash
//    - 0xb8...58: pubkey hash
//
// Example 3:
// From tx in main blockchain:
// Blk 338156, 1b02d1c8cfef60a189017b9a420c682cf4a0028175f2f563209e4ff61c8c3620:22
//
//    a8a2588ba5b9e763011dd46a006572d820e448e12d2bbb38640bc718e6
//    <----><-------------------------------------------------->
//      |                             |
//   header code             compressed txout
//
//  - header code: 0xa8a258 (not coinbase, height 338156)
//  - compressed txout:
//    - 0x8ba5b9e763: VLQ-encoded compressed amount for 366875659 (3.66875659 BTC)
//    - 0x01: special script type pay-to-script-hash
//    - 0x1d...e6: script hash
// -----------------------------------------------------------------------------

// maxUint32VLQSerializeSize is the maximum number of bytes a max uint32 takes
// to serialize as a VLQ.
var maxUint32VLQSerializeSize = serializeSizeVLQ(1<<32 - 1)

// outpointKeyPool defines a concurrent safe free list of byte slices used to
// provide temporary buffers for outpoint database keys.
var outpointKeyPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, daghash.HashSize+maxUint32VLQSerializeSize)
		return &b // Pointer to slice to avoid boxing alloc.
	},
}

// outpointKey returns a key suitable for use as a database key in the UTXO set
// while making use of a free list.  A new buffer is allocated if there are not
// already any available on the free list.  The returned byte slice should be
// returned to the free list by using the recycleOutpointKey function when the
// caller is done with it _unless_ the slice will need to live for longer than
// the caller can calculate such as when used to write to the database.
func outpointKey(outpoint wire.OutPoint) *[]byte {
	// A VLQ employs an MSB encoding, so they are useful not only to reduce
	// the amount of storage space, but also so iteration of UTXOs when
	// doing byte-wise comparisons will produce them in order.
	key := outpointKeyPool.Get().(*[]byte)
	idx := uint64(outpoint.Index)
	*key = (*key)[:daghash.HashSize+serializeSizeVLQ(idx)]
	copy(*key, outpoint.TxID[:])
	putVLQ((*key)[daghash.HashSize:], idx)
	return key
}

// recycleOutpointKey puts the provided byte slice, which should have been
// obtained via the outpointKey function, back on the free list.
func recycleOutpointKey(key *[]byte) {
	outpointKeyPool.Put(key)
}

// utxoEntryHeaderCode returns the calculated header code to be used when
// serializing the provided utxo entry.
func utxoEntryHeaderCode(entry *UTXOEntry) uint64 {

	// As described in the serialization format comments, the header code
	// encodes the height shifted over one bit and the block reward flag in the
	// lowest bit.
	headerCode := uint64(entry.BlockChainHeight()) << 1
	if entry.IsBlockReward() {
		headerCode |= 0x01
	}

	return headerCode
}

// serializeUTXOEntry returns the entry serialized to a format that is suitable
// for long-term storage.  The format is described in detail above.
func serializeUTXOEntry(entry *UTXOEntry) ([]byte, error) {

	// Encode the header code.
	headerCode := utxoEntryHeaderCode(entry)

	// Calculate the size needed to serialize the entry.
	size := serializeSizeVLQ(headerCode) +
		compressedTxOutSize(uint64(entry.Amount()), entry.PkScript())

	// Serialize the header code followed by the compressed unspent
	// transaction output.
	serialized := make([]byte, size)
	offset := putVLQ(serialized, headerCode)
	offset += putCompressedTxOut(serialized[offset:], uint64(entry.Amount()),
		entry.PkScript())

	return serialized, nil
}

// deserializeOutPoint decodes an outPoint from the passed serialized byte
// slice into a new wire.OutPoint using a format that is suitable for long-
// term storage. this format is described in detail above.
func deserializeOutPoint(serialized []byte) (*wire.OutPoint, error) {
	if len(serialized) <= daghash.HashSize {
		return nil, errDeserialize("unexpected end of data")
	}

	txID := daghash.TxID{}
	txID.SetBytes(serialized[:daghash.HashSize])
	index, _ := deserializeVLQ(serialized[daghash.HashSize:])
	return wire.NewOutPoint(&txID, uint32(index)), nil
}

// deserializeUTXOEntry decodes a UTXO entry from the passed serialized byte
// slice into a new UTXOEntry using a format that is suitable for long-term
// storage.  The format is described in detail above.
func deserializeUTXOEntry(serialized []byte) (*UTXOEntry, error) {
	// Deserialize the header code.
	code, offset := deserializeVLQ(serialized)
	if offset >= len(serialized) {
		return nil, errDeserialize("unexpected end of data after header")
	}

	// Decode the header code.
	//
	// Bit 0 indicates whether the containing transaction is a block reward.
	// Bits 1-x encode height of containing transaction.
	isBlockReward := code&0x01 != 0
	blockChainHeight := code >> 1

	// Decode the compressed unspent transaction output.
	amount, pkScript, _, err := decodeCompressedTxOut(serialized[offset:])
	if err != nil {
		return nil, errDeserialize(fmt.Sprintf("unable to decode "+
			"UTXO: %s", err))
	}

	entry := &UTXOEntry{
		amount:           amount,
		pkScript:         pkScript,
		blockChainHeight: blockChainHeight,
		packedFlags:      0,
	}
	if isBlockReward {
		entry.packedFlags |= tfBlockReward
	}

	return entry, nil
}

// dbPutUTXODiff uses an existing database transaction to update the UTXO set
// in the database based on the provided UTXO view contents and state.  In
// particular, only the entries that have been marked as modified are written
// to the database.
func dbPutUTXODiff(dbTx database.Tx, diff *UTXODiff) error {
	utxoBucket := dbTx.Metadata().Bucket(utxoSetBucketName)
	for outPoint := range diff.toRemove {
		key := outpointKey(outPoint)
		err := utxoBucket.Delete(*key)
		recycleOutpointKey(key)
		if err != nil {
			return err
		}
	}

	for outPoint, entry := range diff.toAdd {
		// Serialize and store the UTXO entry.
		serialized, err := serializeUTXOEntry(entry)
		if err != nil {
			return err
		}

		key := outpointKey(outPoint)
		err = utxoBucket.Put(*key, serialized)
		// NOTE: The key is intentionally not recycled here since the
		// database interface contract prohibits modifications.  It will
		// be garbage collected normally when the database is done with
		// it.
		if err != nil {
			return err
		}
	}

	return nil
}

type dagState struct {
	TipHashes         []*daghash.Hash
	LastFinalityPoint *daghash.Hash
}

// serializeDAGState returns the serialization of the DAG state.
// This is data to be stored in the DAG state bucket.
func serializeDAGState(state *dagState) ([]byte, error) {
	return json.Marshal(state)
}

// deserializeDAGState deserializes the passed serialized DAG state.
// This is data stored in the DAG state bucket and is updated after
// every block is connected to the DAG.
func deserializeDAGState(serializedData []byte) (*dagState, error) {
	var state *dagState
	err := json.Unmarshal(serializedData, &state)
	if err != nil {
		return nil, database.Error{
			ErrorCode:   database.ErrCorruption,
			Description: "corrupt DAG state",
		}
	}

	return state, nil
}

// dbPutDAGState uses an existing database transaction to store the latest
// tip hashes of the DAG.
func dbPutDAGState(dbTx database.Tx, state *dagState) error {
	serializedData, err := serializeDAGState(state)

	if err != nil {
		return err
	}

	return dbTx.Metadata().Put(dagStateKeyName, serializedData)
}

// createDAGState initializes both the database and the DAG state to the
// genesis block.  This includes creating the necessary buckets, so it
// must only be called on an uninitialized database.
func (dag *BlockDAG) createDAGState() error {
	// Create the initial the database DAG state including creating the
	// necessary index buckets and inserting the genesis block.
	err := dag.db.Update(func(dbTx database.Tx) error {
		meta := dbTx.Metadata()

		// Create the bucket that houses the block index data.
		_, err := meta.CreateBucket(blockIndexBucketName)
		if err != nil {
			return err
		}

		// Create the buckets that house the utxo set, the utxo diffs, and their
		// version.
		_, err = meta.CreateBucket(utxoSetBucketName)
		if err != nil {
			return err
		}

		_, err = meta.CreateBucket(utxoDiffsBucketName)
		if err != nil {
			return err
		}

		err = dbPutVersion(dbTx, utxoSetVersionKeyName,
			latestUTXOSetBucketVersion)
		if err != nil {
			return err
		}

		// Create the bucket that houses the registered subnetworks.
		_, err = meta.CreateBucket(subnetworksBucketName)
		if err != nil {
			return err
		}

		if err := dbPutLocalSubnetworkID(dbTx, dag.subnetworkID); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func dbPutLocalSubnetworkID(dbTx database.Tx, subnetworkID *subnetworkid.SubnetworkID) error {
	if subnetworkID == nil {
		return dbTx.Metadata().Put(localSubnetworkKeyName, []byte{})
	}
	return dbTx.Metadata().Put(localSubnetworkKeyName, subnetworkID[:])
}

// initDAGState attempts to load and initialize the DAG state from the
// database.  When the db does not yet contain any DAG state, both it and the
// DAG state are initialized to the genesis block.
func (dag *BlockDAG) initDAGState() error {
	// Determine the state of the DAG database. We may need to initialize
	// everything from scratch or upgrade certain buckets.
	var initialized bool
	err := dag.db.View(func(dbTx database.Tx) error {
		initialized = dbTx.Metadata().Get(dagStateKeyName) != nil
		if initialized {
			var localSubnetworkID *subnetworkid.SubnetworkID
			localSubnetworkIDBytes := dbTx.Metadata().Get(localSubnetworkKeyName)
			if len(localSubnetworkIDBytes) != 0 {
				localSubnetworkID = &subnetworkid.SubnetworkID{}
				localSubnetworkID.SetBytes(localSubnetworkIDBytes)
			}
			if !localSubnetworkID.IsEqual(dag.subnetworkID) {
				return fmt.Errorf("Cannot start btcd with subnetwork ID %s because"+
					" its database is already built with subnetwork ID %s. If you"+
					" want to switch to a new database, please reset the"+
					" database by starting btcd with --reset-db flag", dag.subnetworkID, localSubnetworkID)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if !initialized {
		// At this point the database has not already been initialized, so
		// initialize both it and the chain state to the genesis block.
		return dag.createDAGState()
	}

	// Attempt to load the DAG state from the database.
	return dag.db.View(func(dbTx database.Tx) error {
		// Fetch the stored DAG tipHashes from the database metadata.
		// When it doesn't exist, it means the database hasn't been
		// initialized for use with the DAG yet, so break out now to allow
		// that to happen under a writable database transaction.
		serializedData := dbTx.Metadata().Get(dagStateKeyName)
		log.Tracef("Serialized DAG tip hashes: %x", serializedData)
		state, err := deserializeDAGState(serializedData)
		if err != nil {
			return err
		}

		// Load all of the headers from the data for the known DAG
		// and construct the block index accordingly.  Since the
		// number of nodes are already known, perform a single alloc
		// for them versus a whole bunch of little ones to reduce
		// pressure on the GC.
		log.Infof("Loading block index...")

		blockIndexBucket := dbTx.Metadata().Bucket(blockIndexBucketName)

		var i int32
		var lastNode *blockNode
		cursor := blockIndexBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			node, err := dag.deserializeBlockNode(cursor.Value())
			if err != nil {
				return err
			}

			if lastNode == nil {
				if !node.hash.IsEqual(dag.dagParams.GenesisHash) {
					return AssertError(fmt.Sprintf("initDAGState: Expected "+
						"first entry in block index to be genesis block, "+
						"found %s", node.hash))
				}
			} else {
				if len(node.parents) == 0 {
					return AssertError(fmt.Sprintf("initDAGState: Could "+
						"not find any parent for block %s", node.hash))
				}
			}

			// Add the node to its parents children, connect it,
			// and add it to the block index.
			node.updateParentsChildren()
			dag.index.addNode(node)

			if node.status.KnownValid() {
				dag.blockCount++
			}

			lastNode = node
			i++
		}

		// Load all of the known UTXO entries and construct the full
		// UTXO set accordingly.  Since the number of entries is already
		// known, perform a single alloc for them versus a whole bunch
		// of little ones to reduce pressure on the GC.
		log.Infof("Loading UTXO set...")

		utxoEntryBucket := dbTx.Metadata().Bucket(utxoSetBucketName)

		// Determine how many UTXO entries will be loaded into the index so we can
		// allocate the right amount.
		var utxoEntryCount int32
		cursor = utxoEntryBucket.Cursor()
		for ok := cursor.First(); ok; ok = cursor.Next() {
			utxoEntryCount++
		}

		fullUTXOCollection := make(utxoCollection, utxoEntryCount)
		for ok := cursor.First(); ok; ok = cursor.Next() {
			// Deserialize the outPoint
			outPoint, err := deserializeOutPoint(cursor.Key())
			if err != nil {
				// Ensure any deserialization errors are returned as database
				// corruption errors.
				if isDeserializeErr(err) {
					return database.Error{
						ErrorCode:   database.ErrCorruption,
						Description: fmt.Sprintf("corrupt outPoint: %s", err),
					}
				}

				return err
			}

			// Deserialize the utxo entry
			entry, err := deserializeUTXOEntry(cursor.Value())
			if err != nil {
				// Ensure any deserialization errors are returned as database
				// corruption errors.
				if isDeserializeErr(err) {
					return database.Error{
						ErrorCode:   database.ErrCorruption,
						Description: fmt.Sprintf("corrupt utxo entry: %s", err),
					}
				}

				return err
			}

			fullUTXOCollection[*outPoint] = entry
		}

		// Apply the loaded utxoCollection to the virtual block.
		dag.virtual.utxoSet.utxoCollection = fullUTXOCollection

		// Apply the stored tips to the virtual block.
		tips := newSet()
		for _, tipHash := range state.TipHashes {
			tip := dag.index.LookupNode(tipHash)
			if tip == nil {
				return AssertError(fmt.Sprintf("initDAGState: cannot find "+
					"DAG tip %s in block index", state.TipHashes))
			}
			tips.add(tip)
		}
		dag.virtual.SetTips(tips)

		// Set the last finality point
		dag.lastFinalityPoint = dag.index.LookupNode(state.LastFinalityPoint)

		return nil
	})
}

// deserializeBlockNode parses a value in the block index bucket and returns a block node.
func (dag *BlockDAG) deserializeBlockNode(blockRow []byte) (*blockNode, error) {
	buffer := bytes.NewReader(blockRow)

	var header wire.BlockHeader
	err := header.Deserialize(buffer)
	if err != nil {
		return nil, err
	}

	node := &blockNode{
		hash:                 header.BlockHash(),
		version:              header.Version,
		bits:                 header.Bits,
		nonce:                header.Nonce,
		timestamp:            header.Timestamp.Unix(),
		hashMerkleRoot:       header.HashMerkleRoot,
		idMerkleRoot:         header.IDMerkleRoot,
		acceptedIDMerkleRoot: header.AcceptedIDMerkleRoot,
		utxoCommitment:       header.UTXOCommitment,
	}

	node.children = newSet()
	node.parents = newSet()

	for _, hash := range header.ParentHashes {
		parent := dag.index.LookupNode(hash)
		if parent == nil {
			return nil, AssertError(fmt.Sprintf("deserializeBlockNode: Could "+
				"not find parent %s for block %s", hash, header.BlockHash()))
		}
		node.parents.add(parent)
	}

	statusByte, err := buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	node.status = blockStatus(statusByte)

	selectedParentHash := &daghash.Hash{}
	if _, err := io.ReadFull(buffer, selectedParentHash[:]); err != nil {
		return nil, err
	}

	// Because genesis doesn't have selected parent, it's serialized as zero hash
	if !selectedParentHash.IsEqual(&daghash.ZeroHash) {
		node.selectedParent = dag.index.LookupNode(selectedParentHash)
	}

	node.blueScore, err = binaryserializer.Uint64(buffer, byteOrder)
	if err != nil {
		return nil, err
	}

	bluesCount, err := wire.ReadVarInt(buffer)
	if err != nil {
		return nil, err
	}

	node.blues = make([]*blockNode, bluesCount)
	for i := uint64(0); i < bluesCount; i++ {
		hash := &daghash.Hash{}
		if _, err := io.ReadFull(buffer, hash[:]); err != nil {
			return nil, err
		}
		node.blues[i] = dag.index.LookupNode(hash)
	}

	node.height = calculateNodeHeight(node)
	node.chainHeight = calculateChainHeight(node)

	return node, nil
}

// dbFetchBlockByNode uses an existing database transaction to retrieve the
// raw block for the provided node, deserialize it, and return a util.Block
// with the height set.
func dbFetchBlockByNode(dbTx database.Tx, node *blockNode) (*util.Block, error) {
	// Load the raw block bytes from the database.
	blockBytes, err := dbTx.FetchBlock(node.hash)
	if err != nil {
		return nil, err
	}

	// Create the encapsulated block and set the height appropriately.
	block, err := util.NewBlockFromBytes(blockBytes)
	if err != nil {
		return nil, err
	}
	block.SetHeight(node.height)

	return block, nil
}

// dbStoreBlockNode stores the block node data into the block
// index bucket. This overwrites the current entry if there exists one.
func dbStoreBlockNode(dbTx database.Tx, node *blockNode) error {
	// Serialize block data to be stored.
	w := bytes.NewBuffer(make([]byte, 0, blockHdrSize+1))
	header := node.Header()
	err := header.Serialize(w)
	if err != nil {
		return err
	}

	err = w.WriteByte(byte(node.status))
	if err != nil {
		return err
	}

	// Because genesis doesn't have selected parent, it's serialized as zero hash
	selectedParentHash := &daghash.ZeroHash
	if node.selectedParent != nil {
		selectedParentHash = node.selectedParent.hash
	}
	_, err = w.Write(selectedParentHash[:])
	if err != nil {
		return err
	}

	err = binaryserializer.PutUint64(w, byteOrder, node.blueScore)
	if err != nil {
		return err
	}

	err = wire.WriteVarInt(w, uint64(len(node.blues)))
	if err != nil {
		return err
	}

	for _, blue := range node.blues {
		_, err = w.Write(blue.hash[:])
		if err != nil {
			return err
		}
	}

	value := w.Bytes()

	// Write block header data to block index bucket.
	blockIndexBucket := dbTx.Metadata().Bucket(blockIndexBucketName)
	key := blockIndexKey(node.hash, uint32(node.height))
	return blockIndexBucket.Put(key, value)
}

// dbStoreBlock stores the provided block in the database if it is not already
// there. The full block data is written to ffldb.
func dbStoreBlock(dbTx database.Tx, block *util.Block) error {
	hasBlock, err := dbTx.HasBlock(block.Hash())
	if err != nil {
		return err
	}
	if hasBlock {
		return nil
	}
	return dbTx.StoreBlock(block)
}

// blockIndexKey generates the binary key for an entry in the block index
// bucket. The key is composed of the block height encoded as a big-endian
// 32-bit unsigned int followed by the 32 byte block hash.
func blockIndexKey(blockHash *daghash.Hash, blockHeight uint32) []byte {
	indexKey := make([]byte, daghash.HashSize+4)
	binary.BigEndian.PutUint32(indexKey[0:4], blockHeight)
	copy(indexKey[4:daghash.HashSize+4], blockHash[:])
	return indexKey
}

// BlockByHash returns the block from the main chain with the given hash with
// the appropriate chain height set.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockByHash(hash *daghash.Hash) (*util.Block, error) {
	// Lookup the block hash in block index and ensure it is in the best
	// chain.
	node := dag.index.LookupNode(hash)
	if node == nil {
		str := fmt.Sprintf("block %s is not in the main chain", hash)
		return nil, errNotInDAG(str)
	}

	// Load the block from the database and return it.
	var block *util.Block
	err := dag.db.View(func(dbTx database.Tx) error {
		var err error
		block, err = dbFetchBlockByNode(dbTx, node)
		return err
	})
	return block, err
}
