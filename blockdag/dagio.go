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

	"github.com/kaspanet/kaspad/dagconfig"
	"github.com/kaspanet/kaspad/dbaccess"
	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/util"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/util/subnetworkid"
	"github.com/kaspanet/kaspad/wire"
)

var (
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
	var notInDAGErr errNotInDAG
	return errors.As(err, &notInDAGErr)
}

// outpointIndexByteOrder is the byte order for serializing the outpoint index.
// It uses big endian to ensure that when outpoint is used as database key, the
// keys will be iterated in an ascending order by the outpoint index.
var outpointIndexByteOrder = binary.BigEndian

func serializeOutpoint(w io.Writer, outpoint *wire.Outpoint) error {
	_, err := w.Write(outpoint.TxID[:])
	if err != nil {
		return err
	}

	return binaryserializer.PutUint32(w, outpointIndexByteOrder, outpoint.Index)
}

var outpointSerializeSize = daghash.TxIDSize + 4

// deserializeOutpoint decodes an outpoint from the passed serialized byte
// slice into a new wire.Outpoint using a format that is suitable for long-
// term storage. This format is described in detail above.
func deserializeOutpoint(r io.Reader) (*wire.Outpoint, error) {
	outpoint := &wire.Outpoint{}
	_, err := r.Read(outpoint.TxID[:])
	if err != nil {
		return nil, err
	}

	outpoint.Index, err = binaryserializer.Uint32(r, outpointIndexByteOrder)
	if err != nil {
		return nil, err
	}

	return outpoint, nil
}

// updateUTXOSet updates the UTXO set in the database based on the provided
// UTXO diff.
func updateUTXOSet(dbContext dbaccess.Context, virtualUTXODiff *UTXODiff) error {
	outpointBuff := bytes.NewBuffer(make([]byte, outpointSerializeSize))
	for outpoint := range virtualUTXODiff.toRemove {
		outpointBuff.Reset()
		err := serializeOutpoint(outpointBuff, &outpoint)
		if err != nil {
			return err
		}

		key := outpointBuff.Bytes()
		err = dbaccess.RemoveFromUTXOSet(dbContext, key)
		if err != nil {
			return err
		}
	}

	// We are preallocating for P2PKH entries because they are the most common ones.
	// If we have entries with a compressed script bigger than P2PKH's, the buffer will grow.
	utxoEntryBuff := bytes.NewBuffer(make([]byte, p2pkhUTXOEntrySerializeSize))

	for outpoint, entry := range virtualUTXODiff.toAdd {
		utxoEntryBuff.Reset()
		outpointBuff.Reset()
		// Serialize and store the UTXO entry.
		err := serializeUTXOEntry(utxoEntryBuff, entry)
		if err != nil {
			return err
		}
		serializedEntry := utxoEntryBuff.Bytes()

		err = serializeOutpoint(outpointBuff, &outpoint)
		if err != nil {
			return err
		}

		key := outpointBuff.Bytes()
		err = dbaccess.AddToUTXOSet(dbContext, key, serializedEntry)
		if err != nil {
			return err
		}
	}

	return nil
}

type dagState struct {
	TipHashes         []*daghash.Hash
	LastFinalityPoint *daghash.Hash
	LocalSubnetworkID *subnetworkid.SubnetworkID
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
		return nil, err
	}

	return state, nil
}

// saveDAGState uses an existing database context to store the latest
// tip hashes of the DAG.
func saveDAGState(dbContext dbaccess.Context, state *dagState) error {
	serializedDAGState, err := serializeDAGState(state)
	if err != nil {
		return err
	}

	return dbaccess.StoreDAGState(dbContext, serializedDAGState)
}

// createDAGState initializes the DAG state to the
// genesis block and the node's local subnetwork id.
func (dag *BlockDAG) createDAGState(localSubnetworkID *subnetworkid.SubnetworkID) error {
	return saveDAGState(dbaccess.NoTx(), &dagState{
		TipHashes:         []*daghash.Hash{dag.dagParams.GenesisHash},
		LastFinalityPoint: dag.dagParams.GenesisHash,
		LocalSubnetworkID: localSubnetworkID,
	})
}

// initDAGState attempts to load and initialize the DAG state from the
// database. When the db does not yet contain any DAG state, both it and the
// DAG state are initialized to the genesis block.
func (dag *BlockDAG) initDAGState() error {
	// Fetch the stored DAG state from the database. If it doesn't exist,
	// it means that kaspad is running for the first time.
	serializedDAGState, err := dbaccess.FetchDAGState(dbaccess.NoTx())
	if dbaccess.IsNotFoundError(err) {
		// Initialize the database and the DAG state to the genesis block.
		return dag.createDAGState(dag.subnetworkID)
	}
	if err != nil {
		return err
	}

	dagState, err := deserializeDAGState(serializedDAGState)
	if err != nil {
		return err
	}

	err = dag.validateLocalSubnetworkID(dagState)
	if err != nil {
		return err
	}

	log.Debugf("Loading block index...")
	unprocessedBlockNodes, err := dag.initBlockIndex()
	if err != nil {
		return err
	}

	log.Debugf("Loading UTXO set...")
	fullUTXOCollection, err := dag.initUTXOSet()
	if err != nil {
		return err
	}

	log.Debugf("Loading reachability data...")
	err = dag.reachabilityTree.init(dbaccess.NoTx())
	if err != nil {
		return err
	}

	log.Debugf("Loading multiset data...")
	err = dag.multisetStore.init(dbaccess.NoTx())
	if err != nil {
		return err
	}

	log.Debugf("Applying the loaded utxoCollection to the virtual block...")
	dag.virtual.utxoSet, err = newFullUTXOSetFromUTXOCollection(fullUTXOCollection)
	if err != nil {
		return errors.Wrap(err, "Error loading UTXOSet")
	}

	log.Debugf("Applying the stored tips to the virtual block...")
	err = dag.initVirtualBlockTips(dagState)
	if err != nil {
		return err
	}

	log.Debugf("Setting the last finality point...")
	var ok bool
	dag.lastFinalityPoint, ok = dag.index.LookupNode(dagState.LastFinalityPoint)
	if !ok {
		return errors.Errorf("finality point block %s "+
			"does not exist in the DAG", dagState.LastFinalityPoint)
	}
	dag.finalizeNodesBelowFinalityPoint(false)

	log.Debugf("Processing unprocessed blockNodes...")
	err = dag.processUnprocessedBlockNodes(unprocessedBlockNodes)
	if err != nil {
		return err
	}

	log.Infof("DAG state initialized.")

	return nil
}

func (dag *BlockDAG) validateLocalSubnetworkID(state *dagState) error {
	if !state.LocalSubnetworkID.IsEqual(dag.subnetworkID) {
		return errors.Errorf("Cannot start kaspad with subnetwork ID %s because"+
			" its database is already built with subnetwork ID %s. If you"+
			" want to switch to a new database, please reset the"+
			" database by starting kaspad with --reset-db flag", dag.subnetworkID, state.LocalSubnetworkID)
	}
	return nil
}

func (dag *BlockDAG) initBlockIndex() (unprocessedBlockNodes []*blockNode, err error) {
	blockIndexCursor, err := dbaccess.BlockIndexCursor(dbaccess.NoTx())
	if err != nil {
		return nil, err
	}
	defer blockIndexCursor.Close()
	for blockIndexCursor.Next() {
		serializedDBNode, err := blockIndexCursor.Value()
		if err != nil {
			return nil, err
		}
		node, err := dag.deserializeBlockNode(serializedDBNode)
		if err != nil {
			return nil, err
		}

		// Check to see if this node had been stored in the the block DB
		// but not yet accepted. If so, add it to a slice to be processed later.
		if node.status == statusDataStored {
			unprocessedBlockNodes = append(unprocessedBlockNodes, node)
			continue
		}

		// If the node is known to be invalid add it as-is to the block
		// index and continue.
		if node.status.KnownInvalid() {
			dag.index.addNode(node)
			continue
		}

		if dag.blockCount == 0 {
			if !node.hash.IsEqual(dag.dagParams.GenesisHash) {
				return nil, errors.Errorf("Expected "+
					"first entry in block index to be genesis block, "+
					"found %s", node.hash)
			}
		} else {
			if len(node.parents) == 0 {
				return nil, errors.Errorf("block %s "+
					"has no parents but it's not the genesis block", node.hash)
			}
		}

		// Add the node to its parents children, connect it,
		// and add it to the block index.
		node.updateParentsChildren()
		dag.index.addNode(node)

		dag.blockCount++
	}
	return unprocessedBlockNodes, nil
}

func (dag *BlockDAG) initUTXOSet() (fullUTXOCollection utxoCollection, err error) {
	fullUTXOCollection = make(utxoCollection)
	cursor, err := dbaccess.UTXOSetCursor(dbaccess.NoTx())
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for cursor.Next() {
		// Deserialize the outpoint
		key, err := cursor.Key()
		if err != nil {
			return nil, err
		}
		outpoint, err := deserializeOutpoint(bytes.NewReader(key.Suffix()))
		if err != nil {
			return nil, err
		}

		// Deserialize the utxo entry
		value, err := cursor.Value()
		if err != nil {
			return nil, err
		}
		entry, err := deserializeUTXOEntry(bytes.NewReader(value))
		if err != nil {
			return nil, err
		}

		fullUTXOCollection[*outpoint] = entry
	}

	return fullUTXOCollection, nil
}

func (dag *BlockDAG) initVirtualBlockTips(state *dagState) error {
	tips := newBlockSet()
	for _, tipHash := range state.TipHashes {
		tip, ok := dag.index.LookupNode(tipHash)
		if !ok {
			return errors.Errorf("cannot find "+
				"DAG tip %s in block index", state.TipHashes)
		}
		tips.add(tip)
	}
	dag.virtual.SetTips(tips)
	return nil
}

func (dag *BlockDAG) processUnprocessedBlockNodes(unprocessedBlockNodes []*blockNode) error {
	for _, node := range unprocessedBlockNodes {
		// Check to see if the block exists in the block DB. If it
		// doesn't, the database has certainly been corrupted.
		blockExists, err := dbaccess.HasBlock(dbaccess.NoTx(), node.hash)
		if err != nil {
			return errors.Wrapf(err, "HasBlock "+
				"for block %s failed: %s", node.hash, err)
		}
		if !blockExists {
			return errors.Errorf("block %s "+
				"exists in block index but not in block db", node.hash)
		}

		// Attempt to accept the block.
		block, err := fetchBlockByHash(dbaccess.NoTx(), node.hash)
		if err != nil {
			return err
		}
		isOrphan, isDelayed, err := dag.ProcessBlock(block, BFWasStored)
		if err != nil {
			log.Warnf("Block %s, which was not previously processed, "+
				"failed to be accepted to the DAG: %s", node.hash, err)
			continue
		}

		// If the block is an orphan or is delayed then it couldn't have
		// possibly been written to the block index in the first place.
		if isOrphan {
			return errors.Errorf("Block %s, which was not "+
				"previously processed, turned out to be an orphan, which is "+
				"impossible.", node.hash)
		}
		if isDelayed {
			return errors.Errorf("Block %s, which was not "+
				"previously processed, turned out to be delayed, which is "+
				"impossible.", node.hash)
		}
	}
	return nil
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
		timestamp:            header.Timestamp.UnixMilliseconds(),
		hashMerkleRoot:       header.HashMerkleRoot,
		acceptedIDMerkleRoot: header.AcceptedIDMerkleRoot,
		utxoCommitment:       header.UTXOCommitment,
	}

	node.children = newBlockSet()
	node.parents = newBlockSet()

	for _, hash := range header.ParentHashes {
		parent, ok := dag.index.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("deserializeBlockNode: Could "+
				"not find parent %s for block %s", hash, header.BlockHash())
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
		var ok bool
		node.selectedParent, ok = dag.index.LookupNode(selectedParentHash)
		if !ok {
			return nil, errors.Errorf("block %s does not exist in the DAG", selectedParentHash)
		}
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

		var ok bool
		node.blues[i], ok = dag.index.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("block %s does not exist in the DAG", selectedParentHash)
		}
	}

	bluesAnticoneSizesLen, err := wire.ReadVarInt(buffer)
	if err != nil {
		return nil, err
	}

	node.bluesAnticoneSizes = make(map[*blockNode]dagconfig.KType)
	for i := uint64(0); i < bluesAnticoneSizesLen; i++ {
		hash := &daghash.Hash{}
		if _, err := io.ReadFull(buffer, hash[:]); err != nil {
			return nil, err
		}
		bluesAnticoneSize, err := binaryserializer.Uint8(buffer)
		if err != nil {
			return nil, err
		}
		blue, ok := dag.index.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("couldn't find block with hash %s", hash)
		}
		node.bluesAnticoneSizes[blue] = dagconfig.KType(bluesAnticoneSize)
	}

	return node, nil
}

// fetchBlockByHash retrieves the raw block for the provided hash,
// deserializes it, and returns a util.Block of it.
func fetchBlockByHash(dbContext dbaccess.Context, hash *daghash.Hash) (*util.Block, error) {
	blockBytes, err := dbaccess.FetchBlock(dbContext, hash)
	if err != nil {
		return nil, err
	}
	return util.NewBlockFromBytes(blockBytes)
}

func storeBlock(dbContext *dbaccess.TxContext, block *util.Block) error {
	blockBytes, err := block.Bytes()
	if err != nil {
		return err
	}
	return dbaccess.StoreBlock(dbContext, block.Hash(), blockBytes)
}

func serializeBlockNode(node *blockNode) ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, wire.MaxBlockHeaderPayload+1))
	header := node.Header()
	err := header.Serialize(w)
	if err != nil {
		return nil, err
	}

	err = w.WriteByte(byte(node.status))
	if err != nil {
		return nil, err
	}

	// Because genesis doesn't have selected parent, it's serialized as zero hash
	selectedParentHash := &daghash.ZeroHash
	if node.selectedParent != nil {
		selectedParentHash = node.selectedParent.hash
	}
	_, err = w.Write(selectedParentHash[:])
	if err != nil {
		return nil, err
	}

	err = binaryserializer.PutUint64(w, byteOrder, node.blueScore)
	if err != nil {
		return nil, err
	}

	err = wire.WriteVarInt(w, uint64(len(node.blues)))
	if err != nil {
		return nil, err
	}

	for _, blue := range node.blues {
		_, err = w.Write(blue.hash[:])
		if err != nil {
			return nil, err
		}
	}

	err = wire.WriteVarInt(w, uint64(len(node.bluesAnticoneSizes)))
	if err != nil {
		return nil, err
	}
	for blue, blueAnticoneSize := range node.bluesAnticoneSizes {
		_, err = w.Write(blue.hash[:])
		if err != nil {
			return nil, err
		}

		err = binaryserializer.PutUint8(w, uint8(blueAnticoneSize))
		if err != nil {
			return nil, err
		}
	}
	return w.Bytes(), nil
}

// blockIndexKey generates the binary key for an entry in the block index
// bucket. The key is composed of the block blue score encoded as a big-endian
// 64-bit unsigned int followed by the 32 byte block hash.
// The blue score component is important for iteration order.
func blockIndexKey(blockHash *daghash.Hash, blueScore uint64) []byte {
	indexKey := make([]byte, daghash.HashSize+8)
	binary.BigEndian.PutUint64(indexKey[0:8], blueScore)
	copy(indexKey[8:daghash.HashSize+8], blockHash[:])
	return indexKey
}

func blockHashFromBlockIndexKey(BlockIndexKey []byte) (*daghash.Hash, error) {
	return daghash.NewHash(BlockIndexKey[8 : daghash.HashSize+8])
}

// BlockByHash returns the block from the DAG with the given hash.
//
// This function is safe for concurrent access.
func (dag *BlockDAG) BlockByHash(hash *daghash.Hash) (*util.Block, error) {
	// Lookup the block hash in block index and ensure it is in the DAG
	node, ok := dag.index.LookupNode(hash)
	if !ok {
		str := fmt.Sprintf("block %s is not in the DAG", hash)
		return nil, errNotInDAG(str)
	}

	block, err := fetchBlockByHash(dbaccess.NoTx(), node.hash)
	if err != nil {
		return nil, err
	}
	return block, err
}

// BlockHashesFrom returns a slice of blocks starting from lowHash
// ordered by blueScore. If lowHash is nil then the genesis block is used.
//
// This method MUST be called with the DAG lock held
func (dag *BlockDAG) BlockHashesFrom(lowHash *daghash.Hash, limit int) ([]*daghash.Hash, error) {
	blockHashes := make([]*daghash.Hash, 0, limit)
	if lowHash == nil {
		lowHash = dag.genesis.hash

		// If we're starting from the beginning we should include the
		// genesis hash in the result
		blockHashes = append(blockHashes, dag.genesis.hash)
	}
	if !dag.IsInDAG(lowHash) {
		return nil, errors.Errorf("block %s not found", lowHash)
	}
	blueScore, err := dag.BlueScoreByBlockHash(lowHash)
	if err != nil {
		return nil, err
	}

	key := blockIndexKey(lowHash, blueScore)
	cursor, err := dbaccess.BlockIndexCursorFrom(dbaccess.NoTx(), key)
	if dbaccess.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "block %s not in block index", lowHash)
	}
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	for cursor.Next() && len(blockHashes) < limit {
		key, err := cursor.Key()
		if err != nil {
			return nil, err
		}
		blockHash, err := blockHashFromBlockIndexKey(key.Suffix())
		if err != nil {
			return nil, err
		}
		blockHashes = append(blockHashes, blockHash)
	}

	return blockHashes, nil
}
