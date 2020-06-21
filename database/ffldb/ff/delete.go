package ff

import (
	"github.com/pkg/errors"
	"os"
)

// deleteFile removes the file for the passed flat file number.
// This function MUST be called with the lruMutex and the openFilesMutex
// held for writes.
func (s *flatFileStore) deleteFile(fileNumber uint32) error {
	if s.isClosed {
		return errors.Errorf("cannot delete file in a closed store %s", s.storeName)
	}

	// Cleanup the file before deleting it
	if file, ok := s.openFiles[fileNumber]; ok {
		file.Lock()
		defer file.Unlock()
		err := file.Close()
		if err != nil {
			return err
		}

		lruElement := s.fileNumberToLRUElement[fileNumber]
		s.openFilesLRU.Remove(lruElement)
		delete(s.openFiles, fileNumber)
		delete(s.fileNumberToLRUElement, fileNumber)
	}

	// Delete the file from disk
	filePath := flatFilePath(s.basePath, s.storeName, fileNumber)
	return errors.WithStack(os.Remove(filePath))
}

func (s *flatFileStore) deleteUpToFile(minFileToKeep uint32, preservedFiles map[uint32]struct{}) error {
	if s.isClosed {
		return errors.Errorf("cannot delete files in a closed store %s",
			s.storeName)
	}

	if minFileToKeep > s.writeCursor.currentFileNumber {
		return errors.New("cannot delete up to location which is after the write cursor")
	}

	s.lruMutex.Lock()
	defer s.lruMutex.Unlock()
	s.openFilesMutex.Lock()
	defer s.openFilesMutex.Unlock()

	for fileNumber := uint32(0); fileNumber < minFileToKeep; fileNumber++ {
		if _, ok := preservedFiles[fileNumber]; ok {
			continue
		}

		exists, err := s.fileExists(fileNumber)
		if err != nil {
			return err
		}
		if !exists {
			continue
		}

		err = s.deleteFile(fileNumber)
		if err != nil {
			return errors.Wrapf(err, "failed to delete file "+
				"number %d in store '%s'", fileNumber, s.storeName)
		}
	}

	return nil
}

func (s *flatFileStore) fileExists(fileNumber uint32) (bool, error) {
	if s.isClosed {
		return false, errors.Errorf("cannot check for existing files in a closed store %s", s.storeName)
	}
	filePath := flatFilePath(s.basePath, s.storeName, fileNumber)
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.WithStack(err)
	}
	return true, nil
}
