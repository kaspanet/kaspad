package ffldb

// initialize initializes the database. If this function fails then the
// database is irrecoverably corrupted.
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
	flatFilesBucketPath := flatFilesBucket.Path()
	flatFilesCursor := db.levelDB.Cursor(flatFilesBucketPath)
	defer func() {
		err := flatFilesCursor.Close()
		if err != nil {
			log.Warnf("cursor failed to close")
		}
	}()

	flatFiles := make(map[string][]byte)
	for flatFilesCursor.Next() {
		storeNameKey, err := flatFilesCursor.Key()
		if err != nil {
			return nil, err
		}
		storeName := string(storeNameKey)

		currentLocation, err := flatFilesCursor.Value()
		if err != nil {
			return nil, err
		}
		flatFiles[storeName] = currentLocation
	}
	return flatFiles, nil
}

// tryRepair attempts to sync the store with the current location value.
// Possible scenarios:
// a. currentLocation and the store are synced. Rollback does nothing.
// b. currentLocation is smaller than the store's location. Rollback truncates
//    the store.
// c. currentLocation is greater than the store's location. Rollback returns an
//    error. This indicates definite database corruption and is irrecoverable.
func (db *ffldb) tryRepair(storeName string, currentLocation []byte) error {
	return db.flatFileDB.Rollback(storeName, currentLocation)
}
