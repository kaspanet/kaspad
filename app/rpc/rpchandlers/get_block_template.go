package rpchandlers

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
	"github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/kaspanet/kaspad/util"
)

// HandleGetBlockTemplate handles the respectively named RPC command
func HandleGetBlockTemplate(context *rpccontext.Context, _ *router.Router, request appmessage.Message) (appmessage.Message, error) {
	getBlockTemplateRequest := request.(*appmessage.GetBlockTemplateRequestMessage)

	payAddress, err := util.DecodeAddress(getBlockTemplateRequest.PayAddress, context.DAG.Params.Prefix)
	if err != nil {
		errorMessage := &appmessage.GetBlockTemplateResponseMessage{}
		errorMessage.Error = appmessage.RPCErrorf("Could not decode address: %s", err)
		return errorMessage, nil
	}

	// When a long poll ID was provided, this is a long poll request by the
	// client to be notified when block template referenced by the ID should
	// be replaced with a new one.
	if getBlockTemplateRequest.LongPollID != "" {
		return handleGetBlockTemplateLongPoll(context, getBlockTemplateRequest.LongPollID, payAddress)
	}

	// Protect concurrent access when updating block templates.
	context.BlockTemplateState.Lock()
	defer context.BlockTemplateState.Unlock()

	// Get and return a block template. A new block template will be
	// generated when the current best block has changed or the transactions
	// in the memory pool have been updated and it has been at least five
	// seconds since the last template was generated. Otherwise, the
	// timestamp for the existing block template is updated (and possibly
	// the difficulty on testnet per the consesus rules).
	err = context.BlockTemplateState.Update(payAddress)
	if err != nil {
		return nil, err
	}
	return context.BlockTemplateState.Response()
}

// handleGetBlockTemplateLongPoll is a helper for handleGetBlockTemplateRequest
// which deals with handling long polling for block templates. When a caller
// sends a request with a long poll ID that was previously returned, a response
// is not sent until the caller should stop working on the previous block
// template in favor of the new one. In particular, this is the case when the
// old block template is no longer valid due to a solution already being found
// and added to the block DAG, or new transactions have shown up and some time
// has passed without finding a solution.
func handleGetBlockTemplateLongPoll(context *rpccontext.Context, longPollID string,
	payAddress util.Address) (*appmessage.GetBlockTemplateResponseMessage, error) {
	state := context.BlockTemplateState

	result, longPollChan, err := state.BlockTemplateOrLongPollChan(longPollID, payAddress)
	if err != nil {
		return nil, err
	}

	if result != nil {
		return result, nil
	}

	// Wait until signal received to send the reply.
	<-longPollChan

	// Get the lastest block template
	state.Lock()
	defer state.Unlock()

	if err := state.Update(payAddress); err != nil {
		return nil, err
	}

	// Include whether or not it is valid to submit work against the old
	// block template depending on whether or not a solution has already
	// been found and added to the block DAG.
	result, err = state.Response()
	if err != nil {
		return nil, err
	}

	return result, nil
}
