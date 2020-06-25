package ff

import "testing"

func TestDeleteUpToFile(t *testing.T) {
	store, teardownFunc := prepareStoreForTest(t, "TestDeleteUpToFile")
	defer teardownFunc()

	// Set the maxFileSize to 16 bytes so that we don't have to write
	// an enormous amount of data to disk to get multiple files, all
	// for the sake of this test.
	currentMaxFileSize := maxFileSize
	maxFileSize = 16
	defer func() {
		maxFileSize = currentMaxFileSize
	}()

	const numberOfFiles = 1000
	dataToWrite := make([]byte, maxFileSize)
	for i := 0; i < numberOfFiles; i++ {
		_, err := store.write(dataToWrite)
		if err != nil {
			t.Fatalf("store.write(): %s", err)
		}
	}

	const minFileToKeep = 400
	preservedFiles := map[uint32]struct{}{
		100: {},
		123: {},
		572: {},
		250: {},
	}

	err := store.deleteUpToFile(minFileToKeep, preservedFiles)
	if err != nil {
		t.Fatalf("store.deleteUpToFile(): %s", err)
	}

	for fileNumber := range preservedFiles {
		exists, err := store.fileExists(fileNumber)
		if err != nil {
			t.Fatalf("store.fileExists(): %s", err)
		}
		if !exists {
			t.Errorf("file %d in preservedFiles was expected to be preserved", fileNumber)
		}
	}

	for fileNumber := uint32(0); fileNumber < numberOfFiles; fileNumber++ {
		exists, err := store.fileExists(fileNumber)
		if err != nil {
			t.Fatalf("store.fileExists(): %s", err)
		}

		if fileNumber < minFileToKeep {
			if _, ok := preservedFiles[fileNumber]; !ok && exists {
				t.Fatalf("file %d is lower than minFileToKeep (%d) "+
					"and was expected to be deleted", fileNumber, minFileToKeep)
			}
			continue
		}
		if !exists {
			t.Fatalf("file %d is greater or equal than minFileToKeep (%d) "+
				"and was expected to be preserved", fileNumber, minFileToKeep)
		}
	}
}
