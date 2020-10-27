package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

// DomainTxContext is a proxy over a dbaccess.TxContext
type DomainTxContext struct {
	dbTx *dbaccess.TxContext
}

// StoreBlockRelation stores the given BlockRelation
func (dtc *DomainTxContext) StoreBlockRelation(blockHash *externalapi.DomainHash, blockRelationData *model.BlockRelations) error {
	// TODO: return dbaccess.StoreBlockRelation(ddc.dbTx, blockHash, blockRelationData)
	return nil
}

// NewDomainTxContext creates a new DomainTXContext
func NewDomainTxContext(dbTx *dbaccess.TxContext) *DomainTxContext {
	return &DomainTxContext{dbTx: dbTx}
}

func (dtc *DomainTxContext) Commit() error {
	return dtc.dbTx.Commit()
}
