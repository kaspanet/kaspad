package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
	"github.com/kaspanet/kaspad/database2/ffldb"
)

type Context interface {
	db() (ffldb.Database, error)
}

type noTxContext struct{}

var noTxContextSingleton = &noTxContext{}

func (*noTxContext) db() (ffldb.Database, error) {
	return database2.DB()
}

// NoTx creates and returns an instance of dbaccess.Context without an attached database transaction
func NoTx() Context {
	return noTxContextSingleton
}

// TxContext represents a database context with an attached database transaction
type TxContext struct {
	dbTransaction *ffldb.Transaction
}

func (ctx *TxContext) db() (ffldb.Database, error) {
	return ctx.dbTransaction, nil
}

// Commit commits the transaction attached to this TxContext
func (ctx *TxContext) Commit() error {
	return ctx.dbTransaction.Commit()
}

// Rollback rolls-back the transaction attached to this TxContext
func (ctx *TxContext) Rollback() error {
	return ctx.dbTransaction.Rollback()
}

// NewTx returns an instance of TxContext with a new database transaction
func NewTx() (*TxContext, error) {
	db, err := database2.DB()
	if err != nil {
		return nil, err
	}
	dbTransaction, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &TxContext{dbTransaction: dbTransaction}, nil
}
