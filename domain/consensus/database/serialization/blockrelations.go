package serialization

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
)

// DomainBlockRelationsToDbBlockRelations converts model.BlockRelations to DbBlockRelations
func DomainBlockRelationsToDbBlockRelations(domainBlockRelations *model.BlockRelations) *DbBlockRelations {
	return &DbBlockRelations{
		Parents:  DomainHashesToDbHashes(domainBlockRelations.Parents),
		Children: DomainHashesToDbHashes(domainBlockRelations.Children),
	}
}

// DbBlockRelationsToDomainBlockRelations converts DbBlockRelations to model.BlockRelations
func DbBlockRelationsToDomainBlockRelations(dbBlockRelations *DbBlockRelations) (*model.BlockRelations, error) {
	domainParentHashes, err := DbHashesToDomainHashes(dbBlockRelations.Parents)
	if err != nil {
		return nil, err
	}
	domainChildHashes, err := DbHashesToDomainHashes(dbBlockRelations.Children)
	if err != nil {
		return nil, err
	}

	return &model.BlockRelations{
		Parents:  domainParentHashes,
		Children: domainChildHashes,
	}, nil
}
