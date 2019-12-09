package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/kaspajson"
	"github.com/kaspanet/kaspad/util/daghash"
)

// handleGetTopHeaders implements the getTopHeaders command.
func handleGetTopHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*kaspajson.GetTopHeadersCmd)

	var startHash *daghash.Hash
	if c.StartHash != nil {
		startHash = &daghash.Hash{}
		err := daghash.Decode(startHash, *c.StartHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.StartHash)
		}
	}
	headers, err := s.cfg.DAG.GetTopHeaders(startHash)
	if err != nil {
		return nil, internalRPCError(err.Error(),
			"Failed to get top headers")
	}

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
