package hashes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/pkg/errors"
	"hash"
)

// HashWriter is used to incrementally hash data without concatenating all of the data to a single buffer
// it exposes an io.Writer api and a Finalize function to get the resulting hash.
// The used hash function is blake2b.
// This can only be created via one of the domain separated constructors
type HashWriter struct {
	hash.Hash
}

// InfallibleWrite is just like write but doesn't return anything
func (h HashWriter) InfallibleWrite(p []byte) {
	// This write can never return an error, this is part of the hash.Hash interface contract.
	_, err := h.Write(p)
	if err != nil {
		panic(errors.Wrap(err, "this should never happen. hash.Hash interface promises to not return errors."))
	}
}

// Finalize returns the resulting hash
func (h HashWriter) Finalize() *externalapi.DomainHash {
	var sum externalapi.DomainHash
	// This should prevent `Sum` for allocating an output buffer, by using the DomainHash buffer. we still copy because we don't want to rely on that.
	copy(sum[:], h.Sum(sum[:0]))
	return &sum
}
