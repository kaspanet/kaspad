package binaryserialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// SerializeHash serializes hash to a slice of bytes
func SerializeHash(hash *externalapi.DomainHash) []byte {
	return hash.ByteSlice()
}

// DeserializeHash deserializes a slice of bytes to a hash
func DeserializeHash(hashBytes []byte) (*externalapi.DomainHash, error) {
	return externalapi.NewDomainHashFromByteSlice(hashBytes)
}

// SerializeHashes serializes a slice of hashes to a slice of bytes
func SerializeHashes(hashes []*externalapi.DomainHash) []byte {
	buff := make([]byte, len(hashes)*externalapi.DomainHashSize)
	for i, hash := range hashes {
		copy(buff[externalapi.DomainHashSize*i:], hash.ByteSlice())
	}

	return buff
}

// DeserializeHashes deserializes a slice of bytes to a slice of hashes
func DeserializeHashes(hashesBytes []byte) ([]*externalapi.DomainHash, error) {
	if len(hashesBytes)%externalapi.DomainHashSize != 0 {
		return nil, errors.Errorf("The length of hashBytes is not divisible by externalapi.DomainHashSize (%d)",
			externalapi.DomainHashSize)
	}

	numHashes := len(hashesBytes) / externalapi.DomainHashSize
	hashes := make([]*externalapi.DomainHash, numHashes)
	for i := 0; i < numHashes; i++ {
		var err error
		start := i * externalapi.DomainHashSize
		end := i*externalapi.DomainHashSize + externalapi.DomainHashSize
		hashes[i], err = externalapi.NewDomainHashFromByteSlice(hashesBytes[start:end])
		if err != nil {
			return nil, err
		}
	}
	return hashes, nil
}
