package ffldb

import "github.com/kaspanet/kaspad/util/daghash"

// BlockRegion specifies a particular region of a block identified by the
// specified hash, given an offset and length.
type BlockRegion struct {
	Hash   *daghash.Hash
	Offset uint32
	Len    uint32
}
