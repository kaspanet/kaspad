package addressindex

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/pkg/errors"
	"io"
)

// outpointIndexByteOrder is the byte order for serializing the outpoint index.
// It uses big endian to ensure that when outpoint is used as database key, the
// keys will be iterated in an ascending order by the outpoint index.
var outpointIndexByteOrder = binary.BigEndian

func DeserializeOutpointCollection(r io.Reader) (OutpointCollection, error) {
	count, err := appmessage.ReadVarInt(r)
	if err != nil {
		return nil, err
	}
	collection := OutpointCollection{}
	for i := uint64(0); i < count; i++ {
		outpoint, err := DeserializeOutpoint(r)
		if err != nil {
			return nil, err
		}
		collection.Add(*outpoint)
	}
	return collection, nil
}

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

func SerializeOutpointCollection(w io.Writer, collection OutpointCollection) error {
	err := appmessage.WriteVarInt(w, uint64(len(collection)))
	if err != nil {
		return err
	}
	for outpoint, _ := range collection {
		err := SerializeOutpoint(w, &outpoint)
		if err != nil {
			return err
		}
	}
	return nil
}
