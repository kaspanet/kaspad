package database2

import (
	"github.com/kaspanet/kaspad/database2/ffldb"
	"github.com/pkg/errors"
)

// db is the kaspad database
var db *ffldb.Database

func DB() (*ffldb.Database, error) {
	if db == nil {
		return nil, errors.New("database is not open")
	}
	return db, nil
}

// Open opens to the database for given path
func Open(path string) error {
	if db != nil {
		return errors.New("database is already open")
	}
	db = ffldb.Open(path)
	return nil
}

// Close closes the database, if it's open
func Close() error {
	if db == nil {
		return nil
	}
	err := db.Close()
	db = nil
	return err
}
