package utxo

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/daghash"
	"github.com/pkg/errors"
)

var (
	// byteOrder is the preferred byte order used for serializing numeric
	// fields for storage in the database.
	byteOrder = binary.LittleEndian
)

var outpointSerializeSize = daghash.TxIDSize + 4

// UpdateUTXOSet updates the UTXO set in the database based on the provided
// UTXO Diff.
func UpdateUTXOSet(dbContext dbaccess.Context, virtualUTXODiff *Diff) error {
	outpointBuff := bytes.NewBuffer(make([]byte, outpointSerializeSize))
	for outpoint := range virtualUTXODiff.ToRemove {
		outpointBuff.Reset()
		err := SerializeOutpoint(outpointBuff, &outpoint)
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

	for outpoint, entry := range virtualUTXODiff.ToAdd {
		utxoEntryBuff.Reset()
		outpointBuff.Reset()
		// Serialize and store the UTXO entry.
		err := SerializeUTXOEntry(utxoEntryBuff, entry)
		if err != nil {
			return err
		}
		serializedEntry := utxoEntryBuff.Bytes()

		err = SerializeOutpoint(outpointBuff, &outpoint)
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

// outpointIndexByteOrder is the byte order for serializing the outpoint index.
// It uses big endian to ensure that when outpoint is used as database key, the
// keys will be iterated in an ascending order by the outpoint index.
var outpointIndexByteOrder = binary.BigEndian

func SerializeOutpoint(w io.Writer, outpoint *appmessage.Outpoint) error {
	_, err := w.Write(outpoint.TxID[:])
	if err != nil {
		return err
	}

	var buf [4]byte
	outpointIndexByteOrder.PutUint32(buf[:], outpoint.Index)
	_, err = w.Write(buf[:])
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeserializeOutpoint decodes an outpoint from the passed serialized byte
// slice into a new appmessage.Outpoint using a format that is suitable for long-
// term storage. This format is described in detail above.
func DeserializeOutpoint(r io.Reader) (*appmessage.Outpoint, error) {
	outpoint := &appmessage.Outpoint{}
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

// serializeBlockUTXODiffData serializes Diff data in the following format:
// 	Name         | Data type | Description
//  ------------ | --------- | -----------
// 	hasDiffChild | Boolean   | Indicates if a Diff child exist
//  diffChild    | Hash      | The diffChild's hash. Empty if hasDiffChild is true.
//  Diff		 | Diff  | The Diff data's Diff
func serializeBlockUTXODiffData(w io.Writer, diffData *blockUTXODiffData) error {
	hasDiffChild := diffData.diffChild != nil
	err := appmessage.WriteElement(w, hasDiffChild)
	if err != nil {
		return err
	}
	if hasDiffChild {
		err := appmessage.WriteElement(w, diffData.diffChild.Hash)
		if err != nil {
			return err
		}
	}

	err = serializeUTXODiff(w, diffData.Diff)
	if err != nil {
		return err
	}

	return nil
}

func (diffStore *DiffStore) deserializeBlockUTXODiffData(serializedDiffData []byte) (*blockUTXODiffData, error) {
	diffData := &blockUTXODiffData{}
	r := bytes.NewBuffer(serializedDiffData)

	var hasDiffChild bool
	err := appmessage.ReadElement(r, &hasDiffChild)
	if err != nil {
		return nil, err
	}

	if hasDiffChild {
		hash := &daghash.Hash{}
		err := appmessage.ReadElement(r, hash)
		if err != nil {
			return nil, err
		}

		// var ok bool
		// diffData.diffChild, ok = diffStore.dag.index.LookupNode(hash)
		// if !ok {
		// 	return nil, errors.Errorf("block %s does not exist in the DAG", hash)
		// }
	}

	diffData.Diff, err = deserializeUTXODiff(r)
	if err != nil {
		return nil, err
	}

	return diffData, nil
}

func deserializeUTXODiff(r io.Reader) (*Diff, error) {
	diff := &Diff{}

	var err error
	diff.ToAdd, err = deserializeUTXOCollection(r)
	if err != nil {
		return nil, err
	}

	diff.ToRemove, err = deserializeUTXOCollection(r)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func deserializeUTXOCollection(r io.Reader) (utxoCollection, error) {
	count, err := appmessage.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := utxoCollection{}
	for i := uint64(0); i < count; i++ {
		utxoEntry, outpoint, err := deserializeUTXO(r)
		if err != nil {
			return nil, err
		}
		collection.Add(*outpoint, utxoEntry)
	}
	return collection, nil
}

func deserializeUTXO(r io.Reader) (*Entry, *appmessage.Outpoint, error) {
	outpoint, err := DeserializeOutpoint(r)
	if err != nil {
		return nil, nil, err
	}

	utxoEntry, err := DeserializeUTXOEntry(r)
	if err != nil {
		return nil, nil, err
	}
	return utxoEntry, outpoint, nil
}

// serializeUTXODiff serializes Diff by serializing
// Diff.ToAdd, Diff.ToRemove and Diff.Multiset one after the other.
func serializeUTXODiff(w io.Writer, diff *Diff) error {
	err := serializeUTXOCollection(w, diff.ToAdd)
	if err != nil {
		return err
	}

	err = serializeUTXOCollection(w, diff.ToRemove)
	if err != nil {
		return err
	}

	return nil
}

// serializeUTXOCollection serializes utxoCollection by iterating over
// the utxo entries and serializing them and their corresponding outpoint
// prefixed by a varint that indicates their size.
func serializeUTXOCollection(w io.Writer, collection utxoCollection) error {
	err := appmessage.WriteVarInt(w, uint64(len(collection)))
	if err != nil {
		return err
	}
	for outpoint, utxoEntry := range collection {
		err := SerializeUTXO(w, utxoEntry, &outpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

// SerializeUTXO serializes a utxo entry-outpoint pair
func SerializeUTXO(w io.Writer, entry *Entry, outpoint *appmessage.Outpoint) error {
	err := SerializeOutpoint(w, outpoint)
	if err != nil {
		return err
	}

	err = SerializeUTXOEntry(w, entry)
	if err != nil {
		return err
	}
	return nil
}

// p2pkhUTXOEntrySerializeSize is the serialized size for a P2PKH UTXO entry.
// 8 bytes (header code) + 8 bytes (amount) + varint for script pub key length of 25 (for P2PKH) + 25 bytes for P2PKH script.
var p2pkhUTXOEntrySerializeSize = 8 + 8 + appmessage.VarIntSerializeSize(25) + 25

// SerializeUTXOEntry encodes the entry to the given io.Writer and use compression if useCompression is true.
// The compression format is described in detail above.
func SerializeUTXOEntry(w io.Writer, entry *Entry) error {
	buf := [8 + 1 + 8]byte{}
	// Encode the blueScore.
	binary.LittleEndian.PutUint64(buf[:8], entry.blockBlueScore)

	// Encode the packedFlags.
	buf[8] = uint8(entry.packedFlags)

	binary.LittleEndian.PutUint64(buf[9:], entry.Amount())

	_, err := w.Write(buf[:])
	if err != nil {
		return errors.WithStack(err)
	}

	err = appmessage.WriteVarInt(w, uint64(len(entry.ScriptPubKey())))
	if err != nil {
		return err
	}

	_, err = w.Write(entry.ScriptPubKey())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// DeserializeUTXOEntry decodes a UTXO entry from the passed reader
// into a new Entry. If isCompressed is used it will decompress
// the entry according to the format that is described in detail
// above.
func DeserializeUTXOEntry(r io.Reader) (*Entry, error) {
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

	entry := &Entry{
		blockBlueScore: blockBlueScore,
		packedFlags:    txoFlags(packedFlags),
	}

	entry.amount, err = binaryserializer.Uint64(r, byteOrder)
	if err != nil {
		return nil, err
	}

	scriptPubKeyLen, err := appmessage.ReadVarInt(r)
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
