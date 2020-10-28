package dbmanager

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
)

type dbCursor struct {
	cursor database.Cursor
}

func (d dbCursor) Next() bool {
	return d.cursor.Next()
}

func (d dbCursor) First() bool {
	return d.cursor.First()
}

func (d dbCursor) Seek(key model.DBKey) error {
	return d.cursor.Seek(dbKeyToDatabaseKey(key))
}

func (d dbCursor) Key() (model.DBKey, error) {
	key, err := d.cursor.Key()
	if err != nil {
		return nil, err
	}

	return newDBKey(key), nil
}

func (d dbCursor) Value() ([]byte, error) {
	return d.cursor.Value()
}

func (d dbCursor) Close() error {
	return d.cursor.Close()
}

func newDBCursor(cursor database.Cursor) model.DBCursor {
	return &dbCursor{cursor: cursor}
}
