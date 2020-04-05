package ff

import (
	"github.com/kaspanet/kaspad/util/panics"
	"github.com/pkg/errors"
	"hash/crc32"
	"os"
	"syscall"
)

// write appends the specified data bytes to the store's write cursor location
// and increments it accordingly. When the data would exceed the max file size
// for the current flat file, this function will close the current file, create
// the next file, update the write cursor, and write the data to the new file.
//
// The write cursor will also be advanced the number of bytes actually written
// in the event of failure.
//
// Format: <data length><data><checksum>
func (s *flatFileStore) write(data []byte) (*flatFileLocation, error) {
	if s.isClosed {
		return nil, errors.Errorf("cannot write to a closed store %s",
			s.storeName)
	}

	// Compute how many bytes will be written.
	// 4 bytes for data length + length of the data + 4 bytes for checksum.
	dataLength := uint32(len(data))
	fullLength := uint32(dataLengthLength) + dataLength + uint32(crc32ChecksumLength)

	// Move to the next file if adding the new data would exceed the max
	// allowed size for the current flat file. Also detect overflow because
	// even though it isn't possible currently, numbers might change in
	// the future to make it possible.
	//
	// NOTE: The writeCursor.currentOffset field isn't protected by the
	// mutex since it's only read/changed during this function which can
	// only be called during a write transaction, of which there can be
	// only one at a time.
	cursor := s.writeCursor
	finalOffset := cursor.currentOffset + fullLength
	if finalOffset < cursor.currentOffset || finalOffset > maxFileSize {
		// This is done under the write cursor lock since the curFileNum
		// field is accessed elsewhere by readers.
		//
		// Close the current write file to force a read-only reopen
		// with LRU tracking. The close is done under the write lock
		// for the file to prevent it from being closed out from under
		// any readers currently reading from it.
		cursor.Lock()
		cursor.currentFile.Lock()
		if cursor.currentFile.file != nil {
			_ = cursor.currentFile.file.Close()
			cursor.currentFile.file = nil
		}
		cursor.currentFile.Unlock()

		// Start writes into next file.
		cursor.currentFileNumber++
		cursor.currentOffset = 0
		cursor.Unlock()
	}

	// All writes are done under the write lock for the file to ensure any
	// readers are finished and blocked first.
	cursor.currentFile.Lock()
	defer cursor.currentFile.Unlock()

	// Open the current file if needed. This will typically only be the
	// case when moving to the next file to write to or on initial database
	// load. However, it might also be the case if rollbacks happened after
	// file writes started during a transaction commit.
	if cursor.currentFile.file == nil {
		file, err := s.openWriteFile(cursor.currentFileNumber)
		if err != nil {
			return nil, err
		}
		cursor.currentFile.file = file
	}

	originalOffset := cursor.currentOffset
	hasher := crc32.New(castagnoli)
	var scratch [4]byte

	// Data length.
	byteOrder.PutUint32(scratch[:], dataLength)
	err := s.writeData(scratch[:], "data length")
	if err != nil {
		return nil, err
	}
	_, _ = hasher.Write(scratch[:])

	// Data.
	err = s.writeData(data[:], "data")
	if err != nil {
		return nil, err
	}
	_, _ = hasher.Write(data)

	// Castagnoli CRC-32 as a checksum of all the previous.
	err = s.writeData(hasher.Sum(nil), "checksum")
	if err != nil {
		return nil, err
	}

	// Sync the file to disk.
	err = cursor.currentFile.file.Sync()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sync file %d "+
			"in store '%s'", cursor.currentFileNumber, s.storeName)
	}

	location := &flatFileLocation{
		fileNumber: cursor.currentFileNumber,
		fileOffset: originalOffset,
		dataLength: fullLength,
	}
	return location, nil
}

// openWriteFile returns a file handle for the passed flat file number in
// read/write mode. The file will be created if needed. It is typically used
// for the current file that will have all new data appended. Unlike openFile,
// this function does not keep track of the open file and it is not subject to
// the maxOpenFiles limit.
func (s *flatFileStore) openWriteFile(fileNumber uint32) (file, error) {
	// The current flat file needs to be read-write so it is possible to
	// append to it. Also, it shouldn't be part of the least recently used
	// file.
	filePath := flatFilePath(s.basePath, s.storeName, fileNumber)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %q",
			filePath)
	}

	return file, nil
}

// writeData is a helper function for write which writes the provided data at
// the current write offset and updates the write cursor accordingly. The field
// name parameter is only used when there is an error to provide a nicer error
// message.
//
// The write cursor will be advanced the number of bytes actually written in the
// event of failure.
//
// NOTE: This function MUST be called with the write cursor current file lock
// held and must only be called during a write transaction so it is effectively
// locked for writes. Also, the write cursor current file must NOT be nil.
func (s *flatFileStore) writeData(data []byte, fieldName string) error {
	cursor := s.writeCursor
	n, err := cursor.currentFile.file.WriteAt(data, int64(cursor.currentOffset))
	cursor.currentOffset += uint32(n)
	if err != nil {
		var pathErr *os.PathError
		if ok := errors.As(err, &pathErr); ok && pathErr.Err == syscall.ENOSPC {
			panics.Exit(log, "No space left on the hard disk.")
		}
		return errors.Wrapf(err, "failed to write %s in store %s to file %d "+
			"at offset %d", fieldName, s.storeName, cursor.currentFileNumber,
			cursor.currentOffset-uint32(n))
	}

	return nil
}
