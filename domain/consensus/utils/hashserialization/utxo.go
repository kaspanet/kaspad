package hashserialization

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// outpointIndexByteOrder is the byte order for serializing the outpoint index.
// It uses big endian to ensure that when outpoint is used as database key, the
// keys will be iterated in an ascending order by the outpoint index.
var outpointIndexByteOrder = binary.BigEndian

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
	_, err := w.Write(outpoint.ID[:])
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
	buf := [8 + 1 + 8]byte{}
	// Encode the blueScore.
	binary.LittleEndian.PutUint64(buf[:8], entry.BlockBlueScore)

	buf[8] = serializeUTXOEntryFlags(entry)

	binary.LittleEndian.PutUint64(buf[9:], entry.Amount)

	_, err := w.Write(buf[:])
	if err != nil {
		return errors.WithStack(err)
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

func serializeUTXOEntryFlags(entry *externalapi.UTXOEntry) uint8 {
	var serializedFlags uint8 = 0

	if entry.IsCoinbase {
		serializedFlags |= 1
	}

	return serializedFlags
}
