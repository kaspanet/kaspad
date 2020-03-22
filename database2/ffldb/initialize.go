package ffldb

// initialize initializes the database. If this function fails it is
// irrecoverable, and likely indicates that database corruption had
// previously occurred.
func (db *ffldb) initialize() error {
	flatFiles, err := db.flatFiles()
	if err != nil {
		return err
	}
	for storeName, currentLocation := range flatFiles {
		err := db.tryRepair(storeName, currentLocation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *ffldb) flatFiles() (map[string][]byte, error) {
	flatFilesCursor := db.ldb.Cursor(flatFilesBucket.Path())
	defer func() {
		err := flatFilesCursor.Close()
		if err != nil {
			log.Warnf("cursor failed to close")
		}
	}()

	flatFiles := make(map[string][]byte)
	for flatFilesCursor.Next() {
		storeName, err := flatFilesCursor.Key()
		if err != nil {
			return nil, err
		}
		currentLocation, err := flatFilesCursor.Key()
		if err != nil {
			return nil, err
		}
		flatFiles[string(storeName)] = currentLocation
	}
	return flatFiles, nil
}

// tryRepair attempts to sync the block store with the current location value.
// Possible scenarios:
// a. currentLocation and the store are synced. Rollback does nothing.
// b. currentLocation is smaller than the store's location. Rollback truncates
//    the store.
// c. currentLocation is greater than the store's location. Rollback returns an
//    error. This indicates definite database corruption and is irrecoverable.
func (db *ffldb) tryRepair(storeName string, currentLocation []byte) error {
	return db.ffdb.Rollback(storeName, currentLocation)
}
