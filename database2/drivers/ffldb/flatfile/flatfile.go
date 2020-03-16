package flatfile

import (
	"container/list"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
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
	// to be closed due to exceeding the the max number of allowed open
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
