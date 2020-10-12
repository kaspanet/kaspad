package hashes

import (
	"github.com/kaspanet/kaspad/domain/consensus/model"
	"math/big"
)

// ToBig converts a model.DomainHash into a big.Int that can be used to
// perform math comparisons.
func ToBig(hash *model.DomainHash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	buf := *hash
	blen := len(buf)
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}
