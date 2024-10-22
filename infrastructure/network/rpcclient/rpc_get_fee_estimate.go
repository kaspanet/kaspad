package rpcclient

import "github.com/kaspanet/kaspad/app/appmessage"

// GetFeeEstimate sends an RPC request respective to the function's name and returns the RPC server's response
func (c *RPCClient) GetFeeEstimate() (*appmessage.GetFeeEstimateResponseMessage, error) {
	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewGetFeeEstimateRequestMessage())
	if err != nil {
		return nil, err
	}
	response, err := c.route(appmessage.CmdGetFeeEstimateResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return nil, err
	}
	resp := response.(*appmessage.GetFeeEstimateResponseMessage)
	if resp.Error != nil {
		return nil, c.convertRPCError(resp.Error)
	}
	return resp, nil
}
