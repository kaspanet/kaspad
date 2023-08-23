package database

import (
	"github.com/c4ei/yunseokyeol/infrastructure/db/database"
)

// ErrNotFound denotes that the requested item was not
// found in the database.
var ErrNotFound = database.ErrNotFound

// IsNotFoundError checks whether an error is an ErrNotFound.
func IsNotFoundError(err error) bool {
	return database.IsNotFoundError(err)
}
