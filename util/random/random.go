package random

import (
	"crypto/rand"
	"io"

	"github.com/kaspanet/kaspad/util/binaryserializer"
)

// randomUint64 returns a cryptographically random uint64 value. This
// unexported version takes a reader primarily to ensure the error paths
// can be properly tested by passing a fake reader in the tests.
func randomUint64(r io.Reader) (uint64, error) {
	rv, err := binaryserializer.Uint64(r)
	if err != nil {
		return 0, err
	}
	return rv, nil
}

// Uint64 returns a cryptographically random uint64 value.
func Uint64() (uint64, error) {
	return randomUint64(rand.Reader)
}
