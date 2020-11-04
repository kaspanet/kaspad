package hashserialization

import (
	"bytes"
	"encoding/binary"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"io"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// outpointIndexByteOrder is the byte order for serializing the outpoint index.
// It uses big endian to ensure that when outpoint is used as database key, the
// keys will be iterated in an ascending order by the outpoint index.
var outpointIndexByteOrder = binary.BigEndian

const (
	outpointLen             = externalapi.DomainHashSize + 4
	entryMinLen             = 8 + 1 + 8 + 8
	averageScriptPubKeySize = 20
)

// ReadOnlyUTXOSetToProtoUTXOSet converts ReadOnlyUTXOSetIterator to ProtoUTXOSet
func ReadOnlyUTXOSetToProtoUTXOSet(iter model.ReadOnlyUTXOSetIterator) (*ProtoUTXOSet, error) {
	protoUTXOSet := &ProtoUTXOSet{
		Utxos: []*ProtoUTXO{},
	}

	for iter.Next() {
		outpoint, entry := iter.Get()

		serializedOutpoint := bytes.NewBuffer(make([]byte, 0, outpointLen))
		err := serializeOutpoint(serializedOutpoint, outpoint)
		if err != nil {
			return nil, err
		}

		serializedEntry := bytes.NewBuffer(make([]byte, 0, entryMinLen+averageScriptPubKeySize))
		err = serializeUTXOEntry(serializedEntry, entry)
		if err != nil {
			return nil, err
		}

		protoUTXOSet.Utxos = append(protoUTXOSet.Utxos, &ProtoUTXO{
			Entry:    serializedEntry.Bytes(),
			Outpoint: serializedOutpoint.Bytes(),
		})
	}
	return protoUTXOSet, nil
}

// SerializeUTXO returns the byte-slice representation for given UTXOEntry
func SerializeUTXO(entry *externalapi.UTXOEntry, outpoint *externalapi.DomainOutpoint) ([]byte, error) {
	w := &bytes.Buffer{}

	err := serializeOutpoint(w, outpoint)
	if err != nil {
		return nil, err
	}

	err = serializeUTXOEntry(w, entry)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

func serializeOutpoint(w io.Writer, outpoint *externalapi.DomainOutpoint) error {
	_, err := w.Write(outpoint.TransactionID[:])
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

func serializeUTXOEntry(w io.Writer, entry *externalapi.UTXOEntry) error {
	err := writeElements(w, entry.BlockBlueScore, entry.Amount, entry.IsCoinbase)
	if err != nil {
		return err
	}

	count := uint64(len(entry.ScriptPublicKey))
	err = WriteElement(w, count)
	if err != nil {
		return err
	}

	_, err = w.Write(entry.ScriptPublicKey)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
