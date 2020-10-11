package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type DomainDBContext struct {
	dbContext *dbaccess.DatabaseContext
}

func (ddc *DomainDBContext) FetchBlockRelation(blockHash *model.DomainHash) (*model.BlockRelations, error) {
	// TODO: return dbaccess.FetchBlockRelations(ddc.dbContext, blockHash)
	return nil, nil
}

func NewDomainDBContext(dbContext *dbaccess.DatabaseContext) *DomainDBContext {
	return &DomainDBContext{dbContext: dbContext}
}
