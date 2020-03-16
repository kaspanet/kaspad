package flatfile

import (
	"github.com/pkg/errors"
	"hash/crc32"
	"os"
)

// read reads the specified flat file record and returns the data. It ensures
// the integrity of the data by comparing the calculated checksum against the
// one stored in the flat file. This function also automatically handles all
// file management such as opening and closing files as necessary to stay
// within the maximum allowed open files limit.
//
// Format: <data length><data><checksum>
func (s *FlatFileStore) read(location *flatFileLocation) ([]byte, error) {
	// Get the referenced flat file handle opening the file as needed. The
	// function also handles closing files as needed to avoid going over the
	// max allowed open files.
	flatFile, err := s.flatFile(location.fileNumber)
	if err != nil {
		return nil, err
	}

	data := make([]byte, location.fileLength)
	n, err := flatFile.file.ReadAt(data, int64(location.fileOffset))
	flatFile.RUnlock()
	if err != nil {
		return nil, errors.Errorf("failed to read data in store '%s' "+
			"from file %d, offset %d: %s", s.storeName, location.fileNumber,
			location.fileOffset, err)
	}

	// Calculate the checksum of the read data and ensure it matches the
	// serialized checksum.
	serializedChecksum := crc32ByteOrder.Uint32(data[n-4:])
	calculatedChecksum := crc32.Checksum(data[:n-4], castagnoli)
	if serializedChecksum != calculatedChecksum {
		return nil, errors.Errorf("data in store '%s' does not match "+
			"checksum - got %x, want %x", s.storeName, calculatedChecksum,
			serializedChecksum)
	}

	// The data excludes the length of the data and the checksum.
	return data[4 : n-4], nil
}

// flatFile attempts to return an existing file handle for the passed flat file
// number if it is already open as well as marking it as most recently used. It
// will also open the file when it's not already open subject to the rules
// described in openFile.
//
// NOTE: The returned flat file will already have the read lock acquired and
// the caller MUST call .RUnlock() to release it once it has finished all read
// operations. This is necessary because otherwise it would be possible for a
// separate goroutine to close the file after it is returned from here, but
// before the caller has acquired a read lock.
func (s *FlatFileStore) flatFile(fileNumber uint32) (*lockableFile, error) {
	// When the requested flat file is open for writes, return it.
	cursor := s.writeCursor
	cursor.RLock()
	if fileNumber == cursor.currentFileNumber && cursor.currentFile.file != nil {
		openFile := cursor.currentFile
		openFile.RLock()
		cursor.RUnlock()
		return openFile, nil
	}
	cursor.RUnlock()

	// Try to return an open file under the overall files read lock.
	s.openFilesMutex.RLock()
	if openFile, ok := s.openFiles[fileNumber]; ok {
		s.lruMutex.Lock()
		s.openFilesLRU.MoveToFront(s.fileNumberToLRUElement[fileNumber])
		s.lruMutex.Unlock()

		openFile.RLock()
		s.openFilesMutex.RUnlock()
		return openFile, nil
	}
	s.openFilesMutex.RUnlock()

	// Since the file isn't open already, need to check the open files map
	// again under write lock in case multiple readers got here and a
	// separate one is already opening the file.
	s.openFilesMutex.Lock()
	if openFlatFile, ok := s.openFiles[fileNumber]; ok {
		openFlatFile.RLock()
		s.openFilesMutex.Unlock()
		return openFlatFile, nil
	}

	// The file isn't open, so open it while potentially closing the least
	// recently used one as needed.
	openFile, err := s.openFile(fileNumber)
	if err != nil {
		s.openFilesMutex.Unlock()
		return nil, err
	}
	openFile.RLock()
	s.openFilesMutex.Unlock()
	return openFile, nil
}

// openFile returns a read-only file handle for the passed flat file number.
// The function also keeps track of the open files, performs least recently
// used tracking, and limits the number of open files to maxOpenFiles by closing
// the least recently used file as needed.
//
// This function MUST be called with the open files mutex (s.openFilesMutex)
// locked for WRITES.
func (s *FlatFileStore) openFile(fileNumber uint32) (*lockableFile, error) {
	// Open the appropriate file as read-only.
	filePath := flatFilePath(s.basePath, s.storeName, fileNumber)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	flatFile := &lockableFile{file: file}

	// Close the least recently used file if the file exceeds the max
	// allowed open files. This is not done until after the file open in
	// case the file fails to open, there is no need to close any files.
	//
	// A write lock is required on the LRU list here to protect against
	// modifications happening as already open files are read from and
	// shuffled to the front of the list.
	//
	// Also, add the file that was just opened to the front of the least
	// recently used list to indicate it is the most recently used file and
	// therefore should be closed last.
	s.lruMutex.Lock()
	lruList := s.openFilesLRU
	if lruList.Len() >= maxOpenFiles {
		lruFileNumber := lruList.Remove(lruList.Back()).(uint32)
		oldFile := s.openFiles[lruFileNumber]

		// Close the old file under the write lock for the file in case
		// any readers are currently reading from it so it's not closed
		// out from under them.
		oldFile.Lock()
		_ = oldFile.file.Close()
		oldFile.Unlock()

		delete(s.openFiles, lruFileNumber)
		delete(s.fileNumberToLRUElement, lruFileNumber)
	}
	s.fileNumberToLRUElement[fileNumber] = lruList.PushFront(fileNumber)
	s.lruMutex.Unlock()

	// Store a reference to it in the open files map.
	s.openFiles[fileNumber] = flatFile

	return flatFile, nil
}
