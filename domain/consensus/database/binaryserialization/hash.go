package binaryserialization

import "github.com/kaspanet/kaspad/domain/consensus/model/externalapi"

// SerializeHash serializes hash to a slice of bytes
func SerializeHash(hash *externalapi.DomainHash) []byte {
	return hash.ByteSlice()
}

// DeserializeHash a slice of bytes to a hash
func DeserializeHash(hashBytes []byte) (*externalapi.DomainHash, error) {
	return externalapi.NewDomainHashFromByteSlice(hashBytes)
}
