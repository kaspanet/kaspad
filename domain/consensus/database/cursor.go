package database

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/kaspanet/kaspad/infrastructure/db/database"
	"github.com/pkg/errors"
)

type dbCursor struct {
	cursor   database.Cursor
	isClosed bool
}

func (d dbCursor) Next() bool {
	if d.isClosed {
		panic("Tried using a closed DBCursor")
	}

	return d.cursor.Next()
}

func (d dbCursor) First() bool {
	if d.isClosed {
		panic("Tried using a closed DBCursor")
	}
	return d.cursor.First()
}

func (d dbCursor) Seek(key model.DBKey) error {
	if d.isClosed {
		return errors.New("Tried using a closed DBCursor")
	}
	return d.cursor.Seek(dbKeyToDatabaseKey(key))
}

func (d dbCursor) Key() (model.DBKey, error) {
	if d.isClosed {
		return nil, errors.New("Tried using a closed DBCursor")
	}
	key, err := d.cursor.Key()
	if err != nil {
		return nil, err
	}

	return newDBKey(key), nil
}

func (d dbCursor) Value() ([]byte, error) {
	if d.isClosed {
		return nil, errors.New("Tried using a closed DBCursor")
	}
	return d.cursor.Value()
}

func (d dbCursor) Close() error {
	if d.isClosed {
		return errors.New("Tried using a closed DBCursor")
	}
	d.isClosed = true
	err := d.cursor.Close()
	if err != nil {
		return err
	}
	d.cursor = nil
	return nil
}

func newDBCursor(cursor database.Cursor) model.DBCursor {
	return &dbCursor{cursor: cursor}
}
