package protocol

import "github.com/kaspanet/kaspad/util/daghash"

type PeerState struct {
	requestedBlocks map[*daghash.Hash]struct{}
}
