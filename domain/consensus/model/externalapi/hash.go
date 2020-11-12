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
	if hash == nil {
		return nil
	}
	return &*hash
}

// DomainHashesToStrings returns a slice of strings representing the hashes in the given slice of hashes
func DomainHashesToStrings(hashes []*DomainHash) []string {
	strings := make([]string, len(hashes))
	for i, hash := range hashes {
		strings[i] = hash.String()
	}

	return strings
}
