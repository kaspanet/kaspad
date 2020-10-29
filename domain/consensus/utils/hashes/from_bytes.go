package hashes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
)

// FromBytes creates a DomainHash from the given byte slice
func FromBytes(hashBytes []byte) (*externalapi.DomainHash, error) {
	if len(hashBytes) != externalapi.DomainHashSize {
		return nil, errors.Errorf("invalid hash size. Want: %d, got: %d",
			externalapi.DomainHashSize, len(hashBytes))
	}
	var domainHash externalapi.DomainHash
	copy(domainHash[:], hashBytes)
	return &domainHash, nil
}
