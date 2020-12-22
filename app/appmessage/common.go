// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package appmessage

import (
	"fmt"
	"io"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/id"
	"github.com/kaspanet/kaspad/util/binaryserializer"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/pkg/errors"
)

// MaxVarIntPayload is the maximum payload size for a variable length integer.
const MaxVarIntPayload = 9

// MaxInvPerMsg is the maximum number of inventory vectors that can be in any type of kaspa inv message.
const MaxInvPerMsg = 1 << 17

// errNonCanonicalVarInt is the common format string used for non-canonically
// encoded variable length integer errors.
var errNonCanonicalVarInt = "non-canonical varint %x - discriminant %x must " +
	"encode a value greater than %x"

// errNoEncodingForType signifies that there's no encoding for the given type.
var errNoEncodingForType = errors.New("there's no encoding for this type")

// int64Time represents a unix timestamp with milliseconds precision encoded with
// an int64. It is used as a way to signal the readElement function how to decode
// a timestamp into a Go mstime.Time since it is otherwise ambiguous.
type int64Time mstime.Time

// ReadElement reads the next sequence of bytes from r using little endian
// depending on the concrete type of element pointed to.
func ReadElement(r io.Reader, element interface{}) error {
	// Attempt to read the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case *int32:
		rv, err := binaryserializer.Uint32(r)
		if err != nil {
			return err
		}
		*e = int32(rv)
		return nil

	case *uint32:
		rv, err := binaryserializer.Uint32(r)
		if err != nil {
			return err
		}
		*e = rv
		return nil

	case *int64:
		rv, err := binaryserializer.Uint64(r)
		if err != nil {
			return err
		}
		*e = int64(rv)
		return nil

	case *uint64:
		rv, err := binaryserializer.Uint64(r)
		if err != nil {
			return err
		}
		*e = rv
		return nil

	case *uint8:
		rv, err := binaryserializer.Uint8(r)
		if err != nil {
			return err
		}
		*e = rv
		return nil

	case *bool:
		rv, err := binaryserializer.Uint8(r)
		if err != nil {
			return err
		}
		if rv == 0x00 {
			*e = false
		} else {
			*e = true
		}
		return nil

	// Unix timestamp encoded as an int64.
	case *int64Time:
		rv, err := binaryserializer.Uint64(r)
		if err != nil {
			return err
		}
		*e = int64Time(mstime.UnixMilliseconds(int64(rv)))
		return nil

	// Message header checksum.
	case *[4]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	// Message header command.
	case *MessageCommand:
		rv, err := binaryserializer.Uint32(r)
		if err != nil {
			return err
		}
		*e = MessageCommand(rv)
		return nil

	// IP address.
	case *[16]byte:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	case *externalapi.DomainHash:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	case *id.ID:
		return e.Deserialize(r)

	case *externalapi.DomainSubnetworkID:
		_, err := io.ReadFull(r, e[:])
		if err != nil {
			return err
		}
		return nil

	case *ServiceFlag:
		rv, err := binaryserializer.Uint64(r)
		if err != nil {
			return err
		}
		*e = ServiceFlag(rv)
		return nil

	case *KaspaNet:
		rv, err := binaryserializer.Uint32(r)
		if err != nil {
			return err
		}
		*e = KaspaNet(rv)
		return nil
	}

	return errors.Wrapf(errNoEncodingForType, "couldn't find a way to read type %T", element)
}

// readElements reads multiple items from r. It is equivalent to multiple
// calls to readElement.
func readElements(r io.Reader, elements ...interface{}) error {
	for _, element := range elements {
		err := ReadElement(r, element)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteElement writes the little endian representation of element to w.
func WriteElement(w io.Writer, element interface{}) error {
	// Attempt to write the element based on the concrete type via fast
	// type assertions first.
	switch e := element.(type) {
	case int32:
		err := binaryserializer.PutUint32(w, uint32(e))
		if err != nil {
			return err
		}
		return nil

	case uint32:
		err := binaryserializer.PutUint32(w, e)
		if err != nil {
			return err
		}
		return nil

	case int64:
		err := binaryserializer.PutUint64(w, uint64(e))
		if err != nil {
			return err
		}
		return nil

	case uint64:
		err := binaryserializer.PutUint64(w, e)
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

	// Message header checksum.
	case [4]byte:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	// Message header command.
	case MessageCommand:
		err := binaryserializer.PutUint32(w, uint32(e))
		if err != nil {
			return err
		}
		return nil

	// IP address.
	case [16]byte:
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

	case *id.ID:
		return e.Serialize(w)

	case *externalapi.DomainSubnetworkID:
		_, err := w.Write(e[:])
		if err != nil {
			return err
		}
		return nil

	case ServiceFlag:
		err := binaryserializer.PutUint64(w, uint64(e))
		if err != nil {
			return err
		}
		return nil

	case KaspaNet:
		err := binaryserializer.PutUint32(w, uint32(e))
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

// ReadVarInt reads a variable length integer from r and returns it as a uint64.
func ReadVarInt(r io.Reader) (uint64, error) {
	discriminant, err := binaryserializer.Uint8(r)
	if err != nil {
		return 0, err
	}

	var rv uint64
	switch discriminant {
	case 0xff:
		sv, err := binaryserializer.Uint64(r)
		if err != nil {
			return 0, err
		}
		rv = sv

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x100000000)
		if rv < min {
			return 0, messageError("readVarInt", fmt.Sprintf(
				errNonCanonicalVarInt, rv, discriminant, min))
		}

	case 0xfe:
		sv, err := binaryserializer.Uint32(r)
		if err != nil {
			return 0, err
		}
		rv = uint64(sv)

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0x10000)
		if rv < min {
			return 0, messageError("readVarInt", fmt.Sprintf(
				errNonCanonicalVarInt, rv, discriminant, min))
		}

	case 0xfd:
		sv, err := binaryserializer.Uint16(r)
		if err != nil {
			return 0, err
		}
		rv = uint64(sv)

		// The encoding is not canonical if the value could have been
		// encoded using fewer bytes.
		min := uint64(0xfd)
		if rv < min {
			return 0, messageError("readVarInt", fmt.Sprintf(
				errNonCanonicalVarInt, rv, discriminant, min))
		}

	default:
		rv = uint64(discriminant)
	}

	return rv, nil
}
