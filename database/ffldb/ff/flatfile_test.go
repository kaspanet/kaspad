package ff

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestFlatFileStoreSanity(t *testing.T) {
	// Open a test store
	path, err := ioutil.TempDir("", "TestFlatFileStoreSanity")
	if err != nil {
		t.Fatalf("TestFlatFileStoreSanity: TempDir unexpectedly "+
			"failed: %s", err)
	}
	name := "test"
	store, err := openFlatFileStore(path, name)
	if err != nil {
		t.Fatalf("TestFlatFileStoreSanity: openFlatFileStore "+
			"unexpectedly failed: %s", err)
	}

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
