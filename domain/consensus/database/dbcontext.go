package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// DomainDBContext is a proxy over a dbaccess.DatabaseContext
type DomainDBContext struct {
	dbContext *dbaccess.DatabaseContext
}

// FetchBlockRelation retrieves the BlockRelation for the given blockHash
func (ddc *DomainDBContext) FetchBlockRelation(blockHash *model.DomainHash) (*model.BlockRelations, error) {
	// TODO: return dbaccess.FetchBlockRelations(ddc.dbContext, blockHash)
	return nil, nil
}

// NewDomainDBContext creates a new DomainDBContext
func NewDomainDBContext(dbContext *dbaccess.DatabaseContext) *DomainDBContext {
	return &DomainDBContext{dbContext: dbContext}
}
