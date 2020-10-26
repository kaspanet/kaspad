package externalapi

import "encoding/hex"

// DomainHashSize of array used to store hashes.
const DomainHashSize = 32

// DomainHash is the domain representation of a daghash.Hash
type DomainHash [DomainHashSize]byte

// String returns the Hash as the hexadecimal string of the byte-reversed
// hash.
func (hash DomainHash) String() string {
	for i := 0; i < DomainHashSize/2; i++ {
		hash[i], hash[DomainHashSize-1-i] = hash[DomainHashSize-1-i], hash[i]
	}
	return hex.EncodeToString(hash[:])
}
