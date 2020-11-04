package dbaccess

import "github.com/kaspanet/kaspad/infrastructure/db/database"

var (
	peersKey = database.MakeBucket().Key([]byte("peers"))
)

// StorePeersState stores the peers state in the database.
func StorePeersState(dbContext database.DataAccessor, peersState []byte) error {
	return dbContext.Put(peersKey, peersState)
}

// FetchPeersState retrieves the peers state from the database.
// Returns ErrNotFound if the state is missing from the database.
func FetchPeersState(dbContext database.DataAccessor) ([]byte, error) {
	return dbContext.Get(peersKey)
}
