package rpc

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/network/rpc/model"
	"github.com/kaspanet/kaspad/util/daghash"
)

const getHeadersMaxHeaders = 2000

// handleGetHeaders implements the getHeaders command.
func handleGetHeaders(s *Server, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*model.GetHeadersCmd)

	lowHash := &daghash.ZeroHash
	if c.LowHash != "" {
		err := daghash.Decode(lowHash, c.LowHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.HighHash)
		}
	}
	highHash := &daghash.ZeroHash
	if c.HighHash != "" {
		err := daghash.Decode(highHash, c.HighHash)
		if err != nil {
			return nil, rpcDecodeHexError(c.HighHash)
		}
	}
	headers, err := s.dag.AntiPastHeadersBetween(lowHash, highHash, getHeadersMaxHeaders)
	if err != nil {
		return nil, &model.RPCError{
			Code:    model.ErrRPCMisc,
			Message: err.Error(),
		}
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
