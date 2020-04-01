package database

import "errors"

// ErrNotFound denotes that the requested item was not
// found in the database.
var ErrNotFound = errors.New("not found")

// IsNotFoundError checks whether an error is an ErrNotFound.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}
