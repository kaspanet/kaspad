package externalapi

// DomainHashSize of array used to store hashes.
const DomainHashSize = 32

// DomainHash is the domain representation of a daghash.Hash
type DomainHash [DomainHashSize]byte
