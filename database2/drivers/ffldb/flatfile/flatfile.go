package flatfile

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"
)

const (
	// maxOpenFiles is the max number of open files to maintain in each store's
	// cache. Note that this does not include the current/write file, so there
	// will typically be one more than this value open.
	maxOpenFiles = 25

	// maxFileSize is the maximum size for each file used to store data.
	//
	// NOTE: The current code uses uint32 for all offsets, so this value
	// must be less than 2^32 (4 GiB). This is also why it's a typed
	// constant.
	maxFileSize uint32 = 512 * 1024 * 1024 // 512 MiB
)

var (
	// byteOrder is the preferred byte order used through the flat files.
	// Sometimes big endian will be used to allow ordered byte sortable
	// integer values.
	byteOrder = binary.LittleEndian

	// castagnoli houses the Catagnoli polynomial used for CRC-32 checksums.
	castagnoli = crc32.MakeTable(crc32.Castagnoli)
)

// flatFileStoreStore houses information used to handle reading and writing data
// into flat files with support for multiple concurrent readers.
type flatFileStore struct {
	// basePath is the base path used for the flat files.
	basePath string

	// storeName is the name of this flat-file store.
	storeName string

	// The following fields are related to the flat files which hold the
	// actual data. The number of open files is limited by maxOpenFiles.
	//
	// obfMutex protects concurrent access to the openFiles map. It is
	// a RWMutex so multiple readers can simultaneously access open files.
	//
	// openFiles houses the open file handles for existing files which have
	// been opened read-only along with an individual RWMutex. This scheme
	// allows multiple concurrent readers to the same file while preventing
	// the file from being closed out from under them.
	//
	// lruMutex protects concurrent access to the least recently used list
	// and lookup map.
	//
	// openFilesLRU tracks how the open files are referenced by pushing the
	// most recently used files to the front of the list thereby trickling
	// the least recently used files to end of the list. When a file needs
	// to be closed due to exceeding the the max number of allowed open
	// files, the one at the end of the list is closed.
	//
	// fileNumToLRUElem is a mapping between a specific file number and the
	// associated list element on the least recently used list.
	//
	// Thus, with the combination of these fields, the database supports
	// concurrent non-blocking reads across multiple and individual files
	// along with intelligently limiting the number of open file handles by
	// closing the least recently used files as needed.
	//
	// NOTE: The locking order used throughout is well-defined and MUST be
	// followed. Failure to do so could lead to deadlocks. In particular,
	// the locking order is as follows:
	//   1) obfMutex
	//   2) lruMutex
	//   3) writeCursor mutex
	//   4) specific file mutexes
	//
	// None of the mutexes are required to be locked at the same time, and
	// often aren't. However, if they are to be locked simultaneously, they
	// MUST be locked in the order previously specified.
	//
	// Due to the high performance and multi-read concurrency requirements,
	// write locks should only be held for the minimum time necessary.
	obfMutex         sync.RWMutex
	lruMutex         sync.Mutex
	openFilesLRU     *list.List // Contains uint32 file numbers.
	fileNumToLRUElem map[uint32]*list.Element
	openFiles        map[uint32]*lockableFile

	// writeCursor houses the state for the current file and location that
	// new data is written to.
	writeCursor *writeCursor
}

// lockableFile represents a flat file on disk that has been opened for either
// read or read/write access. It also contains a read-write mutex to support
// multiple concurrent readers.
type lockableFile struct {
	sync.RWMutex
	file filer
}

// filer is an interface which acts very similar to a *os.File and is typically
// implemented by it. It exists so the test code can provide mock files for
// properly testing corruption and file system issues.
type filer interface {
	io.Closer
	io.WriterAt
	io.ReaderAt
	Truncate(size int64) error
	Sync() error
}

// writeCursor represents the current file and offset of the flat file on disk
// for performing all writes. It also contains a read-write mutex to support
// multiple concurrent readers which can reuse the file handle.
type writeCursor struct {
	sync.RWMutex

	// currentFile is the current file that will be appended to when writing
	// new data.
	currentFile *lockableFile

	// currentFileNumber is the current file number and is used to allow
	// readers to use the same open file handle.
	currentFileNumber uint32

	// currentOffset is the offset in the current file where the next new
	// data will be written.
	currentOffset uint32
}

// flatFileLocation identifies a particular flat file location.
type flatFileLocation struct {
	fileNumber uint32
	fileOffset uint32
	fileLength uint32
}

// newFlatFileStore returns a new flat file store with the current file number
// and offset set and all fields initialized.
func newFlatFileStore(basePath string, storeName string) *flatFileStore {
	// Look for the end of the latest file to determine what the write cursor
	// position is from the viewpoint of the flat files on disk.
	fileNumber, fileOffset := scanFlatFiles(basePath, storeName)

	store := &flatFileStore{
		basePath:         basePath,
		storeName:        storeName,
		openFiles:        make(map[uint32]*lockableFile),
		openFilesLRU:     list.New(),
		fileNumToLRUElem: make(map[uint32]*list.Element),

		writeCursor: &writeCursor{
			currentFile:       &lockableFile{},
			currentFileNumber: fileNumber,
			currentOffset:     fileOffset,
		},
	}
	return store
}

// scanFlatFiles searches the database directory for all flat files for a given
// store to find the end of the most recent file. This position is considered
// the current write cursor.
func scanFlatFiles(dbPath string, storeName string) (fileNumber uint32, fileLength uint32) {
	for {
		filePath := flatFilePath(dbPath, storeName, fileNumber)
		stat, err := os.Stat(filePath)
		if err != nil {
			break
		}
		fileLength = uint32(stat.Size())

		fileNumber++
	}

	log.Tracef("Scan for store '%s' found latest file #%d with length %d",
		storeName, fileNumber, fileLength)
	return fileNumber, fileLength
}

// flatFilePath return the file path for the provided store's flat file number.
func flatFilePath(dbPath string, storeName string, fileNumber uint32) string {
	// Choose 9 digits of precision for the filenames. 9 digits provide
	// 10^9 files @ 512MiB each a total of ~476.84PiB.

	fileName := fmt.Sprintf("%s-%09d.fdb", storeName, fileNumber)
	return filepath.Join(dbPath, fileName)
}

// write appends the specified rdata bytes to the store's write cursor location
// and increments it accordingly. When the data would exceed the max file size
// for the current flat file, this function will close the current file, create
// the next file, update the write cursor, and write the data to the new file.
//
// The write cursor will also be advanced the number of bytes actually written
// in the event of failure.
//
// Format: <data length><data><checksum>
func (s *flatFileStore) write(data []byte) (*flatFileLocation, error) {
	// Compute how many bytes will be written.
	// 4 bytes for data length + length of the data + 4 bytes for checksum.
	dataLength := uint32(len(data))
	fullLength := dataLength + 8

	// Move to the next file if adding the new data would exceed the max
	// allowed size for the current flat file. Also detect overflow because
	// even though it isn't possible currently, numbers/ might change in
	// the future to make it possible.
	//
	// NOTE: The writeCursor.offset field isn't protected by the mutex
	// since it's only read/changed during this function which can only be
	// called during a write transaction, of which there can be only one at
	// a time.
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

	// Block length.
	byteOrder.PutUint32(scratch[:], dataLength)
	if err := s.writeData(scratch[:], "block length"); err != nil {
		return nil, err
	}
	_, _ = hasher.Write(scratch[:])

	// Data.
	if err := s.writeData(data[:], "data"); err != nil {
		return nil, err
	}
	_, _ = hasher.Write(data)

	// Castagnoli CRC-32 as a checksum of all the previous.
	if err := s.writeData(hasher.Sum(nil), "checksum"); err != nil {
		return nil, err
	}

	location := &flatFileLocation{
		fileNumber: cursor.currentFileNumber,
		fileOffset: originalOffset,
		fileLength: fullLength,
	}
	return location, nil
}

// openWriteFile returns a file handle for the passed flat file number in
// read/write mode. The file will be created if needed. It is typically used
// for the current file that will have all new data appended. Unlike openFile,
// this function does not keep track of the open file and it is not subject to
// the maxOpenFiles limit.
func (s *flatFileStore) openWriteFile(fileNumber uint32) (filer, error) {
	// The current flat file needs to be read-write so it is possible to
	// append to it. Also, it shouldn't be part of the least recently used
	// file.
	filePath := flatFilePath(s.basePath, s.storeName, fileNumber)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.Errorf("failed to open file %q: %s",
			filePath, err)
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
			panic("No space left on the hard disk, exiting...")
		}
		return errors.Errorf("failed to write %s in store %s to file %d "+
			"at offset %d: %s", fieldName, s.storeName, cursor.currentFileNumber,
			cursor.currentOffset-uint32(n), err)
	}

	return nil
}
