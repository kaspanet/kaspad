package ff

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"hash/crc32"
	"os"
	"path/filepath"
	"sync"
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

	// crc32ByteOrder is the byte order used for CRC-32 checksums.
	crc32ByteOrder = binary.BigEndian

	// crc32ChecksumLength is the length in bytes of a CRC-32 checksum.
	crc32ChecksumLength = 4

	// dataLengthLength is the length in bytes of the "data length" section
	// of a serialized entry in a flat file store.
	dataLengthLength = 4

	// castagnoli houses the Catagnoli polynomial used for CRC-32 checksums.
	castagnoli = crc32.MakeTable(crc32.Castagnoli)
)

// flatFileStore houses information used to handle reading and writing data
// into flat files with support for multiple concurrent readers.
type flatFileStore struct {
	// basePath is the base path used for the flat files.
	basePath string

	// storeName is the name of this flat-file store.
	storeName string

	// The following fields are related to the flat files which hold the
	// actual data. The number of open files is limited by maxOpenFiles.
	//
	// openFilesMutex protects concurrent access to the openFiles map. It
	// is a RWMutex so multiple readers can simultaneously access open
	// files.
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
	// to be closed due to exceeding the max number of allowed open
	// files, the one at the end of the list is closed.
	//
	// fileNumberToLRUElement is a mapping between a specific file number and
	// the associated list element on the least recently used list.
	//
	// Thus, with the combination of these fields, the database supports
	// concurrent non-blocking reads across multiple and individual files
	// along with intelligently limiting the number of open file handles by
	// closing the least recently used files as needed.
	//
	// NOTE: The locking order used throughout is well-defined and MUST be
	// followed. Failure to do so could lead to deadlocks. In particular,
	// the locking order is as follows:
	//   1) openFilesMutex
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
	openFilesMutex         sync.RWMutex
	openFiles              map[uint32]*lockableFile
	lruMutex               sync.Mutex
	openFilesLRU           *list.List // Contains uint32 file numbers.
	fileNumberToLRUElement map[uint32]*list.Element

	// writeCursor houses the state for the current file and location that
	// new data is written to.
	writeCursor *writeCursor

	// isClosed is true when the store is closed. Any operations on a closed
	// store will fail.
	isClosed bool
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

// openFlatFileStore returns a new flat file store with the current file number
// and offset set and all fields initialized.
func openFlatFileStore(basePath string, storeName string) (*flatFileStore, error) {
	// Look for the end of the latest file to determine what the write cursor
	// position is from the viewpoint of the flat files on disk.
	fileNumber, fileOffset, err := findCurrentLocation(basePath, storeName)
	if err != nil {
		return nil, err
	}

	store := &flatFileStore{
		basePath:               basePath,
		storeName:              storeName,
		openFiles:              make(map[uint32]*lockableFile),
		openFilesLRU:           list.New(),
		fileNumberToLRUElement: make(map[uint32]*list.Element),
		writeCursor: &writeCursor{
			currentFile:       &lockableFile{},
			currentFileNumber: fileNumber,
			currentOffset:     fileOffset,
		},
		isClosed: false,
	}
	return store, nil
}

func (s *flatFileStore) Close() error {
	if s.isClosed {
		return errors.Errorf("cannot close a closed store %s",
			s.storeName)
	}
	s.isClosed = true

	// Close the write cursor. We lock the write cursor here
	// to let it finish any undergoing writing.
	s.writeCursor.Lock()
	defer s.writeCursor.Unlock()
	err := s.writeCursor.currentFile.Close()
	if err != nil {
		return err
	}

	// Close all open files
	for _, openFile := range s.openFiles {
		err := openFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *flatFileStore) currentLocation() *flatFileLocation {
	return &flatFileLocation{
		fileNumber: s.writeCursor.currentFileNumber,
		fileOffset: s.writeCursor.currentOffset,
		dataLength: 0,
	}
}

// findCurrentLocation searches the database directory for all flat files for a given
// store to find the end of the most recent file. This position is considered
// the current write cursor.
func findCurrentLocation(dbPath string, storeName string) (fileNumber uint32, fileLength uint32, err error) {
	currentFileNumber := uint32(0)
	currentFileLength := uint32(0)
	for {
		currentFilePath := flatFilePath(dbPath, storeName, currentFileNumber)
		stat, err := os.Stat(currentFilePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return 0, 0, errors.WithStack(err)
			}
			if currentFileNumber > 0 {
				fileNumber = currentFileNumber - 1
			}
			fileLength = currentFileLength
			break
		}
		currentFileLength = uint32(stat.Size())
		currentFileNumber++
	}

	log.Tracef("Scan for store '%s' found latest file #%d with length %d",
		storeName, fileNumber, fileLength)
	return fileNumber, fileLength, nil
}

// flatFilePath return the file path for the provided store's flat file number.
func flatFilePath(dbPath string, storeName string, fileNumber uint32) string {
	// Choose 9 digits of precision for the filenames. 9 digits provide
	// 10^9 files @ 512MiB each a total of ~476.84PiB.

	fileName := fmt.Sprintf("%s-%09d.fdb", storeName, fileNumber)
	return filepath.Join(dbPath, fileName)
}
