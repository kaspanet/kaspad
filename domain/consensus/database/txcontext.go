package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/dbaccess"
)

type DomainTxContext struct {
	dbTx *dbaccess.TxContext
}

func (dtc *DomainTxContext) StoreBlockRelation(blockHash *model.DomainHash, blockRelationData *model.BlockRelations) error {
	// TODO: return dbaccess.StoreBlockRelation(ddc.dbTx, blockHash, blockRelationData)
	return nil
}

func NewDomainTxContext(dbTx *dbaccess.TxContext) *DomainTxContext {
	return &DomainTxContext{dbTx: dbTx}
}
