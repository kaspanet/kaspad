package ff

// FlatFileDB is a flat-file database. It supports opening
// multiple flat-file stores. See flatFileStore for further
// details.
type FlatFileDB struct {
	path           string
	flatFileStores map[string]*flatFileStore
}

// NewFlatFileDB opens the flat-file database defined by
// the given path.
func NewFlatFileDB(path string) *FlatFileDB {
	return &FlatFileDB{
		path:           path,
		flatFileStores: make(map[string]*flatFileStore),
	}
}

// Close closes the flat-file database.
func (ffdb *FlatFileDB) Close() error {
	for _, store := range ffdb.flatFileStores {
		err := store.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Write appends the specified data bytes to the specified store.
// It returns a serialized location handle that's meant to be
// stored and later used when querying the data that has just now
// been inserted.
// See flatFileStore.write() for further details.
func (ffdb *FlatFileDB) Write(storeName string, data []byte) ([]byte, error) {
	store, err := ffdb.store(storeName)
	if err != nil {
		return nil, err
	}
	location, err := store.write(data)
	if err != nil {
		return nil, err
	}
	return serializeLocation(location), nil
}

// Read reads data from the specified flat file store at the
// location specified by the given serialized location handle.
// It returns ErrNotFound if the location does not exist.
// See flatFileStore.read() for further details.
func (ffdb *FlatFileDB) Read(storeName string, serializedLocation []byte) ([]byte, error) {
	store, err := ffdb.store(storeName)
	if err != nil {
		return nil, err
	}
	location, err := deserializeLocation(serializedLocation)
	if err != nil {
		return nil, err
	}
	return store.read(location)
}

// CurrentLocation returns the serialized location handle to
// the current location within the flat file store defined
// storeName. It is mainly to be used to rollback flat-file
// stores in case of data incongruency.
func (ffdb *FlatFileDB) CurrentLocation(storeName string) ([]byte, error) {
	store, err := ffdb.store(storeName)
	if err != nil {
		return nil, err
	}
	currentLocation := store.currentLocation()
	return serializeLocation(currentLocation), nil
}

// Rollback truncates the flat-file store defined by the given
// storeName to the location defined by the given serialized
// location handle.
func (ffdb *FlatFileDB) Rollback(storeName string, serializedLocation []byte) error {
	store, err := ffdb.store(storeName)
	if err != nil {
		return err
	}
	location, err := deserializeLocation(serializedLocation)
	if err != nil {
		return err
	}
	return store.rollback(location)
}

func (ffdb *FlatFileDB) store(storeName string) (*flatFileStore, error) {
	store, ok := ffdb.flatFileStores[storeName]
	if !ok {
		var err error
		store, err = openFlatFileStore(ffdb.path, storeName)
		if err != nil {
			return nil, err
		}
		ffdb.flatFileStores[storeName] = store
	}
	return store, nil
}
