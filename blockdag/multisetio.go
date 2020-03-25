package blockdag

import (
	"encoding/binary"
	"github.com/kaspanet/kaspad/ecc"
	"io"
	"math/big"
)

const multisetPointSize = 32

// serializeMultiset serializes an ECMH multiset. The serialization
// is done by taking the (x,y) coordinnates of the multiset point and
// padding each one of them with 32 byte (it'll be 32 byte in most
// cases anyway except one of the coordinates is zero) and writing
// them one after the other.
func serializeMultiset(w io.Writer, ms *ecc.Multiset) error {
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

// deserializeMultiset deserializes an EMCH multiset.
// See serializeMultiset for more details.
func deserializeMultiset(r io.Reader) (*ecc.Multiset, error) {
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
	return ecc.NewMultisetFromPoint(ecc.S256(), &x, &y), nil
}
