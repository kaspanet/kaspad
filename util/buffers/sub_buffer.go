package buffers

import (
	"bytes"
	"github.com/pkg/errors"
)

// SubBuffer lets you write to an existing buffer
// and let you check with the `Bytes()` method what
// has been written to the underlying buffer using
// the sub buffer.
type SubBuffer struct {
	buff       *bytes.Buffer
	start, end int
}

// Bytes returns all the bytes that were written to the sub buffer.
func (s *SubBuffer) Bytes() []byte {
	return s.buff.Bytes()[s.start:s.end]
}

// Write writes to the sub buffer's underlying buffer
// and increases s.end by the number of bytes written
// so s.Bytes() will be able to return the written bytes.
func (s *SubBuffer) Write(p []byte) (int, error) {
	if s.buff.Len() > s.end || s.buff.Len() < s.start {
		return 0, errors.New("a sub buffer cannot be written after another entity wrote or read from its " +
			"underlying buffer")
	}

	n, err := s.buff.Write(p)
	if err != nil {
		return 0, err
	}

	s.end += n

	return n, nil
}

// NewSubBuffer returns a new sub buffer.
func NewSubBuffer(buff *bytes.Buffer) *SubBuffer {
	return &SubBuffer{
		buff:  buff,
		start: buff.Len(),
		end:   buff.Len(),
	}
}
