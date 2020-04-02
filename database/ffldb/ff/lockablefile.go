package ff

import (
	"github.com/pkg/errors"
	"io"
	"sync"
)

// lockableFile represents a flat file on disk that has been opened for either
// read or read/write access. It also contains a read-write mutex to support
// multiple concurrent readers.
type lockableFile struct {
	sync.RWMutex
	file

	isClosed bool
}

// file is an interface which acts very similar to a *os.File and is typically
// implemented by it. It exists so the test code can provide mock files for
// properly testing corruption and file system issues.
type file interface {
	io.Closer
	io.WriterAt
	io.ReaderAt
	Truncate(size int64) error
	Sync() error
}

func (lf *lockableFile) Close() error {
	if lf.isClosed {
		return errors.Errorf("cannot close an already closed file")
	}
	lf.isClosed = true

	lf.Lock()
	defer lf.Unlock()

	return errors.WithStack(lf.file.Close())
}
