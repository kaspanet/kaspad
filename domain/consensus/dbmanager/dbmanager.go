package dbmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type dbManager struct {
	db database.Database
}

func (dbw *dbManager) Get(key model.DBKey) ([]byte, error) {
	panic("unimplemented")
}

func (dbw *dbManager) Has(key model.DBKey) (bool, error) {
	panic("unimplemented")
}

func (dbw *dbManager) Delete(key model.DBKey) error {
	panic("unimplemented")
}

func (dbw *dbManager) Cursor(bucket model.DBBucket) (model.DBCursor, error) {
	panic("unimplemented")
}

func (dbw *dbManager) Begin() (model.DBTransaction, error) {
	panic("unimplemented")
}

// New returns wraps the given database as an instance of model.DBManager
func New(db database.Database) model.DBManager {
	return &dbManager{db: db}
}
