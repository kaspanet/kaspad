package dbaccess

import "github.com/kaspanet/kaspad/infrastructure/database"

var (
	peersKey = database.MakeBucket().Key([]byte("peers"))
)

// StorePeersState stores the peers state in the database.
func StorePeersState(context Context, peersState []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}
	return accessor.Put(peersKey, peersState)
}

// FetchPeersState retrieves the peers state from the database.
// Returns ErrNotFound if the state is missing from the database.
func FetchPeersState(context Context) ([]byte, error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, err
	}
	return accessor.Get(peersKey)
}
