package externalapi

import "encoding/hex"

// DomainHashSize of array used to store hashes.
const DomainHashSize = 32

// DomainHash is the domain representation of a Hash
type DomainHash struct {
	hashArray *[DomainHashSize]byte
}

// String returns the Hash as the hexadecimal string of the hash.
func (hash DomainHash) String() string {
	return hex.EncodeToString(hash.hashArray[:])
}

// BytesArray returns the bytes in this hash represented as a bytes array.
// The hash bytes are cloned, therefore it is safe to modify the resulting array.
func (hash *DomainHash) BytesArray() *[DomainHashSize]byte {
	arrayClone := *hash.hashArray
	return &arrayClone
}

// BytesArray returns the bytes in this hash represented as a bytes slice.
// The hash bytes are cloned, therefore it is safe to modify the resulting slice.
func (hash *DomainHash) BytesSlice() []byte {
	return hash.BytesArray()[:]
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal and Clone accordingly.
var _ DomainHash = DomainHash{hashArray: &[DomainHashSize]byte{}}

// Equal returns whether hash equals to other
func (hash *DomainHash) Equal(other *DomainHash) bool {
	if hash == nil || other == nil {
		return hash == other
	}

	return *hash.hashArray == *other.hashArray
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

// DomainHashesToStrings returns a slice of strings representing the hashes in the given slice of hashes
func DomainHashesToStrings(hashes []*DomainHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}

	return strings
}
