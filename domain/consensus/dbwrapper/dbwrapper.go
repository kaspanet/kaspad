package dbwrapper

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type dbWrapper struct {
	db database.Database
}

func (dbw *dbWrapper) Get(key model.DBKey) ([]byte, error) {
	panic("unimplemented")
}

func (dbw *dbWrapper) Has(key model.DBKey) (bool, error) {
	panic("unimplemented")
}

func (dbw *dbWrapper) Delete(key model.DBKey) error {
	panic("unimplemented")
}

func (dbw *dbWrapper) Cursor(bucket model.DBBucket) (model.DBCursor, error) {
	panic("unimplemented")
}

func (dbw *dbWrapper) Begin() (model.DBTransaction, error) {
	panic("unimplemented")
}

// New returns wraps the given database as an instance of model.DBManager
func New(db database.Database) model.DBManager {
	return &dbWrapper{db: db}
}
