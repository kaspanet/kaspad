package hashes

import (
	"crypto/sha256"
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"github.com/pkg/errors"
	"hash"
)

// HashWriter is used to incrementally hash data without concatenating all of the data to a single buffer
// it exposes an io.Writer api and a Finalize function to get the resulting hash.
// HashWriter.Write(slice).Finalize == HashH(slice)
type HashWriter struct {
	inner hash.Hash
}

// DoubleHashWriter is used to incrementally double hash data without concatenating all of the data to a single buffer
// it exposes an io.Writer api and a Finalize function to get the resulting hash.
// DoubleHashWriter.Write(slice).Finalize == DoubleHashH(slice)
type DoubleHashWriter struct {
	inner hash.Hash
}

// NewHashWriter returns a new Hash Writer
func NewHashWriter() *HashWriter {
	return &HashWriter{sha256.New()}
}

// Write will always return (len(p), nil)
func (h *HashWriter) Write(p []byte) (n int, err error) {
	return h.inner.Write(p)
}

// Finalize returns the resulting hash
func (h *HashWriter) Finalize() model.DomainHash {
	res := h.inner.Sum(nil)
	if len(res) != model.HashSize {
		panic(errors.Errorf("should never fail, sha256.Sum result is expected to be %d bytes, but is "+
			"%d bytes", model.HashSize, len(res)))
	}

	var domainHash model.DomainHash
	copy(domainHash[:], res)
	return domainHash
}

// NewDoubleHashWriter Returns a new DoubleHashWriter
func NewDoubleHashWriter() *DoubleHashWriter {
	return &DoubleHashWriter{sha256.New()}
}

// Write will always return (len(p), nil)
func (h *DoubleHashWriter) Write(p []byte) (n int, err error) {
	return h.inner.Write(p)
}

// Finalize returns the resulting double hash
func (h *DoubleHashWriter) Finalize() model.DomainHash {
	firstHashInTheSum := h.inner.Sum(nil)
	return sha256.Sum256(firstHashInTheSum)
}
