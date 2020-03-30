package dbaccess

import "github.com/kaspanet/kaspad/database2"

// IsNotFoundError checks whether an error is an ErrNotFound.
func IsNotFoundError(err error) bool {
	return database2.IsNotFoundError(err)
}
