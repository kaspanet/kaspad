package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/domain/consensus/ruleerrors"
	"github.com/kaspanet/kaspad/domain/consensus/utils/consensushashing"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
)

// HandleSubmitBlock handles the respectively named RPC command
func HandleSubmitBlock(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	submitBlockRequest := request.(*appmessage.SubmitBlockRequestMessage)

	msgBlock := submitBlockRequest.Block
	domainBlock := appmessage.MsgBlockToDomainBlock(msgBlock)

	err := context.ProtocolManager.AddBlock(domainBlock)
	if err != nil {
		if !errors.As(err, &ruleerrors.RuleError{}) {
			return nil, err
		}
		errorMessage := &appmessage.SubmitBlockResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Block rejected. Reason: %s", err)
		return errorMessage, nil
	}

	log.Infof("Accepted block %s via submitBlock", consensushashing.BlockHash(domainBlock))

	response := appmessage.NewSubmitBlockResponseMessage()
	return response, nil
}
