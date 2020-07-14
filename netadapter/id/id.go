package id

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// IDLength of array used to store the ID.
const IDLength = 16

// ID identifies a network connection
type ID struct {
	bytes []byte
}

// GenerateID generates a new ID
func GenerateID() (*ID, error) {
	id := new(ID)
	err := id.Deserialize(rand.Reader)
	if err != nil {
		return nil, err
	}
	return id, nil
}

// IsEqual returns whether id equals to other.
func (id *ID) IsEqual(other *ID) bool {
	return bytes.Equal(id.bytes, other.bytes)
}

func (id *ID) String() string {
	return hex.EncodeToString(id.bytes)
}

// Deserialize decodes a block from r into the receiver.
func (id *ID) Deserialize(r io.Reader) error {
	id.bytes = make([]byte, IDLength)
	_, err := io.ReadFull(r, id.bytes)
	return err
}

// Serialize serializes the receiver into the given writer.
func (id *ID) Serialize(w io.Writer) error {
	_, err := w.Write(id.bytes)
	return err
}

// FromBytes returns an ID deserialized from the given byte slice.
func FromBytes(serializedID []byte) *ID {
	r := bytes.NewReader(serializedID)
	newID := new(ID)
	err := newID.Deserialize(r)
	if err != nil {
		panic(err)
	}
	return newID
}
