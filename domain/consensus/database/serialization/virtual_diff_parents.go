package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
)

// VirtualDiffParentsToDBHeaderVirtualDiffParents converts a slice of hashes to DbVirtualDiffParents
func VirtualDiffParentsToDBHeaderVirtualDiffParents(tips []*externalapi.DomainHash) *DbVirtualDiffParents {
	return &DbVirtualDiffParents{
		VirtualDiffParents: DomainHashesToDbHashes(tips),
	}
}

// DBVirtualDiffParentsToVirtualDiffParents converts DbHeaderTips to a slice of hashes
func DBVirtualDiffParentsToVirtualDiffParents(dbVirtualDiffParents *DbVirtualDiffParents) ([]*externalapi.DomainHash, error) {
	return DbHashesToDomainHashes(dbVirtualDiffParents.VirtualDiffParents)
}
