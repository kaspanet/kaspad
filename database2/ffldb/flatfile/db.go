package flatfile

// FFDB is a flat-file database. It supports opening multiple
// flat-file stores. See flatFileStore for further details.
type FFDB struct {
	path           string
	flatFileStores map[string]*flatFileStore
}

// NewFlatFileDatabase opens the flat-file database defined by
// the given path.
func NewFlatFileDatabase(path string) *FFDB {
	return &FFDB{
		path:           path,
		flatFileStores: make(map[string]*flatFileStore),
	}
}

// Close closes the flat-file database.
func (ffdb *FFDB) Close() error {
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
func (ffdb *FFDB) Write(storeName string, data []byte) ([]byte, error) {
	store := ffdb.store(storeName)
	location, err := store.write(data)
	if err != nil {
		return nil, err
	}
	serializedLocation := serializeLocation(location)
	return serializedLocation, nil
}

// Read reads data from the specified flat file store at the
// location specified by the given serialized location handle.
// See flatFileStore.read() for further details.
func (ffdb *FFDB) Read(storeName string, serializedLocation []byte) ([]byte, error) {
	store := ffdb.store(storeName)
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
func (ffdb *FFDB) CurrentLocation(storeName string) []byte {
	store := ffdb.store(storeName)
	currentLocation := store.currentLocation()
	return serializeLocation(currentLocation)
}

// Rollback truncates the flat-file store defined by the given
// storeName to the location defined by the given serialized
// location handle.
func (ffdb *FFDB) Rollback(storeName string, serializedLocation []byte) error {
	store := ffdb.store(storeName)
	location, err := deserializeLocation(serializedLocation)
	if err != nil {
		return err
	}
	return store.rollback(location)
}

func (ffdb *FFDB) store(storeName string) *flatFileStore {
	store, ok := ffdb.flatFileStores[storeName]
	if !ok {
		store = openFlatFileStore(ffdb.path, storeName)
		ffdb.flatFileStores[storeName] = store
	}
	return store
}
