package externalapi

// HashSize of array used to store hashes.
const HashSize = 32

// DomainHash is the domain representation of a daghash.Hash
type DomainHash [HashSize]byte
