package dbaccess

var (
	dagStateKey = []byte("dag-state")
)

// StoreDAGState stores the DAG state in the database.
func StoreDAGState(context Context, dagState []byte) error {
	accessor, err := context.accessor()
	if err != nil {
		return err
	}
	return accessor.Put(dagStateKey, dagState)
}

// FetchDAGState retrieves the DAG state from the database.
func FetchDAGState(context Context) (data []byte, found bool, err error) {
	accessor, err := context.accessor()
	if err != nil {
		return nil, false, err
	}
	return accessor.Get(dagStateKey)
}
