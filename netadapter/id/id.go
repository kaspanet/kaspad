package id

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// IDLength of array used to store the ID.
const IDLength = 16

// ID identifies a network connection
type ID struct {
	bytes []byte
}

// GenerateID generates a new ID
func GenerateID() (*ID, error) {
	idBytes := make([]byte, IDLength)
	_, err := rand.Read(idBytes)
	if err != nil {
		return nil, err
	}
	return NewID(idBytes)
}

// NewID creates an ID from the given bytes
func NewID(bytes []byte) (*ID, error) {
	if len(bytes) != IDLength {
		return nil, errors.New("invalid bytes length")
	}
	return &ID{bytes: bytes}, nil
}

// Bytes returns the serialized bytes for the ID.
func (id *ID) Bytes() []byte {
	bytesCopy := make([]byte, IDLength)
	copy(bytesCopy, id.bytes)
	return bytesCopy
}

// IsEqual returns whether id equals to other.
func (id *ID) IsEqual(other *ID) bool {
	return bytes.Equal(id.bytes, other.bytes)
}

func (id *ID) String() string {
	return hex.EncodeToString(id.bytes)
}
