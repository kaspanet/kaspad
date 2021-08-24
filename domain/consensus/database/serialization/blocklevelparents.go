package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// DbBlockLevelParentsToDomainBlockLevelParents converts a DbBlockLevelParents to a BlockLevelParents
func DbBlockLevelParentsToDomainBlockLevelParents(dbBlockLevelParents *DbBlockLevelParents) (externalapi.BlockLevelParents, error) {
	domainBlockLevelParents := make(externalapi.BlockLevelParents, len(dbBlockLevelParents.ParentHashes))
	for i, parentHash := range dbBlockLevelParents.ParentHashes {
		var err error
		domainBlockLevelParents[i], err = externalapi.NewDomainHashFromByteSlice(parentHash.Hash)
		if err != nil {
			return nil, err
		}
	}
	return domainBlockLevelParents, nil
}

// DomainBlockLevelParentsToDbBlockLevelParents converts a BlockLevelParents to a DbBlockLevelParents
func DomainBlockLevelParentsToDbBlockLevelParents(domainBlockLevelParents externalapi.BlockLevelParents) *DbBlockLevelParents {
	parentHashes := make([]*DbHash, len(domainBlockLevelParents))
	for i, parentHash := range domainBlockLevelParents {
		parentHashes[i] = &DbHash{Hash: parentHash.ByteSlice()}
	}
	return &DbBlockLevelParents{ParentHashes: parentHashes}
}

// DomainParentsToDbParents converts a slice of BlockLevelParents to a slice of DbBlockLevelParents
func DomainParentsToDbParents(domainParents []externalapi.BlockLevelParents) []*DbBlockLevelParents {
	dbParents := make([]*DbBlockLevelParents, len(domainParents))
	for i, domainBlockLevelParents := range domainParents {
		dbParents[i] = DomainBlockLevelParentsToDbBlockLevelParents(domainBlockLevelParents)
	}
	return dbParents
}

// DbParentsToDomainParents converts a slice of DbBlockLevelParents to a slice of BlockLevelParents
func DbParentsToDomainParents(dbParents []*DbBlockLevelParents) ([]externalapi.BlockLevelParents, error) {
	domainParents := make([]externalapi.BlockLevelParents, len(dbParents))
	for i, domainBlockLevelParents := range dbParents {
		var err error
		domainParents[i], err = DbBlockLevelParentsToDomainBlockLevelParents(domainBlockLevelParents)
		if err != nil {
			return nil, err
		}
	}
	return domainParents, nil
}
