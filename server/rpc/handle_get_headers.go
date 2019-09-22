package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/daglabs/btcd/btcjson"
	"github.com/daglabs/btcd/util/daghash"
)

// handleGetHeaders implements the getHeaders command.
//
// NOTE: This is a btcsuite extension originally ported from
// github.com/decred/dcrd.
func handleGetHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*btcjson.GetHeadersCmd)

	startHash := &daghash.ZeroHash
	if c.StartHash != "" {
		err := daghash.Decode(startHash, c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.StopHash)
		}
	}
	stopHash := &daghash.ZeroHash
	if c.StopHash != "" {
		err := daghash.Decode(stopHash, c.StopHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.StopHash)
		}
	}
	headers := s.cfg.SyncMgr.GetBlueBlocksHeadersBetween(startHash, stopHash)

	// Return the serialized block headers as hex-encoded strings.
	hexBlockHeaders := make([]string, len(headers))
	var buf bytes.Buffer
	for i, h := range headers {
		err := h.Serialize(&buf)
		if err != nil {
			return nil, internalRPCError(err.Error(),
				"Failed to serialize block header")
		}
		hexBlockHeaders[i] = hex.EncodeToString(buf.Bytes())
		buf.Reset()
	}
	return hexBlockHeaders, nil
}
