package dbaccess

import (
	"github.com/kaspanet/kaspad/database2/ffldb"
	"github.com/pkg/errors"
)

// dbSingleton is an instance of the kaspad database
var dbSingleton *ffldb.FFLDB

// db returns a reference to the database
func db() (*ffldb.FFLDB, error) {
	if dbSingleton == nil {
		return nil, errors.New("database is not open")
	}
	return dbSingleton, nil
}

// Open opens to the database for given path
func Open(path string) error {
	if dbSingleton != nil {
		return errors.New("database is already open")
	}

	openedDB, err := ffldb.Open(path)
	if err != nil {
		return err
	}

	dbSingleton = openedDB
	return nil
}

// Close closes the database, if it's open
func Close() error {
	if dbSingleton == nil {
		return nil
	}
	err := dbSingleton.Close()
	dbSingleton = nil
	return err
}
