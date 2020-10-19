package multiset

import (
	"encoding/binary"
	"github.com/kaspanet/go-secp256k1"
	"io"
)

const multisetPointSize = 32

var (
	// byteOrder is the preferred byte order used for serializing numeric
	// fields for storage in the database.
	byteOrder = binary.LittleEndian
)

// serializeMultiset serializes an ECMH multiset.
func serializeMultiset(w io.Writer, ms *secp256k1.MultiSet) error {
	serialized := ms.Serialize()
	err := binary.Write(w, byteOrder, serialized)
	if err != nil {
		return err
	}
	return nil
}

// deserializeMultiset deserializes an EMCH multiset.
func deserializeMultiset(r io.Reader) (*secp256k1.MultiSet, error) {
	serialized := &secp256k1.SerializedMultiSet{}
	err := binary.Read(r, byteOrder, serialized[:])
	if err != nil {
		return nil, err
	}
	return secp256k1.DeserializeMultiSet(serialized)
}
