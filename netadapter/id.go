package netadapter

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
	return CreateID(bytes)
}

// CreateID creates an ID from the given bytes
func CreateID(bytes []byte) (*ID, error) {
	if len(bytes) != idLength {
		return nil, errors.New("invalid bytes length")
	}
	return &ID{bytes: bytes}, nil
}

func (id *ID) String() string {
	return hex.EncodeToString(id.bytes)
}
