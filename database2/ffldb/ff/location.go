package ff

import "github.com/pkg/errors"

// flatFileLocationSerializedSize is the size in bytes of a serialized flat
// file location. See serializeLocation for further details.
const flatFileLocationSerializedSize = 12

// flatFileLocation identifies a particular flat file location.
type flatFileLocation struct {
	fileNumber uint32
	fileOffset uint32
	dataLength uint32
}

// serializeLocation returns the serialization of the passed flat file location
// of certain data. This to later on be used for retrieval of said data.
// The serialized location format is:
//
//  [0:4]  File Number (4 bytes)
//  [4:8]  File offset (4 bytes)
//  [8:12] Data length (4 bytes)
func serializeLocation(location *flatFileLocation) []byte {
	var serializedLocation [flatFileLocationSerializedSize]byte
	byteOrder.PutUint32(serializedLocation[0:4], location.fileNumber)
	byteOrder.PutUint32(serializedLocation[4:8], location.fileOffset)
	byteOrder.PutUint32(serializedLocation[8:12], location.dataLength)
	return serializedLocation[:]
}

// deserializeLocation deserializes the passed serialized flat file location.
// See serializeLocation for further details.
func deserializeLocation(serializedLocation []byte) (*flatFileLocation, error) {
	if len(serializedLocation) != flatFileLocationSerializedSize {
		return nil, errors.Errorf("unexpected serializedLocation length: %d",
			len(serializedLocation))
	}
	location := &flatFileLocation{
		fileNumber: byteOrder.Uint32(serializedLocation[0:4]),
		fileOffset: byteOrder.Uint32(serializedLocation[4:8]),
		dataLength: byteOrder.Uint32(serializedLocation[8:12]),
	}
	return location, nil
}
