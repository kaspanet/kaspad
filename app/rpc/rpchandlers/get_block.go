package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleGetBlock handles the respectively named RPC command
func HandleGetBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlockRequest := request.(*appmessage.GetBlockRequestMessage)

	// Load the raw block bytes from the database.
	hash, err := externalapi.NewDomainHashFromString(getBlockRequest.Hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Hash could not be parsed: %s", err)
		return errorMessage, nil
	}

	header, err := context.Domain.Consensus().GetBlockHeader(hash)
	if err != nil {
		errorMessage := &appmessage.GetBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block %s not found", hash)
		return errorMessage, nil
	}

	response := appmessage.NewGetBlockResponseMessage()

	blockVerboseData, err := context.BuildBlockVerboseData(header, nil, getBlockRequest.IncludeTransactionVerboseData)
	if err != nil {
		if errors.Is(err, rpccontext.ErrBuildBlockVerboseDataInvalidBlock) {
			errorMessage := &appmessage.GetBlockResponseMessage{}
			errorMessage.Error = appmessage.RPCErrorf("Block %s is invalid", hash)
			return errorMessage, nil
		}
		return nil, err
	}

	response.BlockVerboseData = blockVerboseData

	return response, nil
}
