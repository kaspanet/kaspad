package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb"
)

// DatabaseContext represents a context in which all database queries run
type DatabaseContext struct {
	db database.Database
	*noTxContext
}

// New creates a new DatabaseContext with database is in the specified `path`
func New(path string) (*DatabaseContext, error) {
	db, err := ffldb.Open(path)
	if err != nil {
		return nil, err
	}

	databaseContext := &DatabaseContext{db: db}
	databaseContext.noTxContext = &noTxContext{backend: databaseContext}

	return databaseContext, nil
}

// Close closes the DatabaseContext's connection, if it's open
func (ctx *DatabaseContext) Close() error {
	return ctx.db.Close()
}
