package dbaccess

import (
	"github.com/kaspanet/kaspad/infrastructure/database"
)

// Context is an interface type representing the context in which queries run, currently relating to the
// existence or non-existence of a database transaction
// Call `.NoTx()` or `.NewTx()` to acquire a Context
type Context interface {
	accessor() (database.DataAccessor, error)
}

type noTxContext struct {
	backend *DatabaseContext
}

func (ctx *noTxContext) accessor() (database.DataAccessor, error) {
	return ctx.backend.db, nil
}

// TxContext represents a database context with an attached database transaction
type TxContext struct {
	dbTransaction database.Transaction
}

// NewTx returns an instance of TxContext with a new database transaction
func (ctx *DatabaseContext) NewTx() (*TxContext, error) {
	dbTransaction, err := ctx.db.Begin()
	if err != nil {
		return nil, err
	}
	return &TxContext{dbTransaction: dbTransaction}, nil
}

func (ctx *TxContext) accessor() (database.DataAccessor, error) {
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

// RollbackUnlessClosed rolls-back the transaction attached to this TxContext,
// unless the transaction had already been closed using either Rollback or Commit.
func (ctx *TxContext) RollbackUnlessClosed() error {
	return ctx.dbTransaction.RollbackUnlessClosed()
}
