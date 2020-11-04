package rpchandlers

import (
	"bytes"
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/utils/hashserialization"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
)

// HandleSubmitBlock handles the respectively named RPC command
func HandleSubmitBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	submitBlockRequest := request.(*appmessage.SubmitBlockRequestMessage)

	// Deserialize the submitted block.
	serializedBlock, err := hex.DecodeString(submitBlockRequest.BlockHex)
	if err != nil {
		errorMessage := &appmessage.SubmitBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block hex could not be parsed: %s", err)
		return errorMessage, nil
	}
	msgBlock := &appmessage.MsgBlock{}
	err = msgBlock.KaspaDecode(bytes.NewReader(serializedBlock), 0)
	if err != nil {
		errorMessage := &appmessage.SubmitBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block decode failed: %s", err)
		return errorMessage, nil
	}
	domainBlock := appmessage.MsgBlockToDomainBlock(msgBlock)

	err = context.ProtocolManager.AddBlock(domainBlock)
	if err != nil {
		errorMessage := &appmessage.SubmitBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block rejected. Reason: %s", err)
		return errorMessage, nil
	}

	log.Infof("Accepted domainBlock %s via submitBlock", hashserialization.BlockHash(domainBlock))

	response := appmessage.NewSubmitBlockResponseMessage()
	return response, nil
}
