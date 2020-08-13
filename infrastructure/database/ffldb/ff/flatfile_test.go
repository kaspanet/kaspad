package ff

import (
	"bytes"
	"github.com/kaspanet/kaspad/infrastructure/database"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func prepareStoreForTest(t *testing.T, testName string) (store *flatFileStore, teardownFunc func()) {
	// Create a temp db to run tests against
	path, err := ioutil.TempDir("", testName)
	if err != nil {
		t.Fatalf("%s: TempDir unexpectedly "+
			"failed: %s", testName, err)
	}
	name := "test"
	store, err = openFlatFileStore(path, name)
	if err != nil {
		t.Fatalf("%s: openFlatFileStore "+
			"unexpectedly failed: %s", testName, err)
	}
	teardownFunc = func() {
		err = store.Close()
		if err != nil {
			t.Fatalf("%s: Close unexpectedly "+
				"failed: %s", testName, err)
		}
	}
	return store, teardownFunc
}

func TestFlatFileStoreSanity(t *testing.T) {
	store, teardownFunc := prepareStoreForTest(t, "TestFlatFileStoreSanity")
	defer teardownFunc()

	// Write something to the store
	writeData := []byte("Hello world!")
	location, err := store.write(writeData)
	if err != nil {
		t.Fatalf("TestFlatFileStoreSanity: Write returned "+
			"unexpected error: %s", err)
	}

	// Read from the location previously written to
	readData, err := store.read(location)
	if err != nil {
		t.Fatalf("TestFlatFileStoreSanity: read returned "+
			"unexpected error: %s", err)
	}

	// Make sure that the written data and the read data are equal
	if !reflect.DeepEqual(readData, writeData) {
		t.Fatalf("TestFlatFileStoreSanity: read data and "+
			"write data are not equal. Wrote: %s, read: %s",
			string(writeData), string(readData))
	}
}

func TestFlatFilePath(t *testing.T) {
	tests := []struct {
		dbPath       string
		storeName    string
		fileNumber   uint32
		expectedPath string
	}{
		{
			dbPath:       "path",
			storeName:    "store",
			fileNumber:   0,
			expectedPath: "path/store-000000000.fdb",
		},
		{
			dbPath:       "path/to/database",
			storeName:    "blocks",
			fileNumber:   123456789,
			expectedPath: "path/to/database/blocks-123456789.fdb",
		},
	}

	for _, test := range tests {
		path := flatFilePath(test.dbPath, test.storeName, test.fileNumber)
		if path != test.expectedPath {
			t.Errorf("TestFlatFilePath: unexpected path. Want: %s, got: %s",
				test.expectedPath, path)
		}
	}
}

func TestFlatFileMultiFileRollback(t *testing.T) {
	store, teardownFunc := prepareStoreForTest(t, "TestFlatFileMultiFileRollback")
	defer teardownFunc()

	// Set the maxFileSize to 16 bytes so that we don't have to write
	// an enormous amount of data to disk to get multiple files, all
	// for the sake of this test.
	currentMaxFileSize := maxFileSize
	maxFileSize = 16
	defer func() {
		maxFileSize = currentMaxFileSize
	}()

	// Write five 8 byte chunks and keep the last location written to
	var lastWriteLocation1 *flatFileLocation
	for i := byte(0); i < 5; i++ {
		writeData := []byte{i, i, i, i, i, i, i, i}
		var err error
		lastWriteLocation1, err = store.write(writeData)
		if err != nil {
			t.Fatalf("TestFlatFileMultiFileRollback: write returned "+
				"unexpected error: %s", err)
		}
	}

	// Grab the current location and the current file number
	currentLocation := store.currentLocation()
	fileNumberBeforeWriting := store.writeCursor.currentFileNumber

	// Write (2 * maxOpenFiles) more 8 byte chunks and keep the last location written to
	var lastWriteLocation2 *flatFileLocation
	for i := byte(0); i < byte(2*maxFileSize); i++ {
		writeData := []byte{0, 1, 2, 3, 4, 5, 6, 7}
		var err error
		lastWriteLocation2, err = store.write(writeData)
		if err != nil {
			t.Fatalf("TestFlatFileMultiFileRollback: write returned "+
				"unexpected error: %s", err)
		}
	}

	// Grab the file number again to later make sure its file no longer exists
	fileNumberAfterWriting := store.writeCursor.currentFileNumber

	// Rollback
	err := store.rollback(currentLocation)
	if err != nil {
		t.Fatalf("TestFlatFileMultiFileRollback: rollback returned "+
			"unexpected error: %s", err)
	}

	// Make sure that lastWriteLocation1 still exists
	expectedData := []byte{4, 4, 4, 4, 4, 4, 4, 4}
	data, err := store.read(lastWriteLocation1)
	if err != nil {
		t.Fatalf("TestFlatFileMultiFileRollback: read returned "+
			"unexpected error: %s", err)
	}
	if !bytes.Equal(data, expectedData) {
		t.Fatalf("TestFlatFileMultiFileRollback: read returned "+
			"unexpected data. Want: %s, got: %s", string(expectedData),
			string(data))
	}

	// Make sure that lastWriteLocation2 does NOT exist
	_, err = store.read(lastWriteLocation2)
	if err == nil {
		t.Fatalf("TestFlatFileMultiFileRollback: read " +
			"unexpectedly succeeded")
	}
	if !database.IsNotFoundError(err) {
		t.Fatalf("TestFlatFileMultiFileRollback: read "+
			"returned unexpected error: %s", err)
	}

	// Make sure that all the appropriate files have been deleted
	for i := fileNumberAfterWriting; i > fileNumberBeforeWriting; i-- {
		filePath := flatFilePath(store.basePath, store.storeName, i)
		if _, err := os.Stat(filePath); err == nil || !os.IsNotExist(err) {
			t.Fatalf("TestFlatFileMultiFileRollback: file "+
				"unexpectedly still exists: %s", filePath)
		}
	}
}
