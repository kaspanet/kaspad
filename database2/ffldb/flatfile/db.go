package flatfile

type FlatFileDB struct {
	path           string
	flatFileStores map[string]*flatFileStore
}

func NewFlatFileDB(path string) *FlatFileDB {
	return &FlatFileDB{
		path:           path,
		flatFileStores: make(map[string]*flatFileStore),
	}
}

func (ffdb *FlatFileDB) Write(storeName string, data []byte) ([]byte, error) {
	store := ffdb.store(storeName)
	location, err := store.write(data)
	if err != nil {
		return nil, err
	}
	serializedLocation := serializeLocation(location)
	return serializedLocation, nil
}

func (ffdb *FlatFileDB) Read(storeName string, serializedLocation []byte) ([]byte, error) {
	store := ffdb.store(storeName)
	location, err := deserializeLocation(serializedLocation)
	if err != nil {
		return nil, err
	}
	return store.read(location)
}

func (ffdb *FlatFileDB) Rollback(storeName string, serializedLocation []byte) error {
	store := ffdb.store(storeName)
	location, err := deserializeLocation(serializedLocation)
	if err != nil {
		return err
	}
	store.rollback(location)
	return nil
}

func (ffdb *FlatFileDB) store(storeName string) *flatFileStore {
	store, ok := ffdb.flatFileStores[storeName]
	if !ok {
		store = openFlatFileStore(ffdb.path, storeName)
		ffdb.flatFileStores[storeName] = store
	}
	return store
}
