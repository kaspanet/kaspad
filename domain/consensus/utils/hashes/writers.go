package hashes

import (
	"crypto/sha256"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"hash"
)

// HashWriter is used to incrementally hash data without concatenating all of the data to a single buffer
// it exposes an io.Writer api and a Finalize function to get the resulting hash.
// The used hash function is double-sha256.
type HashWriter struct {
	inner hash.Hash
}

// NewHashWriter Returns a new HashWriter
func NewHashWriter() *HashWriter {
	return &HashWriter{sha256.New()}
}

// Write will always return (len(p), nil)
func (h *HashWriter) Write(p []byte) (n int, err error) {
	return h.inner.Write(p)
}

// Finalize returns the resulting double hash
func (h *HashWriter) Finalize() externalapi.DomainHash {
	firstHashInTheSum := h.inner.Sum(nil)
	return sha256.Sum256(firstHashInTheSum)
}

// HashData hashes the given byte slice
func HashData(data []byte) externalapi.DomainHash {
	w := NewHashWriter()
	_, err := w.Write(data)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. SHA256's digest should never return an error"))
	}

	return w.Finalize()
}
