package flatfile

import (
	"container/list"
	"hash/crc32"
	"io"
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
