package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashes"
)

// DbHashToDomainHash converts a DbHash to a DomainHash
func DbHashToDomainHash(dbHash *DbHash) (*externalapi.DomainHash, error) {
	return hashes.FromBytes(dbHash.Hash)
}

// DomainHashToDbHash converts a DomainHash to a DbHash
func DomainHashToDbHash(domainHash *externalapi.DomainHash) *DbHash {
	return &DbHash{Hash: domainHash[:]}
}
