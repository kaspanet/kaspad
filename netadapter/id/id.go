package id

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

const idLength = 16

// ID identifies a network connection
type ID struct {
	bytes []byte
}

// GenerateID generates a new ID
func GenerateID() (*ID, error) {
	bytes := make([]byte, idLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return NewID(bytes)
}

// NewID creates an ID from the given bytes
func NewID(bytes []byte) (*ID, error) {
	if len(bytes) != idLength {
		return nil, errors.New("invalid bytes length")
	}
	return &ID{bytes: bytes}, nil
}

func (id *ID) String() string {
	return hex.EncodeToString(id.bytes)
}
