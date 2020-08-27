package rpchandlers

import (
	"fmt"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/app/rpc/rpcerrors"
	"github.com/kaspanet/kaspad/domain/dagconfig"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetBlockTemplate handles the respectively named RPC command
func HandleGetBlockTemplate(context *rpccontext.Context, request appmessage.Message) (appmessage.Message, error) {
	getBlockTemplateRequest := request.(*appmessage.GetBlockTemplateRequestMessage)

	// Return an error if there are no peers connected since there is no
	// way to relay a found block or receive transactions to work on.
	// However, allow this state when running in the simulation test mode.
	if context.DAG.Params != &dagconfig.SimnetParams && context.ConnectionManager.ConnectionCount() == 0 {
		return nil, &rpcerrors.RPCError{
			Message: "Kaspad is not connected",
		}
	}

	payAddress, err := util.DecodeAddress(getBlockTemplateRequest.PayAddress, context.DAG.Params.Prefix)
	if err != nil {
		return nil, &rpcerrors.RPCError{
			Message: fmt.Sprintf("Could not decode address: %s", err),
		}
	}

	// Protect concurrent access when updating block templates.
	context.BlockTemplateGenerator.Lock()
	defer context.BlockTemplateGenerator.Unlock()

	// Get and return a block template. A new block template will be
	// generated when the current best block has changed or the transactions
	// in the memory pool have been updated and it has been at least five
	// seconds since the last template was generated. Otherwise, the
	// timestamp for the existing block template is updated (and possibly
	// the difficulty on testnet per the consesus rules).
	if err := context.BlockTemplateGenerator.Update(payAddress); err != nil {
		return nil, err
	}
	return context.BlockTemplateGenerator.Response(), nil
}
