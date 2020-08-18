package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/infrastructure/network/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

const getTopHeadersMaxHeaders = getHeadersMaxHeaders

// handleGetTopHeaders implements the getTopHeaders command.
func handleGetTopHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetTopHeadersCmd)

	var highHash *daghash.Hash
	if c.HighHash != nil {
		highHash = &daghash.Hash{}
		err := daghash.Decode(highHash, *c.HighHash)
		if err != nil {
			return nil, rpcDecodeHexError(*c.HighHash)
		}
	}
	headers, err := s.dag.GetTopHeaders(highHash, getTopHeadersMaxHeaders)
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
