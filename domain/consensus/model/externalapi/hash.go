package externalapi

import (
	"encoding/hex"

	"github.com/pkg/errors"
)

// DomainHashSize of array used to store hashes.
const DomainHashSize = 32

// DomainHash is the domain representation of a Hash
type DomainHash struct {
	hashArray [DomainHashSize]byte
}

func NewDomainHashFromByteArray(hashBytes *[DomainHashSize]byte) *DomainHash {
	return &DomainHash{
		hashArray: *hashBytes,
	}
}

func NewDomainHashFromByteSlice(hashBytes []byte) (*DomainHash, error) {
	if len(hashBytes) != DomainHashSize {
		return nil, errors.Errorf("invalid hash size. Want: %d, got: %d",
			DomainHashSize, len(hashBytes))
	}
	domainHash := DomainHash{
		hashArray: [DomainHashSize]byte{},
	}
	copy(domainHash.hashArray[:], hashBytes)
	return &domainHash, nil
}

func NewDomainHashFromString(hashString string) (*DomainHash, error) {
	expectedLength := DomainHashSize * 2
	// Return error if hash string is too long.
	if len(hashString) != expectedLength {
		return nil, errors.Errorf("hash string length is %d, while it should be be %d",
			len(hashString), expectedLength)
	}

	hashBytes, err := hex.DecodeString(hashString)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return NewDomainHashFromByteSlice(hashBytes)
}

// String returns the Hash as the hexadecimal string of the hash.
func (hash DomainHash) String() string {
	return hex.EncodeToString(hash.hashArray[:])
}

// BytesArray returns the bytes in this hash represented as a bytes array.
// The hash bytes are cloned, therefore it is safe to modify the resulting array.
func (hash *DomainHash) BytesArray() *[DomainHashSize]byte {
	arrayClone := hash.hashArray
	return &arrayClone
}

// BytesArray returns the bytes in this hash represented as a bytes slice.
// The hash bytes are cloned, therefore it is safe to modify the resulting slice.
func (hash *DomainHash) BytesSlice() []byte {
	return hash.BytesArray()[:]
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ DomainHash = DomainHash{hashArray: [DomainHashSize]byte{}}

// Equal returns whether hash equals to other
func (hash *DomainHash) Equal(other *DomainHash) bool {
	if hash == nil || other == nil {
		return hash == other
	}

	return hash.hashArray == other.hashArray
}

// CloneHashes returns a clone of the given hashes slice.
// Note: since DomainHash is a read-only type, the clone is shallow
func CloneHashes(hashes []*DomainHash) []*DomainHash {
	clone := make([]*DomainHash, len(hashes))
	copy(clone, hashes)
	return clone
}

// HashesEqual returns whether the given hash slices are equal.
func HashesEqual(a, b []*DomainHash) bool {
	if len(a) != len(b) {
		return false
	}

	for i, hash := range a {
		if !hash.Equal(b[i]) {
			return false
		}
	}
	return true
}
