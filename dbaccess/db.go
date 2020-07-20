package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb"
)

type DatabaseContext struct {
	db database.Database
	*noTxContext
}

func New(path string) (*DatabaseContext, error) {
	db, err := ffldb.Open(path)
	if err != nil {
		return nil, err
	}

	databaseContext := &DatabaseContext{db: db}
	databaseContext.noTxContext = &noTxContext{backend: databaseContext}

	return databaseContext, nil
}

// Close closes the database, if it's open
func (ctx *DatabaseContext) Close() error {
	return ctx.db.Close()
}
