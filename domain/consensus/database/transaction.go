package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type dbTransaction struct {
	transaction database.Transaction
}

func (d *dbTransaction) Get(key model.DBKey) ([]byte, error) {
	return d.transaction.Get(dbKeyToDatabaseKey(key))
}

func (d *dbTransaction) Has(key model.DBKey) (bool, error) {
	return d.transaction.Has(dbKeyToDatabaseKey(key))
}

func (d *dbTransaction) Cursor(bucket model.DBBucket) (model.DBCursor, error) {
	cursor, err := d.transaction.Cursor(dbBucketToDatabaseBucket(bucket))
	if err != nil {
		return nil, err
	}
	return newDBCursor(cursor), nil
}

func (d *dbTransaction) Put(key model.DBKey, value []byte) error {
	return d.transaction.Put(dbKeyToDatabaseKey(key), value)
}

func (d *dbTransaction) Delete(key model.DBKey) error {
	return d.transaction.Delete(dbKeyToDatabaseKey(key))
}

func (d *dbTransaction) Rollback() error {
	return d.transaction.Rollback()
}

func (d *dbTransaction) Commit() error {
	return d.transaction.Commit()
}

func (d *dbTransaction) RollbackUnlessClosed() error {
	return d.transaction.RollbackUnlessClosed()
}

func newDBTransaction(transaction database.Transaction) model.DBTransaction {
	return &dbTransaction{transaction: transaction}
}
