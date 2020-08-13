package daghash

import (
	"crypto/sha256"
	"fmt"
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
func (h *HashWriter) Finalize() Hash {
	res := Hash{}
	// Can never happen, Sha256's Sum is 32 bytes.
	err := res.SetBytes(h.inner.Sum(nil))
	if err != nil {
		panic(fmt.Sprintf("Should never fail, sha256.Sum is 32 bytes and so is daghash.Hash: '%+v'", err))
	}
	return res
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
func (h *DoubleHashWriter) Finalize() Hash {
	firstHashInTheSum := h.inner.Sum(nil)
	return sha256.Sum256(firstHashInTheSum)
}
