package binaryserialization

import (
	"encoding/binary"
	"github.com/pkg/errors"
)

const uint64Length = 8

// SerializeUint64 serializes a uint64
func SerializeUint64(value uint64) []byte {
	var keyBytes [uint64Length]byte
	binary.LittleEndian.PutUint64(keyBytes[:], value)
	return keyBytes[:]
}

// DeserializeUint64 deserializes bytes to uint64
func DeserializeUint64(valueBytes []byte) (uint64, error) {
	if len(valueBytes) != uint64Length {
		return 0, errors.Errorf("the given value is %d bytes so it cannot be deserialized into uint64",
			len(valueBytes))
	}
	return binary.LittleEndian.Uint64(valueBytes), nil
}
