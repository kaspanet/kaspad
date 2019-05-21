package blockdag

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/big"

	"github.com/daglabs/btcd/btcec"
	"github.com/daglabs/btcd/util/daghash"
	"github.com/daglabs/btcd/wire"
)

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

func (diffStore *utxoDiffStore) deserializeBlockUTXODiffData(serializedDiffDataBytes []byte) (*blockUTXODiffData, error) {
	diffData := &blockUTXODiffData{}
	serializedDiffData := bytes.NewBuffer(serializedDiffDataBytes)

	var hasDiffChild bool
	err := wire.ReadElement(serializedDiffData, &hasDiffChild)
	if err != nil {
		return nil, err
	}

	if hasDiffChild {
		hash := &daghash.Hash{}
		err := wire.ReadElement(serializedDiffData, hash)
		if err != nil {
			return nil, err
		}
		diffData.diffChild = diffStore.dag.index.LookupNode(hash)
	}

	diffData.diff = &UTXODiff{}

	diffData.diff.toAdd, err = deserializeDiffEntries(serializedDiffData)
	if err != nil {
		return nil, err
	}

	diffData.diff.toRemove, err = deserializeDiffEntries(serializedDiffData)
	if err != nil {
		return nil, err
	}

	diffData.diff.diffMultiset, err = deserializeMultiset(serializedDiffData)
	if err != nil {
		return nil, err
	}

	return diffData, nil
}

func deserializeDiffEntries(r io.Reader) (utxoCollection, error) {
	count, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := utxoCollection{}
	for i := uint64(0); i < count; i++ {
		outPointSize, err := wire.ReadVarInt(r)
		if err != nil {
			return nil, err
		}

		serializedOutPoint := make([]byte, outPointSize)
		err = binary.Read(r, byteOrder, serializedOutPoint)
		if err != nil {
			return nil, err
		}
		outPoint, err := deserializeOutPoint(serializedOutPoint)
		if err != nil {
			return nil, err
		}

		utxoEntrySize, err := wire.ReadVarInt(r)
		if err != nil {
			return nil, err
		}
		serializedEntry := make([]byte, utxoEntrySize)
		err = binary.Read(r, byteOrder, serializedEntry)
		if err != nil {
			return nil, err
		}
		utxoEntry, err := deserializeUTXOEntry(serializedEntry)
		if err != nil {
			return nil, err
		}
		collection.add(*outPoint, utxoEntry)
	}
	return collection, nil
}

// deserializeMultiset deserializes an EMCH multiset.
// See serializeMultiset for more details.
func deserializeMultiset(r io.Reader) (*btcec.Multiset, error) {
	xBytes := make([]byte, multisetPointSize)
	yBytes := make([]byte, multisetPointSize)
	err := binary.Read(r, byteOrder, xBytes)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, byteOrder, yBytes)
	if err != nil {
		return nil, err
	}
	var x, y big.Int
	x.SetBytes(xBytes)
	y.SetBytes(yBytes)
	return btcec.NewMultisetFromPoint(btcec.S256(), &x, &y), nil
}

// serializeBlockUTXODiffData serializes diff data in the following format:
// 	Name         | Data type | Description
//	------------ | --------- | -----------
// 	hasDiffChild | Boolean   | Indicates if a diff child exist
//  diffChild    | Hash      | The diffChild's hash. Empty if hasDiffChild is true.
//  diff		 | UTXODiff  | The diff data's diff
func serializeBlockUTXODiffData(diffData *blockUTXODiffData) ([]byte, error) {
	w := &bytes.Buffer{}
	hasDiffChild := diffData.diffChild != nil
	err := wire.WriteElement(w, hasDiffChild)
	if err != nil {
		return nil, err
	}
	if hasDiffChild {
		err := wire.WriteElement(w, diffData.diffChild.hash)
		if err != nil {
			return nil, err
		}
	}

	err = serializeUTXODiff(w, diffData.diff)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// serializeUTXODiff serializes UTXODiff by serializing
// UTXODiff.toAdd, UTXODiff.toRemove and UTXODiff.Multiset one after the other.
func serializeUTXODiff(w io.Writer, diff *UTXODiff) error {
	err := serializeUTXOCollection(w, diff.toAdd)
	if err != nil {
		return err
	}

	err = serializeUTXOCollection(w, diff.toRemove)
	if err != nil {
		return err
	}
	err = serializeMultiset(w, diff.diffMultiset)
	if err != nil {
		return err
	}
	return nil
}

// serializeUTXOCollection serializes utxoCollection by iterating over
// the utxo entries and serializing them and their corresponding outpoint
// prefixed by a varint that indicates their size.
func serializeUTXOCollection(w io.Writer, collection utxoCollection) error {
	err := wire.WriteVarInt(w, uint64(len(collection)))
	if err != nil {
		return err
	}
	for outPoint, utxoEntry := range collection {
		err := serializeUTXO(w, utxoEntry, &outPoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// serializeMultiset serializes an ECMH multiset. The serialization
// is done by taking the (x,y) coordinnates of the multiset point and
// padding each one of them with 32 byte (it'll be 32 byte in most
// cases anyway except one of the coordinates is zero) and writing
// them one after the other.
func serializeMultiset(w io.Writer, ms *btcec.Multiset) error {
	x, y := ms.Point()
	xBytes := make([]byte, multisetPointSize)
	copy(xBytes, x.Bytes())
	yBytes := make([]byte, multisetPointSize)
	copy(yBytes, y.Bytes())

	err := binary.Write(w, byteOrder, xBytes)
	if err != nil {
		return err
	}
	err = binary.Write(w, byteOrder, yBytes)
	if err != nil {
		return err
	}
	return nil
}

// serializeUTXO serializes a utxo entry-outpoint pair
func serializeUTXO(w io.Writer, entry *UTXOEntry, outPoint *wire.OutPoint) error {
	serializedOutPoint := *outpointKey(*outPoint)
	err := wire.WriteVarInt(w, uint64(len(serializedOutPoint)))
	if err != nil {
		return err
	}

	err = binary.Write(w, byteOrder, serializedOutPoint)
	if err != nil {
		return err
	}

	serializedUTXOEntry := serializeUTXOEntry(entry)
	err = wire.WriteVarInt(w, uint64(len(serializedUTXOEntry)))
	if err != nil {
		return err
	}
	err = binary.Write(w, byteOrder, serializedUTXOEntry)
	if err != nil {
		return err
	}
	return nil
}

// serializeUTXOEntry returns the entry serialized to a format that is suitable
// for long-term storage.  The format is described in detail above.
func serializeUTXOEntry(entry *UTXOEntry) []byte {

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

	return serialized
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
