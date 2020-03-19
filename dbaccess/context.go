package dbaccess

import (
	"github.com/kaspanet/kaspad/database2"
)

// Context is an interface type representing the context in which queries run, currently relating to the
// existence or non-existence of a database transaction
// Call `.NoTx()` or `.NewTx()` to acquire a Context
type Context interface {
	db() (database2.Database, error)
}

type noTxContext struct{}

var noTxContextSingleton = &noTxContext{}

func (*noTxContext) db() (database2.Database, error) {
	return db()
}

// NoTx creates and returns an instance of dbaccess.Context without an attached database transaction
func NoTx() Context {
	return noTxContextSingleton
}

// TxContext represents a database context with an attached database transaction
type TxContext struct {
	dbTransaction database2.Transaction
}

func (ctx *TxContext) db() (database2.Database, error) {
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
	db, err := db()
	if err != nil {
		return nil, err
	}
	dbTransaction, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &TxContext{dbTransaction: dbTransaction}, nil
}
