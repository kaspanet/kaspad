package dbaccess

import (
	"github.com/kaspanet/kaspad/database"
	"github.com/kaspanet/kaspad/database/ffldb"
	"github.com/pkg/errors"
)

// dbSingleton is a handle to an instance of the kaspad database
var dbSingleton database.Database

// db returns a handle to the database
func db() (database.Database, error) {
	if dbSingleton == nil {
		return nil, errors.New("database is not open")
	}
	return dbSingleton, nil
}

// Open opens the database for given path
func Open(path string) error {
	if dbSingleton != nil {
		return errors.New("database is already open")
	}

	db, err := ffldb.Open(path)
	if err != nil {
		return err
	}

	dbSingleton = db
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
