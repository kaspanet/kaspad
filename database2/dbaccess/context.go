package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
)

type Context interface {
	db() (database2.Database, error)
}

type noTxContext struct{}

var noTxContextSingleton = &noTxContext{}

func (*noTxContext) db() (database2.Database, error) {
	return database2.DB()
}

// NoTx creates and returns an instance of dbaccess.Context without an attached database transaction
func NoTx() Context {
	return noTxContextSingleton
}

// TxContext represents a database context with an attached database transaction
type TxContext struct {
	dbInstance database2.Database
}

func (ctx *TxContext) db() (database2.Database, error) {
	return ctx.dbInstance, nil
}

// Commit commits the transaction attached to this TxContext
func (ctx *TxContext) Commit() error {
	return ctx.dbInstance.Commit()
}

// Rollback rolls-back the transaction attached to this TxContext
func (ctx *TxContext) Rollback() error {
	return ctx.dbInstance.Rollback()
}

// NewTx returns an instance of TxContext with a new database transaction
func NewTx() (*TxContext, error) {
	db, err := database2.DB()
	if err != nil {
		return nil, err
	}
	return &TxContext{dbInstance: db.Begin()}, nil
}
