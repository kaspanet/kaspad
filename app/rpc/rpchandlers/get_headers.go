package rpchandlers

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util/daghash"
)

// HandleGetHeaders handles the respectively named RPC command
func HandleGetHeaders(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getHeadersRequest := request.(*appmessage.GetHeadersRequestMessage)
	dag := context.DAG

	var startHash *daghash.Hash
	if getHeadersRequest.StartHash != "" {
		var err error
		startHash, err = daghash.NewHashFromStr(getHeadersRequest.StartHash)
		if err != nil {
			errorMessage := &appmessage.GetHeadersResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Start hash could not be parsed: %s", err)
			return errorMessage, nil
		}
	}

	const getHeadersDefaultLimit uint64 = 2000
	limit := getHeadersDefaultLimit
	if getHeadersRequest.Limit != 0 {
		limit = getHeadersRequest.Limit
	}

	headers, err := dag.GetHeaders(startHash, limit, getHeadersRequest.IsAscending)
	if err != nil {
		errorMessage := &appmessage.GetHeadersResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Error getting the headers: %s", err)
		return errorMessage, nil
	}

	headersHex := make([]string, len(headers))
	var buf bytes.Buffer
	for i, header := range headers {
		err := header.Serialize(&buf)
		if err != nil {
			errorMessage := &appmessage.GetHeadersResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Failed to serialize block header: %s", err)
			return errorMessage, nil
		}
		headersHex[i] = hex.EncodeToString(buf.Bytes())
		buf.Reset()
	}
	return appmessage.NewGetHeadersResponseMessage(headersHex), nil
}
