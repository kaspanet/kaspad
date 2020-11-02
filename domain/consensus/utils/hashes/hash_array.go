package hashes

import (
	"bytes"

	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

func DeserializeHashSlice(hashesBytes []byte) ([]*externalapi.DomainHash, error) {
	if len(hashesBytes)%externalapi.DomainHashSize != 0 {
		return nil, errors.Errorf("serialized hashes length is %d bytes, while it should be a multiple of %d",
			len(hashesBytes), externalapi.DomainHashSize)
	}

	hashes := make([]*externalapi.DomainHash, 0, len(hashesBytes)/externalapi.DomainHashSize)

	for i := 0; i < len(hashesBytes); i += externalapi.DomainHashSize {
		hashBytes := hashesBytes[i : i+externalapi.DomainHashSize]
		hash, err := FromBytes(hashBytes)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, hash)
	}

	return hashes, nil
}

func SerializeHashSlice(hashes []*externalapi.DomainHash) []byte {
	hashesBytes := make([][]byte, 0, len(hashes))

	for _, hash := range hashes {
		hashesBytes = append(hashesBytes, hash[:])
	}

	return bytes.Join(hashesBytes, []byte{})
}
