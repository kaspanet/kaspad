package hashserialization

import (
	"encoding/binary"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/pkg/errors"
)

var (
	// littleEndian is a convenience variable since binary.LittleEndian is
	// quite long.
	littleEndian = binary.LittleEndian

	// bigEndian is a convenience variable since binary.BigEndian is quite
	// long.
	bigEndian = binary.BigEndian
)

// errNoEncodingForType signifies that there's no encoding for the given type.
var errNoEncodingForType = errors.New("there's no encoding for this type")

// WriteElement writes the little endian representation of element to w.
func WriteElement(w io.Writer, element interface{}) error {
	// Attempt to write the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case int32:
		err := binaryserializer.PutUint32(w, littleEndian, uint32(e))
		if err != nil {
			return err
		}
		return nil

	case uint32:
		err := binaryserializer.PutUint32(w, littleEndian, e)
		if err != nil {
			return err
		}
		return nil

	case int64:
		err := binaryserializer.PutUint64(w, littleEndian, uint64(e))
		if err != nil {
			return err
		}
		return nil

	case uint64:
		err := binaryserializer.PutUint64(w, littleEndian, e)
		if err != nil {
			return err
		}
		return nil

	case uint8:
		err := binaryserializer.PutUint8(w, e)
		if err != nil {
			return err
		}
		return nil

	case bool:
		var err error
		if e {
			err = binaryserializer.PutUint8(w, 0x01)
		} else {
			err = binaryserializer.PutUint8(w, 0x00)
		}
		if err != nil {
			return err
		}
		return nil

	case externalapi.DomainHash:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case *externalapi.DomainHash:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case externalapi.DomainSubnetworkID:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case *externalapi.DomainSubnetworkID:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil
	}

	return errors.Wrapf(errNoEncodingForType, "couldn't find a way to write type %T", element)
}

// writeElements writes multiple items to w. It is equivalent to multiple
// calls to writeElement.
func writeElements(w io.Writer, elements ...interface{}) error {
	for _, element := range elements {
		err := WriteElement(w, element)
		if err != nil {
			return err
		}
	}
	return nil
}
