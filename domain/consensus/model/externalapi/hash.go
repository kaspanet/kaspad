package externalapi

import "encoding/hex"

// DomainHashSize of array used to store hashes.
const DomainHashSize = 32

// DomainHash is the domain representation of a Hash
type DomainHash [DomainHashSize]byte

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash DomainHash) String() string {
	for i := 0; i < DomainHashSize/2; i++ {
		hash[i], hash[DomainHashSize-1-i] = hash[DomainHashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}

// Clone clones the hash
func (hash *DomainHash) Clone() *DomainHash {
	hashClone := *hash
	return &hashClone
}

// If this doesn't compile, it means the type definition has been changed, so it's
// an indication to update Equal accordingly.
var _ DomainHash = [DomainHashSize]byte{}

// Equal returns whether hash equals to other
func (hash *DomainHash) Equal(other *DomainHash) bool {
	if hash == nil || other == nil {
		return hash == other
	}

	return *hash == *other
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

// CloneHashes returns a clone of the given hashes slice
func CloneHashes(hashes []*DomainHash) []*DomainHash {
	clone := make([]*DomainHash, len(hashes))
	for i, hash := range hashes {
		clone[i] = hash.Clone()
	}
	return clone
}

// DomainHashesToStrings returns a slice of strings representing the hashes in the given slice of hashes
func DomainHashesToStrings(hashes []*DomainHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}

	return strings
}
