package rpc

import (
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleResolveFinalityConflict implements the resolveFinalityConflict command.
func handleResolveFinalityConflict(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.ResolveFinalityConflictCmd)

	validBlockHashes := make([]*daghash.Hash, 0, len(c.ValidBlockHashes))
	for _, validBlockHashHex := range c.ValidBlockHashes {
		validBlockHash, err := daghash.NewHashFromStr(validBlockHashHex)
		if err != nil {
			return nil, err
		}
		validBlockHashes = append(validBlockHashes, validBlockHash)
	}

	invalidBlockHashes := make([]*daghash.Hash, 0, len(c.InvalidBlockHashes))
	for _, invalidBlockHashHex := range c.InvalidBlockHashes {
		invalidBlockHash, err := daghash.NewHashFromStr(invalidBlockHashHex)
		if err != nil {
			return nil, err
		}
		invalidBlockHashes = append(invalidBlockHashes, invalidBlockHash)
	}

	return nil, s.dag.ResolveFinalityConflict(c.FinalityConflictID, validBlockHashes, invalidBlockHashes)
}
