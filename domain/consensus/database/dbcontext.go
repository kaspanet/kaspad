package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// DomainDBContext is a proxy over a dbaccess.DatabaseContext
type DomainDBContext struct {
	dbContext *dbaccess.DatabaseContext
}

// FetchBlockRelation retrieves the BlockRelation for the given blockHash
func (ddc *DomainDBContext) FetchBlockRelation(blockHash *externalapi.DomainHash) (*model.BlockRelations, error) {
	// TODO: return dbaccess.FetchBlockRelations(ddc.dbContext, blockHash)
	return nil, nil
}

// NewTx returns an instance of DomainTxContext with a new database transaction
func (ddc *DomainDBContext) NewTx() (*DomainTxContext, error) {
	txContext, err := ddc.dbContext.NewTx()

	if err != nil {
		return nil, err
	}

	return NewDomainTxContext(txContext), nil
}

// NewDomainDBContext creates a new DomainDBContext
func NewDomainDBContext(dbContext *dbaccess.DatabaseContext) *DomainDBContext {
	return &DomainDBContext{dbContext: dbContext}
}
