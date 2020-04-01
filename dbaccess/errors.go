package dbaccess

import "github.com/kaspanet/kaspad/database"

// IsNotFoundError checks whether an error is an ErrNotFound.
func IsNotFoundError(err error) bool {
	return database.IsNotFoundError(err)
}
