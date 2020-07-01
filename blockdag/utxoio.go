package blockdag

import (
	"bytes"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/kaspanet/kaspad/wire"
	"github.com/pkg/errors"
	"io"
)

// serializeBlockUTXODiffData serializes diff data in the following format:
// 	Name         | Data type | Description
//  ------------ | --------- | -----------
// 	hasDiffChild | Boolean   | Indicates if a diff child exist
//  diffChild    | Hash      | The diffChild's hash. Empty if hasDiffChild is true.
//  diff		 | UTXODiff  | The diff data's diff
func serializeBlockUTXODiffData(w io.Writer, diffData *blockUTXODiffData) error {
	hasDiffChild := diffData.diffChild != nil
	err := wire.WriteElement(w, hasDiffChild)
	if err != nil {
		return err
	}
	if hasDiffChild {
		err := wire.WriteElement(w, diffData.diffChild.hash)
		if err != nil {
			return err
		}
	}

	err = serializeUTXODiff(w, diffData.diff)
	if err != nil {
		return err
	}

	return nil
}

func (diffStore *utxoDiffStore) deserializeBlockUTXODiffData(serializedDiffData []byte) (*blockUTXODiffData, error) {
	diffData := &blockUTXODiffData{}
	r := bytes.NewBuffer(serializedDiffData)

	var hasDiffChild bool
	err := wire.ReadElement(r, &hasDiffChild)
	if err != nil {
		return nil, err
	}

	if hasDiffChild {
		hash := &daghash.Hash{}
		err := wire.ReadElement(r, hash)
		if err != nil {
			return nil, err
		}

		var ok bool
		diffData.diffChild, ok = diffStore.dag.index.LookupNode(hash)
		if !ok {
			return nil, errors.Errorf("block %s does not exist in the DAG", hash)
		}
	}

	diffData.diff, err = deserializeUTXODiff(r)
	if err != nil {
		return nil, err
	}

	return diffData, nil
}

func deserializeUTXODiff(r io.Reader) (*UTXODiff, error) {
	diff := &UTXODiff{}

	var err error
	diff.toAdd, err = deserializeUTXOCollection(r)
	if err != nil {
		return nil, err
	}

	diff.toRemove, err = deserializeUTXOCollection(r)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func deserializeUTXOCollection(r io.Reader) (utxoCollection, error) {
	count, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := utxoCollection{}
	for i := uint64(0); i < count; i++ {
		utxoEntry, outpoint, err := deserializeUTXO(r)
		if err != nil {
			return nil, err
		}
		collection.add(*outpoint, utxoEntry)
	}
	return collection, nil
}

func deserializeUTXO(r io.Reader) (*UTXOEntry, *wire.Outpoint, error) {
	outpoint, err := deserializeOutpoint(r)
	if err != nil {
		return nil, nil, err
	}

	utxoEntry, err := deserializeUTXOEntry(r)
	if err != nil {
		return nil, nil, err
	}
	return utxoEntry, outpoint, nil
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
	for outpoint, utxoEntry := range collection {
		err := serializeUTXO(w, utxoEntry, &outpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// serializeUTXO serializes a utxo entry-outpoint pair
func serializeUTXO(w io.Writer, entry *UTXOEntry, outpoint *wire.Outpoint) error {
	err := serializeOutpoint(w, outpoint)
	if err != nil {
		return err
	}

	err = serializeUTXOEntry(w, entry)
	if err != nil {
		return err
	}
	return nil
}

// p2pkhUTXOEntrySerializeSize is the serialized size for a P2PKH UTXO entry.
// 8 bytes (header code) + 8 bytes (amount) + varint for script pub key length of 25 (for P2PKH) + 25 bytes for P2PKH script.
var p2pkhUTXOEntrySerializeSize = 8 + 8 + wire.VarIntSerializeSize(25) + 25

// serializeUTXOEntry encodes the entry to the given io.Writer and use compression if useCompression is true.
// The compression format is described in detail above.
func serializeUTXOEntry(w io.Writer, entry *UTXOEntry) error {
	// Encode the blueScore.
	err := binaryserializer.PutUint64(w, byteOrder, entry.blockBlueScore)
	if err != nil {
		return err
	}

	// Encode the packedFlags.
	err = binaryserializer.PutUint8(w, uint8(entry.packedFlags))
	if err != nil {
		return err
	}

	err = binaryserializer.PutUint64(w, byteOrder, entry.Amount())
	if err != nil {
		return err
	}

	err = wire.WriteVarInt(w, uint64(len(entry.ScriptPubKey())))
	if err != nil {
		return err
	}

	_, err = w.Write(entry.ScriptPubKey())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// deserializeUTXOEntry decodes a UTXO entry from the passed reader
// into a new UTXOEntry. If isCompressed is used it will decompress
// the entry according to the format that is described in detail
// above.
func deserializeUTXOEntry(r io.Reader) (*UTXOEntry, error) {
	// Deserialize the blueScore.
	blockBlueScore, err := binaryserializer.Uint64(r, byteOrder)
	if err != nil {
		return nil, err
	}

	// Decode the packedFlags.
	packedFlags, err := binaryserializer.Uint8(r)
	if err != nil {
		return nil, err
	}

	entry := &UTXOEntry{
		blockBlueScore: blockBlueScore,
		packedFlags:    txoFlags(packedFlags),
	}

	entry.amount, err = binaryserializer.Uint64(r, byteOrder)
	if err != nil {
		return nil, err
	}

	scriptPubKeyLen, err := wire.ReadVarInt(r)
	if err != nil {
		return nil, err
	}

	entry.scriptPubKey = make([]byte, scriptPubKeyLen)
	_, err = r.Read(entry.scriptPubKey)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return entry, nil
}
