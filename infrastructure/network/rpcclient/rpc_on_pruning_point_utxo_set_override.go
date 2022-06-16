package rpcclient

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	routerpkg "github.com/kaspanet/kaspad/infrastructure/network/netadapter/router"
	"github.com/pkg/errors"
	"github.com/kaspanet/kaspad/app/rpc/rpccontext"
)

// RegisterPruningPointUTXOSetNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it starts listening for the appropriate notification using the given handler function
func (c *RPCClient) RegisterPruningPointUTXOSetNotifications(onPruningPointUTXOSetNotifications func()) error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewNotifyPruningPointUTXOSetOverrideRequestMessage(rpccontext.DefaultNotificationId))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdNotifyPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	notifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.NotifyPruningPointUTXOSetOverrideResponseMessage)
	if notifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(notifyPruningPointUTXOSetOverrideResponse.Error)
	}
	spawn("RegisterPruningPointUTXOSetNotifications", func() {
		for {
			notification, err := c.route(appmessage.CmdPruningPointUTXOSetOverrideNotificationMessage).Dequeue()
			if err != nil {
				if errors.Is(err, routerpkg.ErrRouteClosed) {
					break
				}
				panic(err)
			}
			_ = notification.(*appmessage.PruningPointUTXOSetOverrideNotificationMessage) // Sanity check the type
			onPruningPointUTXOSetNotifications()
		}
	})
	return nil
}

// UnregisterPruningPointUTXOSetNotifications sends an RPC request respective to the function's name and returns the RPC server's response.
// Additionally, it stops listening for the appropriate notification using the given handler function
func (c *RPCClient) UnregisterPruningPointUTXOSetNotifications() error {

	err := c.rpcRouter.outgoingRoute().Enqueue(appmessage.NewStopNotifyingPruningPointUTXOSetOverrideRequestMessage(rpccontext.DefaultNotificationId))
	if err != nil {
		return err
	}
	response, err := c.route(appmessage.CmdStopNotifyingPruningPointUTXOSetOverrideResponseMessage).DequeueWithTimeout(c.timeout)
	if err != nil {
		return err
	}
	stopNotifyPruningPointUTXOSetOverrideResponse := response.(*appmessage.StopNotifyingPruningPointUTXOSetOverrideResponseMessage)
	if stopNotifyPruningPointUTXOSetOverrideResponse.Error != nil {
		return c.convertRPCError(stopNotifyPruningPointUTXOSetOverrideResponse.Error)
	}
	return nil
}
